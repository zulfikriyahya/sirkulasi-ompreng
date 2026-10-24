package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mbg/internal/model"
	"mbg/internal/sinkron"
)

func isiMaster(t *testing.T, s *Service) {
	t.Helper()
	p, err := s.GetPengaturan()
	harusOK(t, err)
	p.NamaInstansi, p.AlamatInstansi = "SMPN 1 Contoh", "Jl. Mawar 1"
	p.PenanggungJawab, p.NIPPenanggungJawab, p.JabatanPenanggungJawab = "Dra. Ani", "0123456789", "Kepala Sekolah"
	p.JedaMinimumMenit = 15
	harusOK(t, s.SimpanPengaturan(p))
	a, err := s.TambahKelas("VII A", "Ani", 1)
	harusOK(t, err)
	b, err := s.TambahKelas("VII B", "Budi", 2)
	harusOK(t, err)
	harusOK(t, s.DaftarkanKartu("0012345678", a, "0123456789", "Siti"))
	harusOK(t, s.DaftarkanKartu("0000000042", a, "", ""))
	harusOK(t, s.DaftarkanKartu("9876543210", b, "1234567890", "Doni"))
	harusOK(t, s.NonaktifkanKelas(b, "uji"))
}

func dumpMaster(t *testing.T, s *Service) string {
	t.Helper()
	var sb strings.Builder
	p, err := s.GetPengaturan()
	harusOK(t, err)
	fmt.Fprintf(&sb, "P|%s|%s|%s|%s|%s|%d\n", p.NamaInstansi, p.AlamatInstansi, p.PenanggungJawab,
		p.NIPPenanggungJawab, p.JabatanPenanggungJawab, p.JedaMinimumMenit)
	rows, err := s.db.Query(`SELECT nama,nama_ketua,urutan,aktif FROM kelas ORDER BY nama`)
	harusOK(t, err)
	for rows.Next() {
		var n, k string
		var u, a int
		harusOK(t, rows.Scan(&n, &k, &u, &a))
		fmt.Fprintf(&sb, "K|%s|%s|%d|%d\n", n, k, u, a)
	}
	rows.Close()
	rows, err = s.db.Query(`SELECT c.uid,k.nama,COALESCE(c.nisn,''),COALESCE(c.nama_siswa,''),c.aktif
		FROM kartu c JOIN kelas k ON k.id=c.kelas_id ORDER BY c.uid`)
	harusOK(t, err)
	for rows.Next() {
		var uid, kelas, nisn, nama string
		var a int
		harusOK(t, rows.Scan(&uid, &kelas, &nisn, &nama, &a))
		fmt.Fprintf(&sb, "C|%s|%s|%s|%s|%d\n", uid, kelas, nisn, nama, a)
	}
	rows.Close()
	return sb.String()
}

func hitung(t *testing.T, s *Service, query string) int {
	t.Helper()
	var n int
	harusOK(t, s.db.QueryRow(query).Scan(&n))
	return n
}

func tulisCSVUji(t *testing.T, isi string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "uji.csv")
	harusOK(t, os.WriteFile(p, []byte("\xEF\xBB\xBF"+isi), 0o600))
	return p
}

// impor: pratinjau lalu terapkan dengan hash dari pratinjau (alur UI).
func impor(t *testing.T, s *Service, path string, o model.OpsiImpor) (model.HasilImpor, error) {
	t.Helper()
	pr, err := s.PratinjauImpor(path, o)
	harusOK(t, err)
	o.Hash = pr.Hash
	return s.TerapkanImpor(path, o)
}

