package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"mbg/internal/model"
	"mbg/internal/timeutil"
	"mbg/internal/whatsapp"
)

const (
	waMaksPercobaan = 8
	waKedaluwarsa   = 24 * time.Hour
)

type eksekutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
}

type waCfg struct {
	aktif                  bool
	baseURL, keyID, secret string
	tplAmbil, tplKembali   string
}

func bacaWACfg(q interface {
	QueryRow(string, ...any) *sql.Row
}) (waCfg, error) {
	var c waCfg
	var a int
	err := q.QueryRow(`SELECT aktif,base_url,key_id,secret,template_ambil,template_kembali FROM wa_config WHERE id=1`).
		Scan(&a, &c.baseURL, &c.keyID, &c.secret, &c.tplAmbil, &c.tplKembali)
	c.aktif = a == 1
	return c, err
}

func nonKosong(s string) string {
	if s = strings.TrimSpace(s); s == "" {
		return "-"
	}
	return s
}

func waf(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// ---------- konfigurasi ----------

func (s *Service) GetWAConfig() (model.WAConfig, error) {
	c, err := bacaWACfg(s.db)
	if err != nil {
		return model.WAConfig{}, err
	}
	return model.WAConfig{Aktif: c.aktif, BaseURL: c.baseURL, KeyID: c.keyID, SecretAda: c.secret != "",
		TemplateAmbil: c.tplAmbil, TemplateKembali: c.tplKembali}, nil
}

// SimpanWAConfig: Secret kosong berarti secret lama dipertahankan.
func (s *Service) SimpanWAConfig(c model.WAConfig) error {
	lama, err := bacaWACfg(s.db)
	if err != nil {
		return err
	}
	c.BaseURL = strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	c.KeyID = strings.TrimSpace(c.KeyID)
	c.TemplateAmbil = strings.TrimSpace(c.TemplateAmbil)
	c.TemplateKembali = strings.TrimSpace(c.TemplateKembali)
	secret := strings.TrimSpace(c.Secret)
	if secret == "" {
		secret = lama.secret
	}
	if c.Aktif {
		if _, err := whatsapp.New(whatsapp.Config{BaseURL: c.BaseURL, KeyID: c.KeyID, Secret: secret}); err != nil {
			return err
		}
		if c.TemplateAmbil == "" || c.TemplateKembali == "" {
			return errors.New("Kode template pengambilan dan pengembalian wajib diisi")
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE wa_config SET aktif=?,base_url=?,key_id=?,secret=?,template_ambil=?,template_kembali=? WHERE id=1`,
		boolInt(c.Aktif), c.BaseURL, c.KeyID, secret, c.TemplateAmbil, c.TemplateKembali); err != nil {
		return err
	}
	// secret sengaja tidak dicatat di audit log
	if err := s.audit(tx, s.now(), "UBAH", "wa_config", "1",
		map[string]any{"aktif": lama.aktif, "base_url": lama.baseURL, "key_id": lama.keyID, "template_ambil": lama.tplAmbil, "template_kembali": lama.tplKembali},
		map[string]any{"aktif": c.Aktif, "base_url": c.BaseURL, "key_id": c.KeyID, "template_ambil": c.TemplateAmbil, "template_kembali": c.TemplateKembali, "secret_diubah": strings.TrimSpace(c.Secret) != ""},
		""); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	select {
	case s.waWake <- struct{}{}:
	default:
	}
	return nil
}

// ---------- mengisi antrean ----------

// antreanWA: best-effort, dipanggil SETELAH transaksi pencatatan commit.
// Galat hanya dicatat ke log agar tidak pernah menggagalkan pencatatan ompreng.
func (s *Service) antreanWA(q eksekutor, jenis, ref string, kelasID int64, v map[string]string) {
	cfg, err := bacaWACfg(q)
	if err != nil || !cfg.aktif {
		return
	}
	var wid int64
	var wali, no, kelas, ketua string
	err = q.QueryRow(`SELECT w.id,w.nama,w.no_wa,k.nama,k.nama_ketua FROM kelas k
		JOIN wali_kelas w ON w.id=k.wali_kelas_id WHERE k.id=? AND w.aktif=1`, kelasID).
		Scan(&wid, &wali, &no, &kelas, &ketua)
	if err != nil {
		return // kelas belum punya wali aktif
	}
	var inst string
	_ = q.QueryRow(`SELECT COALESCE(nama_instansi,'') FROM pengaturan WHERE id=1`).Scan(&inst)
	v["nama_wali"], v["kelas"], v["ketua"], v["instansi"] = wali, kelas, nonKosong(ketua), nonKosong(inst)
	b, _ := json.Marshal(v)
	if _, err := q.Exec(`INSERT OR IGNORE INTO wa_antrean (jenis,reference_id,wali_id,nama_wali,no_wa,variabel,dibuat) VALUES (?,?,?,?,?,?,?)`,
		jenis, ref, wid, wali, no, string(b), waf(s.now())); err != nil {
		log.Printf("antrean WA: %v", err)
		return
	}
	select {
	case s.waWake <- struct{}{}:
	default:
	}
}

// ---------- worker ----------

// MulaiAntreanWA: dipanggil sekali dari main saat aplikasi mulai.
func (s *Service) MulaiAntreanWA() {
	if s.waCancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.waCancel = cancel
	go func() {
		t := time.NewTicker(20 * time.Second)
		defer t.Stop()
		for {
			s.prosesAntreanWA(ctx)
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			case <-s.waWake:
			}
		}
	}()
}

func (s *Service) BerhentiAntreanWA() {
	if s.waCancel != nil {
		s.waCancel()
	}
}

type itemWA struct {
	id        int64
	jenis, no string
	vars, ref string
	coba      int
}

func (s *Service) prosesAntreanWA(ctx context.Context) {
	cfg, err := bacaWACfg(s.db)
	if err != nil || !cfg.aktif {
		return
	}
	now := s.now()
	_, _ = s.db.Exec(`UPDATE wa_antrean SET status='GAGAL',galat='Kedaluwarsa (lebih dari 24 jam di antrean)' WHERE status='ANTRE' AND dibuat<?`,
		waf(now.Add(-waKedaluwarsa)))
	_, _ = s.db.Exec(`DELETE FROM wa_antrean WHERE status IN ('TERKIRIM','GAGAL') AND dibuat<?`,
		waf(now.Add(-7*24*time.Hour)))
	cl, err := whatsapp.New(whatsapp.Config{BaseURL: cfg.baseURL, KeyID: cfg.keyID, Secret: cfg.secret})
	if err != nil {
		log.Printf("WA: %v", err)
		return
	}
	rows, err := s.db.Query(`SELECT id,jenis,no_wa,variabel,reference_id,percobaan FROM wa_antrean
		WHERE status='ANTRE' AND (coba_lagi IS NULL OR coba_lagi<=?) ORDER BY id LIMIT 20`, waf(now))
	if err != nil {
		return
	}
	var items []itemWA
	for rows.Next() { // dibaca habis dulu: koneksi database hanya satu
		var it itemWA
		if err := rows.Scan(&it.id, &it.jenis, &it.no, &it.vars, &it.ref, &it.coba); err != nil {
			rows.Close()
			return
		}
		items = append(items, it)
	}
	rows.Close()

	for _, it := range items {
		if ctx.Err() != nil {
			return
		}
		var v map[string]string
		_ = json.Unmarshal([]byte(it.vars), &v)
		tpl := cfg.tplAmbil
		if it.jenis == "KEMBALI" {
			tpl = cfg.tplKembali
		}
		h, err := cl.Kirim(ctx, whatsapp.Pesan{TemplateCode: tpl, Recipient: it.no, Variables: v, ReferenceID: it.ref})
		if ctx.Err() != nil {
			return // aplikasi ditutup: biarkan tetap ANTRE
		}
		t := s.now()
		switch {
		case err == nil:
			_, _ = s.db.Exec(`UPDATE wa_antrean SET status='TERKIRIM',job_id=?,status_layanan=?,galat=NULL,terkirim=?,percobaan=percobaan+1 WHERE id=?`,
				h.JobID, h.Status, waf(t), it.id)
		case errors.Is(err, whatsapp.ErrAutentikasi), errors.Is(err, whatsapp.ErrDibatasi):
			// konfigurasi/jam salah atau API key dijeda: hentikan seluruh batch, coba lagi 10 menit lagi
			_, _ = s.db.Exec(`UPDATE wa_antrean SET galat=?,coba_lagi=? WHERE status='ANTRE'`,
				err.Error(), waf(t.Add(10*time.Minute)))
			return
		case whatsapp.IsSementara(err):
			coba := it.coba + 1
			if coba >= waMaksPercobaan {
				_, _ = s.db.Exec(`UPDATE wa_antrean SET status='GAGAL',galat=?,percobaan=? WHERE id=?`, err.Error(), coba, it.id)
			} else {
				d := time.Duration(1<<coba) * time.Minute
				if d > time.Hour {
					d = time.Hour
				}
				_, _ = s.db.Exec(`UPDATE wa_antrean SET galat=?,percobaan=?,coba_lagi=? WHERE id=?`, err.Error(), coba, waf(t.Add(d)), it.id)
			}
			return // jaringan/server bermasalah: jangan menghabiskan timeout per pesan
		default:
			_, _ = s.db.Exec(`UPDATE wa_antrean SET status='GAGAL',galat=?,percobaan=percobaan+1 WHERE id=?`, err.Error(), it.id)
			log.Printf("WA %s ditolak: %v", it.ref, err)
		}
	}
}

func teksStatusKembali(sisa int) string {
	if sisa > 0 {
		return fmt.Sprintf("Kurang %d Ompreng", sisa)
	}
	return "Lengkap"
}

// ---------- riwayat, uji, kirim ulang ----------

func (s *Service) RingkasanWA() (model.WARingkasan, error) {
	var r model.WARingkasan
	rows, err := s.db.Query(`SELECT status,COUNT(*) FROM wa_antrean GROUP BY status`)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			return r, err
		}
		switch st {
		case "ANTRE":
			r.Antre = n
		case "TERKIRIM":
			r.Terkirim = n
		case "GAGAL":
			r.Gagal = n
		}
	}
	return r, rows.Err()
}

// DaftarAntreanWA: status kosong = semua. Terbaru di atas.
func (s *Service) DaftarAntreanWA(status string, limit int) ([]model.WAAntrean, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.Query(`SELECT id,jenis,COALESCE(nama_wali,''),no_wa,status,COALESCE(status_layanan,''),
		percobaan,COALESCE(galat,''),dibuat,COALESCE(coba_lagi,''),COALESCE(terkirim,'')
		FROM wa_antrean WHERE (?='' OR status=?) ORDER BY id DESC LIMIT ?`, status, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.WAAntrean{}
	for rows.Next() {
		var a model.WAAntrean
		if err := rows.Scan(&a.ID, &a.Jenis, &a.NamaWali, &a.NoWA, &a.Status, &a.StatusLayanan,
			&a.Percobaan, &a.Galat, &a.Dibuat, &a.CobaLagi, &a.Terkirim); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// UlangiWA: mengembalikan pesan GAGAL ke antrean (umur dihitung ulang agar tidak langsung kedaluwarsa).
func (s *Service) UlangiWA(id int64) error {
	res, err := s.db.Exec(`UPDATE wa_antrean SET status='ANTRE',galat=NULL,percobaan=0,coba_lagi=NULL,dibuat=?
		WHERE id=? AND status='GAGAL'`, waf(s.now()), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("Hanya pesan berstatus GAGAL yang bisa diulang")
	}
	select {
	case s.waWake <- struct{}{}:
	default:
	}
	return nil
}

// UjiKirimWA: kirim langsung (sinkron) memakai template pengambilan dan data contoh.
// Tidak masuk antrean, jadi hasilnya langsung terlihat oleh operator.
func (s *Service) UjiKirimWA(noWA string) (model.WAUji, error) {
	var no strings.Builder
	for _, r := range noWA {
		if r >= '0' && r <= '9' {
			no.WriteRune(r)
		}
	}
	if no.Len() < 8 {
		return model.WAUji{}, errors.New("Nomor WhatsApp tidak valid")
	}
	cfg, err := bacaWACfg(s.db)
	if err != nil {
		return model.WAUji{}, err
	}
	cl, err := whatsapp.New(whatsapp.Config{BaseURL: cfg.baseURL, KeyID: cfg.keyID, Secret: cfg.secret})
	if err != nil {
		return model.WAUji{}, err
	}
	if cfg.tplAmbil == "" {
		return model.WAUji{}, errors.New("Kode template pengambilan belum diisi")
	}
	now := s.now()
	inst := ""
	_ = s.db.QueryRow(`SELECT COALESCE(nama_instansi,'') FROM pengaturan WHERE id=1`).Scan(&inst)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	h, err := cl.Kirim(ctx, whatsapp.Pesan{
		TemplateCode: cfg.tplAmbil,
		Recipient:    no.String(),
		ReferenceID:  fmt.Sprintf("mbg-uji-%d", now.UnixNano()),
		Variables: map[string]string{
			"nama_wali": "Wali Kelas (uji)", "kelas": "VII A", "ketua": "Contoh",
			"pengambil": "Contoh", "jumlah_ambil": "30",
			"tanggal": timeutil.HariTanggalIndo(now), "jam": timeutil.Jam(now),
			"instansi": nonKosong(inst),
		},
	})
	if err != nil {
		return model.WAUji{}, err
	}
	return model.WAUji{JobID: h.JobID, Status: h.Status}, nil
}
