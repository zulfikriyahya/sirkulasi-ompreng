package service

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"mbg/internal/model"
	"mbg/internal/reader"
	"mbg/internal/sinkron"
	"mbg/internal/timeutil"
)

// pembaca: dipenuhi *sql.DB dan *sql.Tx, supaya perencanaan impor bisa dijalankan
// di luar transaksi (pratinjau) maupun di dalamnya (terapkan).
type pembaca interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type kelasLama struct {
	id     int64
	ketua  string
	urutan int
	aktif  bool
}

type kartuLama struct {
	kelas, nisn, nama string
	aktif             bool
}

type kelasRencana struct {
	baris  int
	nama   string
	ketua  string
	urutan int
	aktif  bool
	aksi   string // BARU | UBAH | SAMA | OTOMATIS | NONAKTIF
	lama   *kelasLama
}

type kartuRencana struct {
	baris                  int
	uid, kelas, nisn, nama string
	aktif                  bool
	aksi                   string // BARU | UBAH | SAMA | HAPUS
	lama                   *kartuLama
}

type rencanaImpor struct {
	kelas          []kelasRencana
	kartu          []kartuRencana
	pengaturan     map[string]string // hanya kunci yang berubah
	pengaturanLama model.Pengaturan
	baris          []model.BarisPratinjau
	r              model.RingkasanImpor
}

func (rc *rencanaImpor) perubahan() int {
	r := rc.r
	return r.KelasBaru + r.KelasUbah + r.KelasNonaktif + r.KelasOtomatis +
		r.KartuBaru + r.KartuUbah + r.KartuHapus + r.PengaturanUbah
}

