package service

import (
	"path/filepath"
	"testing"
	"time"

	"mbg/internal/db"
	"mbg/internal/model"
	"mbg/internal/timeutil"
)

const (
	uidA = "1111111111" // Ani, VII A
	uidB = "2222222222" // Budi, VII A
)

// siapkanAmbil: VII A mengambil 10 pukul 10:00; jam di-maju ke 10:30.
func siapkanAmbil(t *testing.T) (*Service, *time.Time, int64) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	clk := time.Date(2026, 1, 12, 10, 0, 0, 0, timeutil.WIB)
	s := NewWithClock(conn, t.TempDir(), func() time.Time { return clk })
	id, err := s.TambahKelas("VII A", "Ketua A", 1)
	harusOK(t, err)
	harusOK(t, s.DaftarkanKartu(uidA, id, "1234567890", "Ani"))
	harusOK(t, s.DaftarkanKartu(uidB, id, "0987654321", "Budi"))
	if r := s.ProsesTap(uidA); r.Aksi != "AMBIL" {
		t.Fatalf("tap ambil: %+v", r)
	}
	harusOK(t, s.SubmitAmbil(id, 10))
	clk = clk.Add(30 * time.Minute)
	return s, &clk, id
}

// siapkanKurang: lanjutan siapkanAmbil; Ani mengembalikan 7 dari 10 (status KURANG) pukul 10:30.
func siapkanKurang(t *testing.T) (*Service, *time.Time, int64) {
	t.Helper()
	s, now, id := siapkanAmbil(t)
	if r := s.ProsesTap(uidA); r.Aksi != "KEMBALI" {
		t.Fatalf("tap kembali: %+v", r)
	}
	harusOK(t, s.SubmitKembali(id, 7))
	return s, now, id
}

func trxPertama(t *testing.T, s *Service) model.Transaksi {
	t.Helper()
	list, err := s.ListTransaksi(model.FilterTransaksi{})
	harusOK(t, err)
	if len(list) != 1 {
		t.Fatalf("mau 1 transaksi, dapat %d", len(list))
	}
	return list[0]
}

func jumlahRincian(t *testing.T, s *Service) (n, sum int) {
	t.Helper()
	harusOK(t, s.db.QueryRow(`SELECT COUNT(*),COALESCE(SUM(jumlah),0) FROM pengembalian`).Scan(&n, &sum))
	return
}

func TestSusulanAlurLengkap(t *testing.T) {
	s, now, id := siapkanKurang(t)

	trx := trxPertama(t, s)
	if trx.Status != "KURANG" || *trx.JumlahKembali != 7 || trx.Keterangan != "Kurang 3 Ompreng" || !trx.Merah {
		t.Fatalf("setelah kembali 7: %+v", trx)
	}

	*now = now.Add(6 * time.Second)
	r := s.ProsesTap(uidB)
	if r.Aksi != "KEMBALI_SUSULAN" || r.Sisa != 3 || r.JumlahAmbil != 10 || r.NamaSiswa != "Budi" {
		t.Fatalf("tap susulan: %+v", r)
	}
	if err := s.SubmitKembali(id, 0); err == nil {
		t.Fatal("susulan 0 harus ditolak")
	}
	if err := s.SubmitKembali(id, 4); err == nil {
		t.Fatal("susulan melebihi sisa harus ditolak")
	}
	harusOK(t, s.SubmitKembali(id, 3))

	trx = trxPertama(t, s)
	if trx.Status != "SELESAI" || *trx.JumlahKembali != 10 || trx.Keterangan != "Lengkap (susulan)" || trx.Merah {
		t.Fatalf("setelah susulan: %+v", trx)
	}
	if trx.NamaKembali != "Ani, Budi" || len(trx.Pengembalian) != 2 {
		t.Fatalf("pengembali ganda: %q rincian=%+v", trx.NamaKembali, trx.Pengembalian)
	}
	if err := s.SubmitKembali(id, 1); err == nil {
		t.Fatal("submit setelah lengkap harus ditolak")
	}

	want := "Kelas VII A sudah mengembalikan ompreng dengan lengkap"
	for _, uid := range []string{uidA, uidB} { // termasuk kartu siswa lain di kelas yang sama
		*now = now.Add(3 * time.Second)
		if r := s.ProsesTap(uid); r.Aksi != "PESAN" || r.Pesan != want {
			t.Fatalf("tap setelah lengkap (%s): %+v", uid, r)
		}
	}

	au, err := s.ListAudit(model.FilterAudit{Entitas: "transaksi"})
	harusOK(t, err)
	ada := false
	for _, a := range au {
		if a.Alasan == "pengembalian susulan" {
			ada = true
		}
	}
	if !ada {
		t.Fatalf("audit susulan tidak ada: %+v", au)
	}
}

