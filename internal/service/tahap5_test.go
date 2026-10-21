package service

import (
	"path/filepath"
	"testing"
	"time"

	"mbg/internal/db"
	"mbg/internal/model"
	"mbg/internal/timeutil"
)

const tglUji = "2026-01-12"

func setupT5(t *testing.T) *Service {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	fixed := time.Date(2026, 1, 12, 10, 0, 0, 0, timeutil.WIB)
	return NewWithClock(conn, t.TempDir(), func() time.Time { return fixed })
}

func harusOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func statusPerKelas(t *testing.T, s *Service) (map[int64]string, model.StatusHari) {
	t.Helper()
	st, err := s.StatusHariIni()
	harusOK(t, err)
	m := map[int64]string{}
	for _, b := range st.Baris {
		m[b.KelasID] = b.Status
	}
	return m, st
}

func TestTutupDanBukaHari(t *testing.T) {
	s := setupT5(t)
	a, _ := s.TambahKelas("VII A", "Ani", 1)
	b, _ := s.TambahKelas("VII B", "Budi", 2)
	c, _ := s.TambahKelas("VII C", "Cici", 3)
	harusOK(t, s.DaftarkanKartu("1111111111", a))
	harusOK(t, s.SubmitAmbil(a, 10))
	harusOK(t, s.SubmitAmbil(b, 8))
	harusOK(t, s.SubmitKembali(b, 8))

	if ok, _ := s.ApakahBolehCetak(tglUji); ok {
		t.Fatal("belum boleh cetak sebelum semua selesai/ditutup")
	}
	harusOK(t, s.TutupHari("uji"))
	m, st := statusPerKelas(t, s)
	if m[a] != "BELUM_KEMBALI" || m[b] != "SELESAI" || m[c] != "TIDAK_AMBIL" {
		t.Fatalf("status setelah tutup salah: %v", m)
	}
	if !st.Ditutup || !st.BolehCetak {
		t.Fatalf("Ditutup=%v BolehCetak=%v", st.Ditutup, st.BolehCetak)
	}
	if err := s.TutupHari("lagi"); err == nil {
		t.Fatal("tutup dua kali seharusnya ditolak")
	}
	if r := s.ProsesTap("1111111111"); r.Pesan != "Hari ini sudah ditutup" {
		t.Fatalf("tap saat ditutup: %+v", r)
	}

	if err := s.BukaHari(tglUji, ""); err == nil {
		t.Fatal("buka tanpa alasan seharusnya ditolak")
	}
	if err := s.BukaHari("2026-01-11", "x"); err == nil {
		t.Fatal("buka hari lampau seharusnya ditolak")
	}
	harusOK(t, s.BukaHari(tglUji, "salah tutup"))
	m, st = statusPerKelas(t, s)
	if m[a] != "AMBIL" || m[b] != "SELESAI" || m[c] != "BELUM_AMBIL" || st.Ditutup {
		t.Fatalf("status setelah buka salah: %v ditutup=%v", m, st.Ditutup)
	}
}

func TestUbahDanHapusTransaksi(t *testing.T) {
	s := setupT5(t)
	a, _ := s.TambahKelas("VII A", "Ani", 1)
	harusOK(t, s.SubmitAmbil(a, 10))
	harusOK(t, s.SubmitKembali(a, 10))
	list, err := s.ListTransaksi(model.FilterTransaksi{})
	harusOK(t, err)
	if len(list) != 1 {
		t.Fatalf("mau 1 transaksi, dapat %d", len(list))
	}
	id := list[0].ID
	ja, jk := 10, 12
	e := model.EditTransaksi{ID: id, JumlahAmbil: &ja, JumlahKembali: &jk, WaktuAmbil: "09:00", WaktuKembali: "10:00", Status: "SELESAI", Alasan: "salah ketik"}
	if err := s.UbahTransaksi(e); err == nil {
		t.Fatal("kembali > ambil seharusnya ditolak saat edit")
	}
	jk = 7
	e.Alasan = ""
	if err := s.UbahTransaksi(e); err == nil {
		t.Fatal("tanpa alasan seharusnya ditolak")
	}
	e.Alasan = "koreksi"
	harusOK(t, s.UbahTransaksi(e))
	list, _ = s.ListTransaksi(model.FilterTransaksi{})
	if list[0].Keterangan != "Kurang 3 Ompreng" || !list[0].Merah {
		t.Fatalf("keterangan setelah edit: %q merah=%v", list[0].Keterangan, list[0].Merah)
	}

	if err := s.HapusTransaksi(id, " "); err == nil {
		t.Fatal("hapus tanpa alasan seharusnya ditolak")
	}
	harusOK(t, s.HapusTransaksi(id, "dobel"))
	list, _ = s.ListTransaksi(model.FilterTransaksi{})
	if len(list) != 0 {
		t.Fatal("transaksi seharusnya sudah terhapus")
	}
	au, err := s.ListAudit(model.FilterAudit{Entitas: "transaksi"})
	harusOK(t, err)
	if len(au) < 3 || au[0].Aksi != "HAPUS" || au[0].Alasan != "dobel" {
		t.Fatalf("audit tidak sesuai: %+v", au)
	}
}