func TestSinkronRoundTripExcel(t *testing.T) {
	asal := setupT5(t)
	isiMaster(t, asal)
	berkas := filepath.Join(t.TempDir(), "master.xlsx")
	d, err := asal.dataMaster(true)
	harusOK(t, err)
	harusOK(t, sinkron.TulisXLSX(berkas, d, nil))

	tujuan := setupT5(t)
	pr, err := tujuan.PratinjauImpor(berkas, model.OpsiImpor{})
	harusOK(t, err)
	r := pr.Ringkasan
	if r.Galat != 0 || r.Peringatan != 0 || r.KelasBaru != 2 || r.KartuBaru != 3 || r.PengaturanUbah != 6 {
		t.Fatalf("pratinjau: %+v", r)
	}
	h, err := impor(t, tujuan, berkas, model.OpsiImpor{})
	harusOK(t, err)
	if h.Ringkasan.KartuBaru != 3 {
		t.Fatalf("hasil: %+v", h)
	}
	got, want := dumpMaster(t, tujuan), dumpMaster(t, asal)
	if got != want {
		t.Fatalf("round-trip tidak identik:\n%s\n---\n%s", got, want)
	}
	if !strings.Contains(got, "C|0012345678|VII A|0123456789|Siti|1") || !strings.Contains(got, "C|0000000042") {
		t.Fatalf("nol di depan UID/NISN hilang:\n%s", got)
	}

	// impor ulang: semuanya sama, tidak ada perubahan
	h, err = impor(t, tujuan, berkas, model.OpsiImpor{})
	harusOK(t, err)
	if !strings.HasPrefix(h.Pesan, "Tidak ada perubahan") {
		t.Fatalf("impor ulang: %+v", h)
	}
}

func TestSinkronGalatBaris(t *testing.T) {
	s := setupT5(t)
	_, err := s.TambahKelas("VII A", "Ani", 1)
	harusOK(t, err)
	p := tulisCSVUji(t, "uid,kelas,nisn,nama_siswa,aktif\n"+
		"1234567890,VII A,1234567890,Ani,1\n"+
		"1234567890,VII A,,Dobel,1\n"+
		"12ab,VII A,,X,1\n"+
		"2222222222,VII Z,,Y,1\n"+
		"3333333333,VII A,12x,Z,1\n")

	audit0, kartu0, kelas0 := hitung(t, s, `SELECT COUNT(*) FROM audit_log`), hitung(t, s, `SELECT COUNT(*) FROM kartu`), hitung(t, s, `SELECT COUNT(*) FROM kelas`)
	pr, err := s.PratinjauImpor(p, model.OpsiImpor{})
	harusOK(t, err)
	if pr.Ringkasan.Galat != 4 || pr.Ringkasan.KartuBaru != 1 {
		t.Fatalf("ringkasan: %+v", pr.Ringkasan)
	}
	var pesan []string
	for _, b := range pr.Baris {
		if b.Aksi == "GALAT" {
			pesan = append(pesan, b.Pesan)
		}
	}
	semua := strings.Join(pesan, "|")
	for _, w := range []string{"ganda", "1-10 digit", "tidak dikenal", "NISN"} {
		if !strings.Contains(semua, w) {
			t.Fatalf("galat %q tidak ada: %s", w, semua)
		}
	}
	// pratinjau tidak menulis apa pun
	if hitung(t, s, `SELECT COUNT(*) FROM audit_log`) != audit0 || hitung(t, s, `SELECT COUNT(*) FROM kartu`) != kartu0 || hitung(t, s, `SELECT COUNT(*) FROM kelas`) != kelas0 {
		t.Fatal("pratinjau menulis ke database")
	}
	// satu galat = seluruh impor batal, termasuk baris yang valid
	if _, err := impor(t, s, p, model.OpsiImpor{}); err == nil {
		t.Fatal("impor dengan galat harus ditolak")
	}
	if hitung(t, s, `SELECT COUNT(*) FROM kartu`) != 0 {
		t.Fatal("impor gagal tidak boleh menulis sebagian")
	}

	// opsi buat kelas otomatis: kelas tidak dikenal tidak lagi galat
	pr, err = s.PratinjauImpor(p, model.OpsiImpor{BuatKelasOtomatis: true})
	harusOK(t, err)
	if pr.Ringkasan.Galat != 3 || pr.Ringkasan.KelasOtomatis != 1 {
		t.Fatalf("otomatis: %+v", pr.Ringkasan)
	}

	p2 := tulisCSVUji(t, "uid,kelas\n2222222222,VII Z\n")
	if _, err := impor(t, s, p2, model.OpsiImpor{}); err == nil {
		t.Fatal("kelas tidak dikenal harus ditolak tanpa opsi otomatis")
	}
	_, err = impor(t, s, p2, model.OpsiImpor{BuatKelasOtomatis: true})
	harusOK(t, err)
	var ketua string
	harusOK(t, s.db.QueryRow(`SELECT nama_ketua FROM kelas WHERE nama='VII Z'`).Scan(&ketua))
	if ketua != "-" || hitung(t, s, `SELECT COUNT(*) FROM kartu WHERE uid='2222222222'`) != 1 {
		t.Fatalf("kelas otomatis: ketua=%q", ketua)
	}
}

