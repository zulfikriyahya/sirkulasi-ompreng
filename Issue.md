## Struktur Project

```bash
sirkulasi-ompreng on main ≡  ?3 ~6
➜ tree
.
├── build
│   ├── appicon.png
│   ├── bin
│   │   ├── mbg-amd64-installer.exe
│   │   └── mbg.exe
│   ├── linux
│   │   ├── 99-mbg-reader.rules
│   │   ├── mbg.desktop
│   │   ├── mbg.png
│   │   └── postinst.sh
│   └── windows
│       ├── icon.ico
│       ├── info.json
│       ├── installer
│       │   ├── project.nsi
│       │   ├── tmp
│       │   │   └── MicrosoftEdgeWebview2Setup.exe
│       │   └── wails_tools.nsh
│       └── wails.exe.manifest
├── dist
│   ├── checksums.txt
│   ├── mbg-0.2.1-1.x86_64.rpm
│   ├── mbg_0.2.1_amd64.deb
│   ├── mbg-amd64-installer.exe
│   └── mbg-linux-amd64
├── firmware
│   └── esp32c3_rc522
│       └── esp32c3_rc522.ino
├── frontend
│   ├── css
│   │   └── style.css
│   ├── images
│   │   └── icon.ico
│   ├── index.html
│   ├── js
│   │   ├── api.js
│   │   ├── dashboard.js
│   │   ├── ekspor.js
│   │   ├── main.js
│   │   ├── master.js
│   │   ├── pembaruan.js
│   │   ├── riwayat.js
│   │   ├── sinkron.js
│   │   └── tap.js
│   └── wailsjs
│       ├── go
│       │   ├── models.ts
│       │   └── service
│       │       ├── Service.d.ts
│       │       └── Service.js
│       └── runtime
│           ├── package.json
│           ├── runtime.d.ts
│           └── runtime.js
├── go.mod
├── go.sum
├── icon.png
├── internal
│   ├── db
│   │   └── db.go
│   ├── export
│   │   ├── excel.go
│   │   ├── export.go
│   │   └── pdf.go
│   ├── model
│   │   └── model.go
│   ├── pembaruan
│   │   ├── pembaruan.go
│   │   └── pembaruan_test.go
│   ├── reader
│   │   ├── serial.go
│   │   ├── serial_test.go
│   │   ├── uid.go
│   │   └── uid_test.go
│   ├── service
│   │   ├── ekspor.go
│   │   ├── ekspor_test.go
│   │   ├── hari.go
│   │   ├── master.go
│   │   ├── pembaruan.go
│   │   ├── pembaruan_test.go
│   │   ├── reader.go
│   │   ├── riwayat.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── sinkron.go
│   │   ├── sinkron_test.go
│   │   ├── susulan_test.go
│   │   ├── tahap5_test.go
│   │   ├── transaksi.go
│   │   ├── wa.go
│   │   ├── wali.go
│   │   ├── wali_test.go
│   │   └── wa_test.go
│   ├── sinkron
│   │   └── sinkron.go
│   ├── timeutil
│   │   └── timeutil.go
│   ├── versi
│   │   └── versi.go
│   └── whatsapp
│       ├── client.go
│       └── client_test.go
├── Issue.md
├── LICENSE
├── main.go
├── nfpm.yaml
├── README.md
└── wails.json

29 directories, 81 files
sirkulasi-ompreng on main ≡  ?3 ~6
➜
```

