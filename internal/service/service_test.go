package service

import (
	"path/filepath"
	"testing"
	"time"

	"mbg/internal/db"
	"mbg/internal/timeutil"
)

func TestAlurTap(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	now := time.Date(2026, 1, 12, 11, 0, 0, 0, timeutil.WIB)
	s := NewWithClock(conn, t.TempDir(), func() time.Time { return now })

	id, err := s.TambahKelas("VII A", "Budi", 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.DaftarkanKartu("12345678", id, "", ""); err != nil {
		t.Fatal(err)
	}

	if r := s.ProsesTap("0012345678"); r.Aksi != "AMBIL" {
		t.Fatalf("tap 1: %+v", r)
	}
	// popup terbuka -> tap lain ditolak
	if r := s.ProsesTap("0099999999"); r.Aksi != "PESAN" {
		t.Fatalf("penguncian popup: %+v", r)
	}
	if err := s.SubmitAmbil(id, 0); err == nil {
		t.Fatal("jumlah 0 harus ditolak")
	}
	if err := s.SubmitAmbil(id, 31); err != nil {
		t.Fatal(err)
	}

	now = now.Add(5 * time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "KONFIRMASI_KEMBALI" {
		t.Fatalf("jeda minimum: %+v", r)
	}
	s.BatalInput()

	now = now.Add(15 * time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "KEMBALI" || r.JumlahAmbil != 31 {
		t.Fatalf("kembali: %+v", r)
	}
	if err := s.SubmitKembali(id, 32); err == nil {
		t.Fatal("kembali > ambil harus ditolak")
	}
	if err := s.SubmitKembali(id, 30); err != nil {
		t.Fatal(err)
	}

	now = now.Add(time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "PESAN" {
		t.Fatalf("sudah selesai: %+v", r)
	}

	st, _ := s.StatusHariIni()
	if !st.SemuaSelesai || st.Baris[0].Keterangan != "Kurang 1 Ompreng" || !st.Baris[0].Merah {
		t.Fatalf("status: %+v", st)
	}
}