func TestSinkronGabungDanGanti(t *testing.T) {
	s := setupT5(t)
	a, _ := s.TambahKelas("VII A", "Ani", 1)
	b, _ := s.TambahKelas("VII B", "Budi", 2)
	harusOK(t, s.DaftarkanKartu("1111111111", a, "", ""))
	harusOK(t, s.DaftarkanKartu("2222222222", b, "", ""))

	d := sinkron.Data{
		AdaKelas: true, AdaKartu: true,
		Kelas: []sinkron.KelasBaris{{Baris: 2, Nama: "VII A", Ketua: "Ani Baru", Urutan: "1", Aktif: "1"}},
		Kartu: []sinkron.KartuBaris{
			{Baris: 2, UID: "1111111111", Kelas: "VII A", Aktif: "1"},
			{Baris: 3, UID: "12345678", Kelas: "VII A", Nama: "Eka", Aktif: "1"}, // nol di depan dinormalkan
		},
	}
	berkas := filepath.Join(t.TempDir(), "uji.xlsx")
	harusOK(t, sinkron.TulisXLSX(berkas, d, nil))

	// Gabung: tidak menghapus apa pun
	_, err := impor(t, s, berkas, model.OpsiImpor{Mode: "GABUNG"})
	harusOK(t, err)
	var ketua string
	harusOK(t, s.db.QueryRow(`SELECT nama_ketua FROM kelas WHERE nama='VII A'`).Scan(&ketua))
	if ketua != "Ani Baru" || hitung(t, s, `SELECT COUNT(*) FROM kartu`) != 3 ||
		hitung(t, s, `SELECT COUNT(*) FROM kartu WHERE uid='0012345678'`) != 1 ||
		hitung(t, s, `SELECT aktif FROM kelas WHERE nama='VII B'`) != 1 {
		t.Fatalf("setelah Gabung: ketua=%q", ketua)
	}

	// Ganti tanpa konfirmasi atau alasan ditolak
	if _, err := impor(t, s, berkas, model.OpsiImpor{Mode: "GANTI"}); err == nil {
		t.Fatal("Ganti tanpa konfirmasi harus ditolak")
	}
	if _, err := impor(t, s, berkas, model.OpsiImpor{Mode: "GANTI", KonfirmasiGanti: true}); err == nil {
		t.Fatal("Ganti tanpa alasan harus ditolak")
	}
	_, err = impor(t, s, berkas, model.OpsiImpor{Mode: "GANTI", KonfirmasiGanti: true, Alasan: "sinkron ulang"})
	harusOK(t, err)
	if hitung(t, s, `SELECT COUNT(*) FROM kartu`) != 2 || hitung(t, s, `SELECT COUNT(*) FROM kartu WHERE uid='2222222222'`) != 0 {
		t.Fatal("Ganti harus menghapus kartu yang tidak ada di berkas")
	}
	if hitung(t, s, `SELECT aktif FROM kelas WHERE nama='VII B'`) != 0 || hitung(t, s, `SELECT aktif FROM kelas WHERE nama='VII A'`) != 1 {
		t.Fatal("Ganti harus menonaktifkan (bukan menghapus) kelas yang tidak ada di berkas")
	}

	// audit: ringkasan per impor dan entri per kartu
	au, err := s.ListAudit(model.FilterAudit{Entitas: "master"})
	harusOK(t, err)
	if len(au) != 2 || au[0].Aksi != "IMPOR_MASTER" || au[0].Alasan != "sinkron ulang" {
		t.Fatalf("audit master: %+v", au)
	}
	au, err = s.ListAudit(model.FilterAudit{Entitas: "kartu"})
	harusOK(t, err)
	hapus := false
	for _, x := range au {
		if x.Aksi == "HAPUS" && x.EntitasID == "2222222222" {
			hapus = true
		}
	}
	if !hapus {
		t.Fatalf("audit hapus kartu tidak ada: %+v", au)
	}
}