## Issues
- tambahkan fitur notifikasi whatsapp yang dikirimkan ke wali kelas
- layanan whatsapp (https://whatsapp.mtsn1pandeglang.sch.id/api/v1/messages)

```txt
Panduan Integrasi API untuk Aplikasi Eksternal

Aplikasi eksternal milik instansi dapat mengirim notifikasi WhatsApp melalui endpoint POST /api/v1/messages menggunakan autentikasi berbasis HMAC-SHA256 - bukan Bearer token. Setiap request wajib menyertakan tiga header: X-API-Key (key_id API key Anda), X-Signature (HMAC-SHA256 dari body request mentah, memakai secret API key sebagai kunci), dan X-Timestamp (Unix timestamp saat request dibuat, dengan toleransi selisih waktu ±5 menit terhadap waktu server untuk mencegah serangan replay). Endpoint ini mendukung pengiriman pesan teks, dokumen, gambar, video, link, dan kontak vCard.

    Pastikan Anda sudah memiliki API key aktif (key_id + secret) - lihat Panduan API Key jika belum.
    Buat atau pilih template pesan yang sudah dikaitkan ke API key ini dari halaman Template, lengkap dengan daftar variabel wajib yang harus disertakan saat pengiriman.
    Normalisasi nomor tujuan (recipient) ke format internasional Indonesia (mis. 6281234567890, boleh dengan atau tanpa awalan 0/62/8 - sistem akan menormalisasi otomatis) sebelum dikirim.
    Susun body JSON berisi template_code, recipient, variables (wajib memuat seluruh variabel_wajib milik template), dan opsional media (untuk dokumen/gambar/video/link/kontak) serta reference_id (untuk idempotency).
    Hitung X-Signature sebagai HMAC-SHA256 dari body request mentah (byte persis yang dikirim) menggunakan secret API key Anda sebagai kunci, lalu encode hasilnya sebagai hex string.
    Sertakan header X-API-Key, X-Signature, dan X-Timestamp pada setiap request - request tanpa HMAC valid atau timestamp kedaluwarsa akan ditolak dengan status 401 Unauthorized.
    Gunakan reference_id yang unik per transaksi/aksi bisnis di sisi Anda untuk memanfaatkan idempotency (window 24 jam) - request dengan reference_id sama dan payload identik tidak akan membuat pengiriman ganda.
    Tangani response 429 (Too Many Requests) sebagai tanda guard rail API key sedang auto-pause akibat lonjakan volume - hubungi admin instansi untuk mengaktifkan kembali setelah penyebab diperiksa.
    Pantau status pengiriman secara real-time via GET /api/v1/messages/{job_id} (status: queued, processing, sent, delivered, read, atau failed), atau lihat riwayatnya di halaman Log Pesan/Antrian Job pada dashboard.

Panduan Pembuatan Template

Template pesan memungkinkan Anda menyusun format pesan WhatsApp sekali saja dan memakainya berulang dengan variabel berbeda-beda di setiap pengiriman - cocok untuk notifikasi berulang seperti presensi, OTP, tagihan, atau pengumuman. Setiap template memiliki kode unik per instansi, kategori bebas (mis. 'Presensi', 'OTP', 'Tagihan'), prioritas pengiriman (High/Medium/Low), 2 hingga 10 varian teks yang dipilih acak setiap kali dikirim (mengurangi pola pengiriman seragam yang berisiko terdeteksi sebagai spam), serta daftar variabel wajib yang harus disertakan aplikasi pengirim.

    Buka halaman Template pada dashboard instansi Anda (menu sidebar 'Template').
    Klik 'Buat Template Baru', isi kode unik (huruf kecil, tanpa spasi, mis. presensi_masuk), kategori bebas sesuai kebutuhan Anda, dan pilih prioritas: High untuk pesan mendesak seperti OTP/reset password, Medium untuk notifikasi transaksional rutin seperti presensi/tagihan, atau Low untuk informasi umum/pengumuman yang tidak mendesak.
    Tulis 2 hingga 10 varian teks dengan format placeholder {{nama_variabel}} - semakin banyak varian yang wajar dan berbeda gaya bahasanya, semakin baik untuk menyamarkan pola pengiriman otomatis.
    Perhatikan format yang dibatasi validator: hindari huruf kapital berlebihan dalam satu kalimat, tidak lebih dari satu tanda seru berurutan (!!), dan maksimal satu tautan (link) per varian - template yang melanggar akan ditolak sistem karena berisiko ditandai spam oleh WhatsApp.
    Isi daftar variabel wajib (dipisah koma, mis. nama_siswa,jam) - setiap variabel yang disebut di dalam teks tapi tidak terdaftar sebagai wajib tetap boleh dikirim kosong tanpa validasi ketat, jadi pastikan variabel penting selalu didaftarkan sebagai wajib.
    Kaitkan template ke satu API key tertentu dari daftar template jika ingin membatasi penggunaannya hanya untuk aplikasi tertentu (satu template hanya bisa terhubung ke satu API key dalam satu waktu; putuskan kaitan dari daftar yang sama jika ingin mengaitkan ulang ke API key lain).
    Simpan template, lalu gunakan kodenya sebagai nilai field template_code saat mengirim pesan melalui endpoint API eksternal.
    Uji coba pengiriman ke nomor WhatsApp Anda sendiri terlebih dahulu untuk memverifikasi seluruh variabel tampil benar sebelum dipakai secara luas ke penerima sesungguhnya.
    Anda dapat mengubah kategori, prioritas, varian, dan variabel wajib kapan saja tanpa mengubah kode template (sehingga integrasi API yang sudah berjalan tidak perlu diubah).

```
```bash
curl -X POST https://whatsapp.mtsn1pandeglang.sch.id/api/v1/messages \
  -H "X-API-Key: key_xxx" \
  -H "X-Signature: <hmac_sha256_dari_body>" \
  -H "X-Timestamp: 1751875200" \
  -H "Content-Type: application/json" \
  -d '{
    "template_code": "presensi_masuk",
    "recipient": "6281234567890",
    "variables": { "nama": "Yahya Zulfikri", "jam": "07:15" },
    "media": null,
    "reference_id": "APP-INTERNAL-12345"
  }'

# Contoh cek status pengiriman:
curl -X GET https://whatsapp.mtsn1pandeglang.sch.id/api/v1/messages/job_xxx \
  -H "X-API-Key: key_xxx" \
  -H "X-Signature: <hmac_sha256_dari_body_kosong>" \
  -H "X-Timestamp: 1751875200"

# Contoh isi salah satu varian template (kategori Presensi, prioritas Medium):
"Halo {{nama}}, kehadiran Anda pada {{tanggal}} pukul {{jam}} telah berhasil dicatat. Terima kasih."

Varian kedua untuk template yang sama (dipilih acak bergantian):
"Presensi tercatat: {{nama}} hadir pada {{tanggal}} pukul {{jam}}."

Dikirim aplikasi eksternal via field "variables":
{ "nama": "Yahya Zulfikri", "tanggal": "7 Juli 2026", "jam": "08:00" }

Contoh lain untuk kategori OTP (prioritas High, variabel wajib: kode_otp,durasi_menit):
"Kode OTP Anda: {{kode_otp}}. Berlaku {{durasi_menit}} menit, jangan bagikan ke siapa pun."
```

## Percakapan sebelumnya:
Pesan masuk ke antrean (outbox) di database, lalu dikirim oleh worker latar belakang. Operator tidak menunggu jaringan, dan pesan tidak hilang kalau internet putus.

- Antrean diisi **setelah** `SubmitAmbil` dan `SubmitKembali` berhasil commit. Kalau pengisian antrean gagal, pencatatan ompreng tetap aman.
- Pesan hanya dibuat jika fitur WhatsApp aktif dan kelas punya wali aktif.
- `reference_id` unik per kejadian (`mbg-ambil-<id transaksi>`, `mbg-kembali-<id pengembalian>`), jadi tidak ada pesan ganda.
- Pesan yang lebih dari 24 jam di antrean ditandai gagal, supaya wali tidak kebanjiran pesan lama saat internet kembali.

## 1. `internal/model/model.go`: tambahkan

```go
type WAConfig struct {
	Aktif           bool   `json:"aktif"`
	BaseURL         string `json:"base_url"`
	KeyID           string `json:"key_id"`
	Secret          string `json:"secret"`      // hanya masuk; tidak pernah dikirim balik ke UI
	SecretAda       bool   `json:"secret_ada"`  // true bila secret sudah tersimpan
	TemplateAmbil   string `json:"template_ambil"`
	TemplateKembali string `json:"template_kembali"`
}
```

## 2. `internal/db/db.go`

Tambahkan di akhir `schema` (sebelum backtick penutup):

```sql
CREATE TABLE IF NOT EXISTS wa_config (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  aktif INTEGER NOT NULL DEFAULT 0,
  base_url TEXT NOT NULL DEFAULT 'https://whatsapp.mtsn1pandeglang.sch.id',
  key_id TEXT NOT NULL DEFAULT '',
  secret TEXT NOT NULL DEFAULT '',
  template_ambil TEXT NOT NULL DEFAULT 'ompreng_ambil',
  template_kembali TEXT NOT NULL DEFAULT 'ompreng_kembali'
);
CREATE TABLE IF NOT EXISTS wa_antrean (
  id INTEGER PRIMARY KEY,
  jenis TEXT NOT NULL,
  reference_id TEXT NOT NULL UNIQUE,
  wali_id INTEGER,
  nama_wali TEXT,
  no_wa TEXT NOT NULL,
  variabel TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'ANTRE',
  job_id TEXT,
  status_layanan TEXT,
  percobaan INTEGER NOT NULL DEFAULT 0,
  galat TEXT,
  dibuat TEXT NOT NULL,
  coba_lagi TEXT,
  terkirim TEXT
);
CREATE INDEX IF NOT EXISTS idx_wa_antrean_status ON wa_antrean(status, id);
```

Di `Migrate`, tepat sebelum `_, err := conn.Exec(`INSERT OR IGNORE INTO pengaturan ...`)`:

```go
	if _, err := conn.Exec(`INSERT OR IGNORE INTO wa_config (id) VALUES (1)`); err != nil {
		return err
	}
```

## 3. `internal/service/service.go`

Tambahkan `"context"` ke import. Di struct `Service`, tambahkan dua field:

```go
	waWake   chan struct{}
	waCancel context.CancelFunc
```

Ubah `NewWithClock`:

```go
func NewWithClock(db *sql.DB, dir string, now func() time.Time) *Service {
	return &Service{db: db, dir: dir, now: now, lastSubmit: map[int64]time.Time{}, waWake: make(chan struct{}, 1)}
}
```

## 4. File baru `internal/service/wa.go`

```go
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
	aktif                       bool
	baseURL, keyID, secret      string
	tplAmbil, tplKembali        string
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
		t := s.now()
		switch {
		case err == nil:
			_, _ = s.db.Exec(`UPDATE wa_antrean SET status='TERKIRIM',job_id=?,status_layanan=?,galat=NULL,terkirim=?,percobaan=percobaan+1 WHERE id=?`,
				h.JobID, h.Status, waf(t), it.id)
		case errors.Is(err, whatsapp.ErrAutentikasi), errors.Is(err, whatsapp.ErrDibatasi):
			// konfigurasi/jam salah atau API key dijeda: hentikan seluruh batch, coba lagi 10 menit lagi
			_, _ = s.db.Exec(`UPDATE wa_antrean SET galat=?,coba_lagi=? WHERE id=?`, err.Error(), waf(t.Add(10*time.Minute)), it.id)
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
```

## 5. `internal/service/transaksi.go`

**a. `SubmitAmbil`:** tepat setelah blok `if err := tx.Commit(); err != nil { return err }`, sebelum `s.lastSubmit[kelasID] = now`, sisipkan:

```go
	s.antreanWA(s.db, "AMBIL", "mbg-ambil-"+itoa(id), kelasID, map[string]string{
		"pengambil":    nonKosong(s.siswaNama),
		"jumlah_ambil": strconv.Itoa(jumlah),
		"tanggal":      timeutil.HariTanggalIndo(now),
		"jam":          timeutil.Jam(now),
	})
```

**b. `SubmitKembali`:** ganti insert pengembalian agar `id`-nya tertangkap:

```go
	resP, err := tx.Exec(`INSERT INTO pengembalian (transaksi_id,waktu,jumlah,nisn,nama_siswa) VALUES (?,?,?,?,?)`,
		id, jam, jumlah, s.siswaNISN, s.siswaNama)
	if err != nil {
		return err
	}
	pid, _ := resP.LastInsertId()
```

Lalu setelah `tx.Commit()`, sebelum `if statusBaru == "KURANG" {`, sisipkan:

```go
	sisa := jAmbil - total
	jenisKembali := "Pengembalian"
	if susulan {
		jenisKembali = "Pengembalian susulan"
	}
	s.antreanWA(s.db, "KEMBALI", "mbg-kembali-"+itoa(pid), kelasID, map[string]string{
		"jenis":          jenisKembali,
		"pengembali":     nonKosong(s.siswaNama),
		"jumlah_ambil":   strconv.Itoa(jAmbil),
		"jumlah_kembali": strconv.Itoa(jumlah),
		"total_kembali":  strconv.Itoa(total),
		"sisa":           strconv.Itoa(sisa),
		"status":         teksStatusKembali(sisa),
		"tanggal":        timeutil.HariTanggalIndo(now),
		"jam":            jam,
	})
```

## 6. File baru `internal/service/wa_test.go`

```go
package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"mbg/internal/db"
	"mbg/internal/model"
	"mbg/internal/timeutil"
)

func TestAntreanWA(t *testing.T) {
	var diterima []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		diterima = append(diterima, m)
		w.Write([]byte(`{"data":{"job_id":"job_1","status":"queued"}}`))
	}))
	defer srv.Close()

	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	s := NewWithClock(conn, t.TempDir(), func() time.Time {
		return time.Date(2026, 10, 24, 7, 15, 0, 0, timeutil.WIB)
	})
	if err := s.SimpanWAConfig(model.WAConfig{Aktif: true, BaseURL: srv.URL, KeyID: "k", Secret: "s",
		TemplateAmbil: "ompreng_ambil", TemplateKembali: "ompreng_kembali"}); err != nil {
		t.Fatal(err)
	}
	kA, _ := s.TambahKelas("VII A", "Andi", 1)
	kB, _ := s.TambahKelas("VII B", "Budi", 2) // tanpa wali
	wid, err := s.TambahWaliKelas("Bu Siti", "0812-3456-7890")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetWaliKelas(kA, wid); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitAmbil(kA, 30); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitAmbil(kB, 25); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitKembali(kA, 28); err != nil {
		t.Fatal(err)
	}

	var n int
	_ = conn.QueryRow(`SELECT COUNT(*) FROM wa_antrean`).Scan(&n)
	if n != 2 {
		t.Fatalf("antrean = %d, want 2 (kelas tanpa wali tidak boleh masuk)", n)
	}
	s.prosesAntreanWA(context.Background())
	s.prosesAntreanWA(context.Background()) // kedua kali tidak boleh mengirim ulang
	if len(diterima) != 2 {
		t.Fatalf("terkirim = %d, want 2", len(diterima))
	}
	if diterima[0]["template_code"] != "ompreng_ambil" || diterima[0]["recipient"] != "6281234567890" {
		t.Errorf("pesan pertama salah: %v", diterima[0])
	}
	v := diterima[1]["variables"].(map[string]any)
	if diterima[1]["template_code"] != "ompreng_kembali" || v["total_kembali"] != "28" || v["sisa"] != "2" || v["status"] != "Kurang 2 Ompreng" {
		t.Errorf("pesan kembali salah: %v", diterima[1])
	}
	_ = conn.QueryRow(`SELECT COUNT(*) FROM wa_antrean WHERE status='TERKIRIM'`).Scan(&n)
	if n != 2 {
		t.Errorf("TERKIRIM = %d, want 2", n)
	}
}
```

## 7. `main.go`

Saya belum melihat `main.go`, jadi tempatnya perlu disesuaikan. Panggil di callback `OnStartup` (setelah konteks Wails diset ke service) dan `OnShutdown`:

```go
svc.MulaiAntreanWA()      // di OnStartup
svc.BerhentiAntreanWA()   // di OnShutdown
```

Kalau ragu menaruhnya di mana, kirim `main.go`.

## 8. Template yang perlu Anda buat di dashboard

Kode template bisa diubah di pengaturan (Tahap 5). Nilai awalnya `ompreng_ambil` dan `ompreng_kembali`. Semua variabel di bawah selalu terisi, tidak pernah kosong (kosong diganti `-`), jadi aman didaftarkan sebagai wajib.

**`ompreng_ambil`** (prioritas Medium), variabel wajib:
`nama_wali,kelas,ketua,pengambil,jumlah_ambil,tanggal,jam,instansi`

Contoh varian: `Assalamualaikum {{nama_wali}}, kelas {{kelas}} telah mengambil {{jumlah_ambil}} ompreng pada {{tanggal}} pukul {{jam}}. Diambil oleh {{pengambil}}. Terima kasih.`

**`ompreng_kembali`** (prioritas Medium), variabel wajib:
`nama_wali,kelas,ketua,jenis,pengembali,jumlah_ambil,jumlah_kembali,total_kembali,sisa,status,tanggal,jam,instansi`

Contoh varian: `{{jenis}} kelas {{kelas}} pada {{tanggal}} pukul {{jam}}: {{jumlah_kembali}} ompreng dikembalikan oleh {{pengembali}}. Total kembali {{total_kembali}} dari {{jumlah_ambil}} ompreng. Status: {{status}}.`

Arti variabel yang kurang jelas:
- `jumlah_kembali` adalah jumlah pada pengembalian kali ini. `total_kembali` adalah akumulasi hari itu.
- `status` berisi `Lengkap` atau `Kurang N Ompreng`.
- `jenis` berisi `Pengembalian` atau `Pengembalian susulan`.

Validator layanan menolak huruf kapital berlebihan dan lebih dari satu `!`. Buat minimal 2 varian dengan gaya kalimat berbeda.

## 9. Verifikasi

```bash
gofmt -l internal && go build ./... && go vet ./... && go test ./...
```

Catatan:
- `GetWAConfig` dan `SimpanWAConfig` otomatis ikut binding Wails saat `wails dev`.
- Secret tersimpan di SQLite lokal sebagai teks biasa. Itu wajar untuk aplikasi desktop, tetapi jangan menaruh folder datanya di tempat yang dibagikan. Secret tidak ikut ekspor sinkronisasi dan tidak masuk audit log.
- Koreksi lewat menu Riwayat (edit transaksi) tidak mengirim pesan, karena yang dikirim hanya kejadian submit dari operator.

Kirim hasil perintah di nomor 9 dan `main.go`. Setelah itu kita masuk ke **Tahap 5**: form pengaturan WhatsApp, tombol uji kirim, dan tabel riwayat notifikasi beserta status pengirimannya.

sirkulasi-ompreng on main ≡  ?3 ~6
➜ gofmt -l internal && go build ./... && go vet ./... && go test ./...
?       mbg     [no test files]
?       mbg/internal/db [no test files]
?       mbg/internal/export     [no test files]
?       mbg/internal/model      [no test files]
ok      mbg/internal/pembaruan  (cached)
ok      mbg/internal/reader     (cached)
ok      mbg/internal/service    (cached)
?       mbg/internal/sinkron    [no test files]
?       mbg/internal/timeutil   [no test files]
?       mbg/internal/versi      [no test files]
ok      mbg/internal/whatsapp   (cached)
sirkulasi-ompreng on main ≡  ?3 ~6
➜ cat main.go
package main

import (
        "context"
        "embed"
        "log"
        "os"
        "path/filepath"

        "github.com/wailsapp/wails/v2/pkg/options/linux"

        "github.com/wailsapp/wails/v2"
        "github.com/wailsapp/wails/v2/pkg/options"
        "github.com/wailsapp/wails/v2/pkg/options/assetserver"

        "mbg/internal/db"
        "mbg/internal/reader"
        "mbg/internal/service"

        wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
        dir := db.DataDir()
        if err := os.MkdirAll(dir, 0o755); err != nil {
                log.Fatalf("gagal membuat folder data %s: %v", dir, err)
        }

        conn, err := db.Open(filepath.Join(dir, "mbg.db"))
        if err != nil {
                log.Fatalf("gagal membuka database: %v", err)
        }
        defer conn.Close()

        svc := service.New(conn, dir)

        var appCtx context.Context
        mgr := reader.NewManager(
                func() string { p, _ := svc.GetPengaturan(); return p.PortSerial },
                func(uid string) { wruntime.EventsEmit(appCtx, "card:tap", uid) },
                func(st reader.Status) {
                        svc.SetStatusReader(st)
                        if appCtx != nil {
                                wruntime.EventsEmit(appCtx, "reader:status", st)
                        }
                },
        )

        err = wails.Run(&options.App{
                Title:       "SI'OM - Sirkulasi Ompreng MBG",
                Width:       1200,
                Height:      800,
                AssetServer: &assetserver.Options{Assets: assets},
                Linux:       &linux.Options{Icon: icon, ProgramName: "mbg"},
                OnStartup: func(ctx context.Context) {
                        appCtx = ctx
                        svc.SetContext(ctx)
                        go mgr.Run(ctx)
                        svc.MulaiAntreanWA()
                },
                OnShutdown: func(ctx context.Context) {
                        svc.BerhentiAntreanWA()
                },
                Bind: []interface{}{svc},
        })
        if err != nil {
                log.Fatalf("wails: %v", err)
        }
}
sirkulasi-ompreng on main ≡  ?3 ~6
➜ 