func TestKembaliLengkapLangsungLaluTapDitolak(t *testing.T) {
	s, now, id := siapkanAmbil(t)
	if r := s.ProsesTap(uidA); r.Aksi != "KEMBALI" {
		t.Fatalf("tap kembali: %+v", r)
	}
	harusOK(t, s.SubmitKembali(id, 10))
	trx := trxPertama(t, s)
	if trx.Status != "SELESAI" || trx.Keterangan != "Lengkap" || trx.Merah {
		t.Fatalf("lengkap langsung: %+v", trx)
	}
	want := "Kelas VII A sudah mengembalikan ompreng dengan lengkap"
	for _, uid := range []string{uidB, uidA} {
		*now = now.Add(3 * time.Second)
		if r := s.ProsesTap(uid); r.Aksi != "PESAN" || r.Pesan != want {
			t.Fatalf("tap (%s): %+v", uid, r)
		}
	}
}

func TestCooldownSetelahKurang(t *testing.T) {
	s, now, _ := siapkanKurang(t)

	*now = now.Add(3 * time.Second)
	if r := s.ProsesTap(uidB); r.Aksi != "PESAN" || r.Pesan != "Baru saja tercatat" {
		t.Fatalf("3 detik setelah kurang harus tertahan: %+v", r)
	}
	*now = now.Add(3 * time.Second) // total 6 detik > 5 detik
	if r := s.ProsesTap(uidB); r.Aksi != "KEMBALI_SUSULAN" {
		t.Fatalf("6 detik setelah kurang harus lolos: %+v", r)
	}
}

func TestTutupDanBukaHariSaatKurang(t *testing.T) {
	s, now, id := siapkanKurang(t)

	harusOK(t, s.TutupHari("tutup"))
	st, err := s.StatusHariIni()
	harusOK(t, err)
	if b := st.Baris[0]; b.Status != "BELUM_KEMBALI" || b.Keterangan != "Kurang 3 Ompreng" || !b.Merah {
		t.Fatalf("setelah tutup: %+v", b)
	}
	*now = now.Add(6 * time.Second)
	if r := s.ProsesTap(uidB); r.Pesan != "Hari ini sudah ditutup" {
		t.Fatalf("tap saat ditutup: %+v", r)
	}
	if err := s.SubmitKembali(id, 1); err == nil {
		t.Fatal("susulan saat ditutup harus ditolak")
	}

	harusOK(t, s.BukaHari(tglUji, "salah tutup"))
	st, _ = s.StatusHariIni()
	if st.Baris[0].Status != "KURANG" || st.Ditutup {
		t.Fatalf("setelah buka: %+v", st)
	}
	*now = now.Add(10 * time.Second)
	if r := s.ProsesTap(uidB); r.Aksi != "KEMBALI_SUSULAN" || r.Sisa != 3 {
		t.Fatalf("susulan setelah buka: %+v", r)
	}
}

func TestEksporSusulan(t *testing.T) {
	s, now, id := siapkanKurang(t)

	l, err := s.susunLaporan(tglUji, tglUji)
	harusOK(t, err)
	if x := l.Baris[0]; x.Keterangan != "Kurang 3 Ompreng" || !x.Merah || x.Pengembali != "Ani" || x.WaktuKembali != "10:30 WIB" {
		t.Fatalf("sebelum susulan: %+v", x)
	}

	*now = now.Add(30 * time.Minute)
	if r := s.ProsesTap(uidB); r.Aksi != "KEMBALI_SUSULAN" {
		t.Fatalf("tap susulan: %+v", r)
	}
	harusOK(t, s.SubmitKembali(id, 3))

	l, err = s.susunLaporan(tglUji, tglUji)
	harusOK(t, err)
	if x := l.Baris[0]; x.Keterangan != "Lengkap (susulan)" || x.Merah || x.Pengembali != "Ani, Budi" ||
		x.WaktuKembali != "11:00 WIB" || x.JumlahKembali != "10" {
		t.Fatalf("setelah susulan: %+v", x)
	}
}