func TestSinkronHashBerubah(t *testing.T) {
	s := setupT5(t)
	_, err := s.TambahKelas("VII A", "Ani", 1)
	harusOK(t, err)
	p := tulisCSVUji(t, "uid,kelas\n1111111111,VII A\n")
	pr, err := s.PratinjauImpor(p, model.OpsiImpor{})
	harusOK(t, err)
	// berkas diubah setelah pratinjau
	harusOK(t, os.WriteFile(p, []byte("uid,kelas\n1111111111,VII A\n2222222222,VII A\n"), 0o600))
	if _, err := s.TerapkanImpor(p, model.OpsiImpor{Hash: pr.Hash}); err == nil || !strings.Contains(err.Error(), "berubah") {
		t.Fatalf("hash berbeda harus ditolak: %v", err)
	}
	if _, err := s.TerapkanImpor(p, model.OpsiImpor{}); err == nil {
		t.Fatal("tanpa hash harus ditolak")
	}
	if hitung(t, s, `SELECT COUNT(*) FROM kartu`) != 0 {
		t.Fatal("impor yang ditolak tidak boleh menulis")
	}
}

func TestSinkronCSVTemplateDanSemikolon(t *testing.T) {
	s := setupT5(t)
	isiMaster(t, s)
	d, err := s.dataMaster(false)
	harusOK(t, err)

	p := filepath.Join(t.TempDir(), "kartu.csv")
	harusOK(t, sinkron.TulisCSVKartu(p, d))
	raw, err := os.ReadFile(p)
	harusOK(t, err)
	if !strings.HasPrefix(string(raw), "\xEF\xBB\xBF") {
		t.Fatal("CSV harus berawalan BOM UTF-8")
	}
	got, err := sinkron.Baca(p)
	harusOK(t, err)
	if len(got.Kartu) != 3 || got.AdaKelas || got.Kartu[0].UID != "0012345678" && got.Kartu[1].UID != "0012345678" {
		t.Fatalf("baca ulang CSV: %+v", got)
	}

	// template Excel bisa diimpor tanpa galat dan tanpa perubahan
	tp := filepath.Join(t.TempDir(), "template.xlsx")
	harusOK(t, sinkron.TemplateXLSX(tp))
	tujuan := setupT5(t)
	pr, err := tujuan.PratinjauImpor(tp, model.OpsiImpor{})
	harusOK(t, err)
	if pr.Ringkasan != (model.RingkasanImpor{}) {
		t.Fatalf("template: %+v", pr.Ringkasan)
	}

	// template CSV terbaca
	tc := filepath.Join(t.TempDir(), "tk.csv")
	harusOK(t, sinkron.TemplateCSV(tc, "kartu"))
	g3, err := sinkron.Baca(tc)
	harusOK(t, err)
	if !g3.AdaKartu || len(g3.Kartu) != 0 {
		t.Fatalf("template CSV: %+v", g3)
	}

	// pemisah titik koma (Excel lokal Indonesia)
	g2, err := sinkron.Baca(tulisCSVUji(t, "uid;kelas\n12345678;VII A\n"))
	harusOK(t, err)
	if len(g2.Kartu) != 1 || g2.Kartu[0].UID != "12345678" || g2.Kartu[0].Kelas != "VII A" {
		t.Fatalf("CSV titik koma: %+v", g2)
	}
}
