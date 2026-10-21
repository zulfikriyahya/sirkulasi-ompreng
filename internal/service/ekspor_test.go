package service

import "testing"

func TestSusunLaporan(t *testing.T) {
	s := setupT5(t)
	a, _ := s.TambahKelas("VII A", "Ani", 1)
	b, _ := s.TambahKelas("VII B", "Budi", 2)
	_, _ = s.TambahKelas("VII C", "Cici", 3)
	harusOK(t, s.SubmitAmbil(a, 10))
	harusOK(t, s.SubmitKembali(a, 8))
	harusOK(t, s.SubmitAmbil(b, 5))
	harusOK(t, s.TutupHari(""))

	l, err := s.susunLaporan(tglUji, tglUji)
	harusOK(t, err)
	if len(l.Baris) != 2 {
		t.Fatalf("mau 2 baris (VII C TIDAK_AMBIL dikecualikan), dapat %d", len(l.Baris))
	}
	if x := l.Baris[0]; x.Keterangan != "Kurang 2 Ompreng" || !x.Merah || x.Tanggal != "12 Januari 2026" || x.WaktuAmbil != "10:00 WIB" {
		t.Fatalf("baris A salah: %+v", x)
	}
	if x := l.Baris[1]; x.Keterangan != "Belum kembali" || !x.Merah || x.JumlahKembali != "-" {
		t.Fatalf("baris B salah: %+v", x)
	}
	if l.Periode != "Hari/Tanggal: Senin, 12 Januari 2026" {
		t.Fatalf("periode: %q", l.Periode)
	}
	if _, err := s.susunLaporan("2026-02-01", "2026-02-02"); err == nil {
		t.Fatal("periode tanpa data seharusnya galat")
	}
	if _, err := s.susunLaporan("2026-02-02", "2026-02-01"); err == nil {
		t.Fatal("rentang terbalik seharusnya galat")
	}
}