func TestMigrasiRincianPengembalian(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	harusOK(t, err)
	t.Cleanup(func() { conn.Close() })

	// simulasi DB lama: tidak ada tabel pengembalian, ada transaksi SELESAI dan AMBIL
	_, err = conn.Exec(`DROP TABLE pengembalian`)
	harusOK(t, err)
	_, err = conn.Exec(`INSERT INTO kelas (id,nama,nama_ketua) VALUES (1,'VII A','Ani'),(2,'VII B','Budi')`)
	harusOK(t, err)
	_, err = conn.Exec(`INSERT INTO transaksi (id,tanggal,kelas_id,nama_kelas,nama_ketua,waktu_ambil,waktu_kembali,jumlah_ambil,jumlah_kembali,status,nama_kembali)
		VALUES (1,'2026-01-10',1,'VII A','Ani','09:00','10:00',10,8,'SELESAI','Budi')`)
	harusOK(t, err)
	_, err = conn.Exec(`INSERT INTO transaksi (id,tanggal,kelas_id,nama_kelas,nama_ketua,waktu_ambil,jumlah_ambil,status)
		VALUES (2,'2026-01-10',2,'VII B','Budi','09:00',5,'AMBIL')`)
	harusOK(t, err)

	for i := 0; i < 2; i++ { // dua kali: harus idempoten
		harusOK(t, db.Migrate(conn))
		var n, jumlah int
		var nama string
		harusOK(t, conn.QueryRow(`SELECT COUNT(*),COALESCE(MAX(jumlah),0),COALESCE(MAX(nama_siswa),'') FROM pengembalian`).Scan(&n, &jumlah, &nama))
		if n != 1 || jumlah != 8 || nama != "Budi" {
			t.Fatalf("migrasi ke-%d: n=%d jumlah=%d nama=%q", i+1, n, jumlah, nama)
		}
	}
}

func TestEditTransaksiMenyinkronkanRincian(t *testing.T) {
	s, now, _ := siapkanKurang(t)
	*now = now.Add(6 * time.Second)
	if r := s.ProsesTap(uidB); r.Aksi != "KEMBALI_SUSULAN" {
		t.Fatalf("tap susulan: %+v", r)
	}
	harusOK(t, s.SubmitKembali(s.idKelasUji(t), 3))
	if n, sum := jumlahRincian(t, s); n != 2 || sum != 10 {
		t.Fatalf("awal: n=%d sum=%d", n, sum)
	}
	trx := trxPertama(t, s)
	ip := func(n int) *int { return &n }
	edit := func(status string, jk *int) model.EditTransaksi {
		return model.EditTransaksi{ID: trx.ID, Status: status, JumlahAmbil: ip(10), JumlahKembali: jk,
			WaktuAmbil: "10:00", WaktuKembali: "11:00", Alasan: "koreksi"}
	}

	harusOK(t, s.UbahTransaksi(edit("SELESAI", ip(10)))) // total sama: rincian dipertahankan
	if n, _ := jumlahRincian(t, s); n != 2 {
		t.Fatalf("total sama harus mempertahankan rincian, n=%d", n)
	}
	harusOK(t, s.UbahTransaksi(edit("SELESAI", ip(9)))) // total berbeda: diganti satu baris koreksi
	if n, sum := jumlahRincian(t, s); n != 1 || sum != 9 {
		t.Fatalf("setelah koreksi: n=%d sum=%d", n, sum)
	}
	if err := s.UbahTransaksi(edit("KURANG", ip(10))); err == nil {
		t.Fatal("KURANG dengan kembali == ambil harus ditolak")
	}
	harusOK(t, s.UbahTransaksi(edit("KURANG", ip(9))))
	if x := trxPertama(t, s); x.Status != "KURANG" || x.Keterangan != "Kurang 1 Ompreng" {
		t.Fatalf("edit ke KURANG: %+v", x)
	}
	harusOK(t, s.UbahTransaksi(model.EditTransaksi{ID: trx.ID, Status: "AMBIL", JumlahAmbil: ip(10), WaktuAmbil: "10:00", Alasan: "batal kembali"}))
	if n, _ := jumlahRincian(t, s); n != 0 {
		t.Fatalf("edit ke AMBIL harus menghapus rincian, n=%d", n)
	}
}

func TestHapusTransaksiMenghapusRincian(t *testing.T) {
	s, _, _ := siapkanKurang(t)
	trx := trxPertama(t, s)
	if n, _ := jumlahRincian(t, s); n != 1 {
		t.Fatalf("rincian awal n=%d", n)
	}
	harusOK(t, s.HapusTransaksi(trx.ID, "dobel"))
	if n, _ := jumlahRincian(t, s); n != 0 {
		t.Fatalf("rincian harus ikut terhapus, n=%d", n)
	}
}

// idKelasUji: ID kelas satu-satunya pada DB uji.
func (s *Service) idKelasUji(t *testing.T) int64 {
	t.Helper()
	var id int64
	harusOK(t, s.db.QueryRow(`SELECT id FROM kelas LIMIT 1`).Scan(&id))
	return id
}