// hitungHash: sidik jari berkas + opsi + rencana. Dicek ulang di dalam transaksi saat menerapkan,
// sehingga berkas atau database yang berubah sejak pratinjau membatalkan impor.
func (rc *rencanaImpor) hitungHash(sha string, o model.OpsiImpor) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%t|%+v\n", sha, o.Mode, o.BuatKelasOtomatis, rc.r)
	for _, b := range rc.baris {
		fmt.Fprintf(h, "%s|%d|%s|%s|%s\n", b.Entitas, b.Baris, b.Kunci, b.Aksi, b.Pesan)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (rc *rencanaImpor) hasil(path, hash string, o model.OpsiImpor) model.HasilPratinjau {
	baris := append([]model.BarisPratinjau{}, rc.baris...)
	rank := func(a string) int {
		switch a {
		case "GALAT":
			return 0
		case "PERINGATAN":
			return 1
		}
		return 2
	}
	sort.SliceStable(baris, func(i, j int) bool { return rank(baris[i].Aksi) < rank(baris[j].Aksi) })
	out := model.HasilPratinjau{Berkas: filepath.Base(path), Mode: o.Mode, Ringkasan: rc.r, Hash: hash}
	if len(baris) > 500 {
		baris = baris[:500]
		out.Terpotong = true
	}
	out.Baris = baris
	return out
}

// ---------- pembantu ----------

func semuaAngka(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func parseAktif(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "1", "ya", "y", "true", "aktif":
		return true, true
	case "0", "tidak", "t", "n", "false", "nonaktif":
		return false, true
	}
	return false, false
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func normalMode(o *model.OpsiImpor) error {
	switch strings.ToUpper(strings.TrimSpace(o.Mode)) {
	case "", "GABUNG":
		o.Mode = "GABUNG"
	case "GANTI":
		o.Mode = "GANTI"
	default:
		return errors.New("Mode impor tidak dikenal")
	}
	return nil
}

func bacaPengaturan(q pembaca) (model.Pengaturan, error) {
	var p model.Pengaturan
	err := q.QueryRow(`SELECT COALESCE(nama_instansi,''),COALESCE(alamat_instansi,''),
		COALESCE(penanggung_jawab,''),COALESCE(nip_penanggung_jawab,''),COALESCE(jabatan_penanggung_jawab,''),
		jeda_minimum_menit,hid_ambang_ms,COALESCE(port_serial,'') FROM pengaturan WHERE id=1`).
		Scan(&p.NamaInstansi, &p.AlamatInstansi, &p.PenanggungJawab, &p.NIPPenanggungJawab,
			&p.JabatanPenanggungJawab, &p.JedaMinimumMenit, &p.HIDAmbangMs, &p.PortSerial)
	return p, err
}

func muatKelasLama(q pembaca) (map[string]*kelasLama, error) {
	rows, err := q.Query(`SELECT id,nama,nama_ketua,urutan,aktif FROM kelas`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]*kelasLama{}
	for rows.Next() {
		var id int64
		var nama, ketua string
		var urutan, aktif int
		if err := rows.Scan(&id, &nama, &ketua, &urutan, &aktif); err != nil {
			return nil, err
		}
		out[nama] = &kelasLama{id: id, ketua: ketua, urutan: urutan, aktif: aktif == 1}
	}
	return out, rows.Err()
}

func muatKartuLama(q pembaca) (map[string]*kartuLama, error) {
	rows, err := q.Query(`SELECT c.uid,k.nama,COALESCE(c.nisn,''),COALESCE(c.nama_siswa,''),c.aktif
		FROM kartu c JOIN kelas k ON k.id=c.kelas_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]*kartuLama{}
	for rows.Next() {
		var uid, kelas, nisn, nama string
		var aktif int
		if err := rows.Scan(&uid, &kelas, &nisn, &nama, &aktif); err != nil {
			return nil, err
		}
		out[uid] = &kartuLama{kelas: kelas, nisn: nisn, nama: nama, aktif: aktif == 1}
	}
	return out, rows.Err()
}

func kunciPengaturanDikenal(k string) bool {
	for _, x := range sinkron.KunciPengaturan {
		if x == k {
			return true
		}
	}
	return false
}

// ---------- perencanaan ----------

// rencanakan: menghitung apa yang akan terjadi tanpa menulis apa pun.
func (s *Service) rencanakan(q pembaca, d sinkron.Data, o model.OpsiImpor) (*rencanaImpor, error) {
	rc := &rencanaImpor{}
	catat := func(ent string, baris int, kunci, aksi, pesan string) {
		rc.baris = append(rc.baris, model.BarisPratinjau{Entitas: ent, Baris: baris, Kunci: kunci, Aksi: aksi, Pesan: pesan})
		switch aksi {
		case "GALAT":
			rc.r.Galat++
		case "PERINGATAN":
			rc.r.Peringatan++
		}
	}

	kelasDB, err := muatKelasLama(q)
	if err != nil {
		return nil, err
	}
	kartuDB, err := muatKartuLama(q)
	if err != nil {
		return nil, err
	}

	// --- pengaturan ---
	if d.Pengaturan != nil {
		lama, err := bacaPengaturan(q)
		if err != nil {
			return nil, err
		}
		rc.pengaturanLama = lama
		cur := map[string]string{
			"nama_instansi": lama.NamaInstansi, "alamat_instansi": lama.AlamatInstansi,
			"penanggung_jawab": lama.PenanggungJawab, "nip_penanggung_jawab": lama.NIPPenanggungJawab,
			"jabatan_penanggung_jawab": lama.JabatanPenanggungJawab,
			"jeda_minimum_menit":       strconv.Itoa(lama.JedaMinimumMenit),
		}
		rc.pengaturan = map[string]string{}
		for _, k := range sinkron.KunciPengaturan {
			v, ada := d.Pengaturan[k]
			if !ada {
				continue
			}
			v = strings.TrimSpace(v)
			if k == "jeda_minimum_menit" {
				n, err := strconv.Atoi(v)
				if err != nil || n < 0 || n > 600 {
					catat("pengaturan", 0, k, "GALAT", "Jeda minimum harus bilangan 0-600")
					continue
				}
				v = strconv.Itoa(n)
			}
			if v != cur[k] {
				rc.pengaturan[k] = v
				rc.r.PengaturanUbah++
				catat("pengaturan", 0, k, "UBAH", "Nilai diperbarui")
			}
		}
		var asing []string
		for k := range d.Pengaturan {
			if !kunciPengaturanDikenal(k) {
				asing = append(asing, k)
			}
		}
		sort.Strings(asing)
		for _, k := range asing {
			catat("pengaturan", 0, k, "PERINGATAN", "Kunci tidak dikenal, diabaikan")
		}
	}

	// --- kelas ---
	dalamBerkas := map[string]bool{}
	otomatis := map[string]bool{}
	dipakai := map[string]bool{}
	if d.AdaKelas {
		if o.Mode == "GANTI" && len(d.Kelas) == 0 {
			catat("kelas", 0, "-", "PERINGATAN", "Berkas tanpa baris kelas: mode Ganti akan menonaktifkan semua kelas")
		}
		for _, b := range d.Kelas {
			nama := strings.TrimSpace(b.Nama)
			if nama == "" {
				catat("kelas", b.Baris, "-", "GALAT", "Nama kelas kosong")
				continue
			}
			if dalamBerkas[nama] {
				catat("kelas", b.Baris, nama, "GALAT", "Nama kelas ganda dalam berkas")
				continue
			}
			dalamBerkas[nama] = true
			ketua := strings.TrimSpace(b.Ketua)
			if ketua == "" {
				catat("kelas", b.Baris, nama, "GALAT", "Nama ketua kosong")
				continue
			}
			urutan := 0
			if u := strings.TrimSpace(b.Urutan); u != "" {
				n, err := strconv.Atoi(u)
				if err != nil {
					catat("kelas", b.Baris, nama, "GALAT", "Urutan harus bilangan bulat")
					continue
				}
				urutan = n
			}
			aktif, ok := parseAktif(b.Aktif)
			if !ok {
				catat("kelas", b.Baris, nama, "GALAT", "Kolom aktif harus 1 atau 0")
				continue
			}
			kr := kelasRencana{baris: b.Baris, nama: nama, ketua: ketua, urutan: urutan, aktif: aktif}
			if lama, ada := kelasDB[nama]; ada {
				kr.lama = lama
				if lama.ketua == ketua && lama.urutan == urutan && lama.aktif == aktif {
					kr.aksi = "SAMA"
					rc.r.KelasSama++
				} else {
					kr.aksi = "UBAH"
					rc.r.KelasUbah++
					catat("kelas", b.Baris, nama, "UBAH", "Data kelas diperbarui")
				}
			} else {
				kr.aksi = "BARU"
				rc.r.KelasBaru++
				catat("kelas", b.Baris, nama, "BARU", "Kelas baru")
			}
			rc.kelas = append(rc.kelas, kr)
		}
	}
	kelasKenal := func(n string) bool { return kelasDB[n] != nil || dalamBerkas[n] || otomatis[n] }

	// --- kartu ---
	uidBerkas := map[string]bool{}
	if d.AdaKartu {
		if o.Mode == "GANTI" && len(d.Kartu) == 0 {
			catat("kartu", 0, "-", "PERINGATAN", "Berkas tanpa baris kartu: mode Ganti akan menghapus semua kartu")
		}
		for _, b := range d.Kartu {
			raw := strings.TrimSpace(b.UID)
			if !semuaAngka(raw) || len(raw) > 10 {
				catat("kartu", b.Baris, raw, "GALAT", "UID bukan angka 1-10 digit")
				continue
			}
			uid, _ := reader.NormalisasiUID(raw)
			if uidBerkas[uid] {
				catat("kartu", b.Baris, uid, "GALAT", "UID ganda dalam berkas")
				continue
			}
			uidBerkas[uid] = true
			kelas := strings.TrimSpace(b.Kelas)
			if kelas == "" {
				catat("kartu", b.Baris, uid, "GALAT", "Kelas kosong")
				continue
			}
			nisn := strings.TrimSpace(b.NISN)
			if nisn != "" && (!semuaAngka(nisn) || len(nisn) > 10) {
				catat("kartu", b.Baris, uid, "GALAT", "NISN harus angka maksimal 10 digit")
				continue
			}
			aktif, ok := parseAktif(b.Aktif)
			if !ok {
				catat("kartu", b.Baris, uid, "GALAT", "Kolom aktif harus 1 atau 0")
				continue
			}
			if !kelasKenal(kelas) {
				if !o.BuatKelasOtomatis {
					catat("kartu", b.Baris, uid, "GALAT", fmt.Sprintf("Kelas %q tidak dikenal", kelas))
					continue
				}
				otomatis[kelas] = true
				rc.kelas = append(rc.kelas, kelasRencana{baris: b.Baris, nama: kelas, ketua: "-", aktif: true, aksi: "OTOMATIS"})
				rc.r.KelasOtomatis++
				catat("kelas", b.Baris, kelas, "OTOMATIS", `Dibuat otomatis (ketua "-")`)
			}
			if nisn != "" && len(nisn) < 10 {
				catat("kartu", b.Baris, uid, "PERINGATAN", "NISN kurang dari 10 digit (nol di depan mungkin hilang)")
			}
			dipakai[kelas] = true
			kr := kartuRencana{baris: b.Baris, uid: uid, kelas: kelas, nisn: nisn, nama: strings.TrimSpace(b.Nama), aktif: aktif}
			if lama, ada := kartuDB[uid]; ada {
				kr.lama = lama
				if lama.kelas == kelas && lama.nisn == kr.nisn && lama.nama == kr.nama && lama.aktif == aktif {
					kr.aksi = "SAMA"
					rc.r.KartuSama++
				} else {
					kr.aksi = "UBAH"
					rc.r.KartuUbah++
					pesan := "Data kartu diperbarui"
					if lama.kelas != kelas {
						pesan = fmt.Sprintf("Pindah kelas %s → %s", lama.kelas, kelas)
					}
					catat("kartu", b.Baris, uid, "UBAH", pesan)
				}
			} else {
				kr.aksi = "BARU"
				rc.r.KartuBaru++
				catat("kartu", b.Baris, uid, "BARU", "Kartu baru, kelas "+kelas)
			}
			rc.kartu = append(rc.kartu, kr)
		}
	}

	// --- mode Ganti ---
	if o.Mode == "GANTI" {
		if d.AdaKartu {
			var uids []string
			for uid := range kartuDB {
				if !uidBerkas[uid] {
					uids = append(uids, uid)
				}
			}
			sort.Strings(uids)
			for _, uid := range uids {
				rc.kartu = append(rc.kartu, kartuRencana{uid: uid, kelas: kartuDB[uid].kelas, aksi: "HAPUS", lama: kartuDB[uid]})
				rc.r.KartuHapus++
				catat("kartu", 0, uid, "HAPUS", "Tidak ada di berkas: kartu dihapus")
			}
		}
		if d.AdaKelas {
			var nama []string
			for n, k := range kelasDB {
				if k.aktif && !dalamBerkas[n] && !dipakai[n] {
					nama = append(nama, n)
				}
			}
			sort.Strings(nama)
			for _, n := range nama {
				rc.kelas = append(rc.kelas, kelasRencana{nama: n, aksi: "NONAKTIF", lama: kelasDB[n]})
				rc.r.KelasNonaktif++
				catat("kelas", 0, n, "NONAKTIF", "Tidak ada di berkas: kelas dinonaktifkan")
			}
		}
	}
	return rc, nil
}

// ---------- pratinjau dan terapkan ----------

// PilihBerkasImpor: dialog pilih berkas .xlsx/.csv. Mengembalikan "" bila dibatalkan.
func (s *Service) PilihBerkasImpor() (string, error) {
	if s.app.ctx == nil {
		return "", errors.New("Dialog file belum siap")
	}
	opts := wruntime.OpenDialogOptions{
		Title:   "Pilih berkas impor master",
		Filters: []wruntime.FileFilter{{DisplayName: "Excel/CSV (*.xlsx;*.csv)", Pattern: "*.xlsx;*.csv"}},
	}
	if d := folderAwal(); d != "" {
		opts.DefaultDirectory = d
	}
	p, err := wruntime.OpenFileDialog(s.app.ctx, opts)
	if err != nil {
		return "", err
	}
	return p, nil
}

// PratinjauImpor: tidak menulis apa pun ke database.
func (s *Service) PratinjauImpor(path string, o model.OpsiImpor) (model.HasilPratinjau, error) {
	var out model.HasilPratinjau
	if err := normalMode(&o); err != nil {
		return out, err
	}
	d, err := sinkron.Baca(path)
	if err != nil {
		return out, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rc, err := s.rencanakan(s.db, d, o)
	if err != nil {
		return out, err
	}
	return rc.hasil(path, rc.hitungHash(d.SHA, o), o), nil
}

// TerapkanImpor: seluruh impor dalam satu transaksi; satu galat berarti tidak ada yang berubah.
// o.Hash harus berasal dari PratinjauImpor dengan berkas dan opsi yang sama.
func (s *Service) TerapkanImpor(path string, o model.OpsiImpor) (model.HasilImpor, error) {
	var out model.HasilImpor
	if err := normalMode(&o); err != nil {
		return out, err
	}
	o.Alasan = strings.TrimSpace(o.Alasan)
	if o.Mode == "GANTI" {
		if !o.KonfirmasiGanti {
			return out, errors.New("Mode Ganti memerlukan konfirmasi eksplisit")
		}
		if o.Alasan == "" {
			return out, errors.New("Alasan wajib diisi untuk mode Ganti")
		}
	}
	d, err := sinkron.Baca(path)
	if err != nil {
		return out, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		return out, err
	}
	defer tx.Rollback()

	rc, err := s.rencanakan(tx, d, o)
	if err != nil {
		return out, err
	}
	if rc.r.Galat > 0 {
		return out, fmt.Errorf("Impor dibatalkan: %d galat pada berkas, tidak ada data yang diubah", rc.r.Galat)
	}
	if o.Hash == "" || o.Hash != rc.hitungHash(d.SHA, o) {
		return out, errors.New("Berkas atau data berubah sejak pratinjau; ulangi pratinjau")
	}
	out.Ringkasan = rc.r
	if rc.perubahan() == 0 {
		out.Pesan = "Tidak ada perubahan: data sudah sama dengan berkas"
		return out, nil
	}
	alasan := o.Alasan
	if alasan == "" {
		alasan = "impor master"
	}

	// pengaturan
	if len(rc.pengaturan) > 0 {
		keys := make([]string, 0, len(rc.pengaturan))
		for k := range rc.pengaturan {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys { // k berasal dari daftar putih KunciPengaturan
			var arg any = rc.pengaturan[k]
			if k == "jeda_minimum_menit" {
				n, _ := strconv.Atoi(rc.pengaturan[k])
				arg = n
			}
			if _, err := tx.Exec(`UPDATE pengaturan SET `+k+`=? WHERE id=1`, arg); err != nil {
				return out, err
			}
		}
		if err := s.audit(tx, now, "UBAH", "pengaturan", "1", rc.pengaturanLama, rc.pengaturan, alasan); err != nil {
			return out, err
		}
	}

	// kelas
	for _, k := range rc.kelas {
		switch k.aksi {
		case "BARU", "OTOMATIS":
			res, err := tx.Exec(`INSERT INTO kelas (nama,nama_ketua,urutan,aktif) VALUES (?,?,?,?)`,
				k.nama, k.ketua, k.urutan, boolInt(k.aktif))
			if err != nil {
				return out, err
			}
			id, _ := res.LastInsertId()
			if err := s.audit(tx, now, "TAMBAH", "kelas", itoa(id), nil,
				map[string]any{"nama": k.nama, "ketua": k.ketua, "urutan": k.urutan, "aktif": k.aktif}, alasan); err != nil {
				return out, err
			}
		case "UBAH":
			if _, err := tx.Exec(`UPDATE kelas SET nama_ketua=?,urutan=?,aktif=? WHERE id=?`,
				k.ketua, k.urutan, boolInt(k.aktif), k.lama.id); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "UBAH", "kelas", itoa(k.lama.id),
				map[string]any{"nama": k.nama, "ketua": k.lama.ketua, "urutan": k.lama.urutan, "aktif": k.lama.aktif},
				map[string]any{"nama": k.nama, "ketua": k.ketua, "urutan": k.urutan, "aktif": k.aktif}, alasan); err != nil {
				return out, err
			}
		case "NONAKTIF":
			if _, err := tx.Exec(`UPDATE kelas SET aktif=0 WHERE id=?`, k.lama.id); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "UBAH", "kelas", itoa(k.lama.id),
				map[string]any{"aktif": 1}, map[string]any{"aktif": 0}, alasan); err != nil {
				return out, err
			}
		}
	}

	// kartu
	idKelas := func(nama string) (int64, error) {
		var id int64
		if err := tx.QueryRow(`SELECT id FROM kelas WHERE nama=?`, nama).Scan(&id); err != nil {
			return 0, fmt.Errorf("Kelas %q tidak ditemukan", nama)
		}
		return id, nil
	}
	for _, c := range rc.kartu {
		switch c.aksi {
		case "BARU":
			kid, err := idKelas(c.kelas)
			if err != nil {
				return out, err
			}
			if _, err := tx.Exec(`INSERT INTO kartu (uid,kelas_id,aktif,nisn,nama_siswa) VALUES (?,?,?,?,?)`,
				c.uid, kid, boolInt(c.aktif), c.nisn, c.nama); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "TAMBAH", "kartu", c.uid, nil,
				map[string]any{"kelas": c.kelas, "nisn": c.nisn, "nama": c.nama, "aktif": c.aktif}, alasan); err != nil {
				return out, err
			}
		case "UBAH":
			kid, err := idKelas(c.kelas)
			if err != nil {
				return out, err
			}
			if _, err := tx.Exec(`UPDATE kartu SET kelas_id=?,aktif=?,nisn=?,nama_siswa=? WHERE uid=?`,
				kid, boolInt(c.aktif), c.nisn, c.nama, c.uid); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "UBAH", "kartu", c.uid,
				map[string]any{"kelas": c.lama.kelas, "nisn": c.lama.nisn, "nama": c.lama.nama, "aktif": c.lama.aktif},
				map[string]any{"kelas": c.kelas, "nisn": c.nisn, "nama": c.nama, "aktif": c.aktif}, alasan); err != nil {
				return out, err
			}
		case "HAPUS":
			if _, err := tx.Exec(`DELETE FROM kartu WHERE uid=?`, c.uid); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "HAPUS", "kartu", c.uid,
				map[string]any{"kelas": c.lama.kelas, "nisn": c.lama.nisn, "nama": c.lama.nama, "aktif": c.lama.aktif}, nil, alasan); err != nil {
				return out, err
			}
		}
	}

	if err := s.audit(tx, now, "IMPOR_MASTER", "master", "", nil,
		map[string]any{"mode": o.Mode, "berkas": filepath.Base(path), "ringkasan": rc.r}, alasan); err != nil {
		return out, err
	}
	if err := tx.Commit(); err != nil {
		return out, err
	}
	r := rc.r
	out.Pesan = fmt.Sprintf("Impor selesai: kelas %d baru, %d diubah, %d dinonaktifkan; kartu %d baru, %d diubah, %d dihapus; pengaturan %d diubah",
		r.KelasBaru+r.KelasOtomatis, r.KelasUbah, r.KelasNonaktif, r.KartuBaru, r.KartuUbah, r.KartuHapus, r.PengaturanUbah)
	return out, nil
}

// ---------- ekspor ----------

func (s *Service) dataMaster(denganPengaturan bool) (sinkron.Data, error) {
	d := sinkron.Data{AdaKelas: true, AdaKartu: true}
	if denganPengaturan {
		p, err := s.GetPengaturan()
		if err != nil {
			return d, err
		}
		d.Pengaturan = map[string]string{
			"nama_instansi": p.NamaInstansi, "alamat_instansi": p.AlamatInstansi,
			"penanggung_jawab": p.PenanggungJawab, "nip_penanggung_jawab": p.NIPPenanggungJawab,
			"jabatan_penanggung_jawab": p.JabatanPenanggungJawab,
			"jeda_minimum_menit":       strconv.Itoa(p.JedaMinimumMenit),
		}
	}
	kelas, err := s.db.Query(`SELECT nama,nama_ketua,urutan,aktif FROM kelas ORDER BY urutan,nama`)
	if err != nil {
		return d, err
	}
	for kelas.Next() {
		var nama, ketua string
		var urutan, aktif int
		if err := kelas.Scan(&nama, &ketua, &urutan, &aktif); err != nil {
			kelas.Close()
			return d, err
		}
		d.Kelas = append(d.Kelas, sinkron.KelasBaris{Nama: nama, Ketua: ketua, Urutan: strconv.Itoa(urutan), Aktif: strconv.Itoa(aktif)})
	}
	if err := kelas.Err(); err != nil {
		kelas.Close()
		return d, err
	}
	kelas.Close()

	kartu, err := s.db.Query(`SELECT c.uid,k.nama,COALESCE(c.nisn,''),COALESCE(c.nama_siswa,''),c.aktif
		FROM kartu c JOIN kelas k ON k.id=c.kelas_id ORDER BY k.urutan,k.nama,c.nama_siswa,c.uid`)
	if err != nil {
		return d, err
	}
	for kartu.Next() {
		var uid, nama, nisn, siswa string
		var aktif int
		if err := kartu.Scan(&uid, &nama, &nisn, &siswa, &aktif); err != nil {
			kartu.Close()
			return d, err
		}
		d.Kartu = append(d.Kartu, sinkron.KartuBaris{UID: uid, Kelas: nama, NISN: nisn, Nama: siswa, Aktif: strconv.Itoa(aktif)})
	}
	err = kartu.Err()
	kartu.Close()
	return d, err
}

func folderAwal() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, sub := range []string{"Documents", "Dokumen", "Downloads", "Unduhan"} {
		p := filepath.Join(h, sub)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return h
}

// tolakDalamRepo: berkas ekspor berisi NISN dan nama siswa, jangan sampai masuk repositori git.
func tolakDalamRepo(p string) error {
	dir, err := filepath.Abs(filepath.Dir(p))
	if err != nil {
		return nil
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return errors.New("Berkas berisi data pribadi siswa (NISN dan nama) dan tidak boleh disimpan di dalam repositori git. Pilih folder lain.")
		}
		induk := filepath.Dir(dir)
		if induk == dir {
			return nil
		}
		dir = induk
	}
}

func (s *Service) pilihSimpan(judul, nama, ext, label string, jagaRepo bool) (string, error) {
	if s.app.ctx == nil {
		return "", errors.New("Dialog file belum siap")
	}
	opts := wruntime.SaveDialogOptions{
		Title: judul, DefaultFilename: nama,
		Filters: []wruntime.FileFilter{{DisplayName: label, Pattern: "*." + ext}},
	}
	if d := folderAwal(); d != "" {
		opts.DefaultDirectory = d
	}
	p, err := wruntime.SaveFileDialog(s.app.ctx, opts)
	if err != nil || p == "" {
		return "", err
	}
	if !strings.HasSuffix(strings.ToLower(p), "."+ext) {
		p += "." + ext
	}
	if jagaRepo {
		if err := tolakDalamRepo(p); err != nil {
			return "", err
		}
	}
	return p, nil
}

// EksporMasterExcel: sheet Pengaturan, Kelas, Kartu. Mengembalikan path, atau "" bila dialog dibatalkan.
func (s *Service) EksporMasterExcel() (string, error) {
	d, err := s.dataMaster(true)
	if err != nil {
		return "", err
	}
	p, err := s.pilihSimpan("Simpan ekspor master", "master-mbg-"+timeutil.Tanggal(s.now())+".xlsx", "xlsx", "Excel (*.xlsx)", true)
	if err != nil || p == "" {
		return "", err
	}
	if err := sinkron.TulisXLSX(p, d, nil); err != nil {
		return "", err
	}
	return p, nil
}

// EksporMasterCSV: jenis "kelas" atau "kartu".
func (s *Service) EksporMasterCSV(jenis string) (string, error) {
	if jenis != "kelas" && jenis != "kartu" {
		return "", errors.New("Jenis data harus kelas atau kartu")
	}
	d, err := s.dataMaster(false)
	if err != nil {
		return "", err
	}
	p, err := s.pilihSimpan("Simpan ekspor "+jenis, "master-"+jenis+"-"+timeutil.Tanggal(s.now())+".csv", "csv", "CSV (*.csv)", true)
	if err != nil || p == "" {
		return "", err
	}
	if jenis == "kelas" {
		err = sinkron.TulisCSVKelas(p, d)
	} else {
		err = sinkron.TulisCSVKartu(p, d)
	}
	if err != nil {
		return "", err
	}
	return p, nil
}

func (s *Service) UnduhTemplateExcel() (string, error) {
	p, err := s.pilihSimpan("Simpan template impor", "template-impor-master.xlsx", "xlsx", "Excel (*.xlsx)", false)
	if err != nil || p == "" {
		return "", err
	}
	if err := sinkron.TemplateXLSX(p); err != nil {
		return "", err
	}
	return p, nil
}

func (s *Service) UnduhTemplateCSV(jenis string) (string, error) {
	if jenis != "kelas" && jenis != "kartu" {
		return "", errors.New("Jenis template harus kelas atau kartu")
	}
	p, err := s.pilihSimpan("Simpan template "+jenis, "template-"+jenis+".csv", "csv", "CSV (*.csv)", false)
	if err != nil || p == "" {
		return "", err
	}
	if err := sinkron.TemplateCSV(p, jenis); err != nil {
		return "", err
	}
	return p, nil
}
