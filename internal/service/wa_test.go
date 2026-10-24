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

func TestUjiDanUlangiWA(t *testing.T) {
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

	if _, err := s.UjiKirimWA("abc"); err == nil {
		t.Error("nomor tidak valid seharusnya ditolak")
	}
	u, err := s.UjiKirimWA("0812-3456-7890")
	if err != nil || u.JobID != "job_1" {
		t.Fatalf("uji = %+v, %v", u, err)
	}
	if len(diterima) != 1 || diterima[0]["template_code"] != "ompreng_ambil" {
		t.Fatalf("pesan uji salah: %v", diterima)
	}

	_, err = conn.Exec(`INSERT INTO wa_antrean (jenis,reference_id,no_wa,variabel,status,galat,dibuat)
		VALUES ('AMBIL','x-1','6281234567890','{}','GAGAL','ditolak',?)`, waf(s.now()))
	if err != nil {
		t.Fatal(err)
	}
	r, _ := s.RingkasanWA()
	if r.Gagal != 1 {
		t.Fatalf("ringkasan = %+v", r)
	}
	if err := s.UlangiWA(9999); err == nil {
		t.Error("ulangi id tak dikenal seharusnya galat")
	}
	var id int64
	_ = conn.QueryRow(`SELECT id FROM wa_antrean WHERE reference_id='x-1'`).Scan(&id)
	if err := s.UlangiWA(id); err != nil {
		t.Fatal(err)
	}
	if err := s.UlangiWA(id); err == nil {
		t.Error("mengulang pesan yang sudah ANTRE seharusnya galat")
	}
	s.prosesAntreanWA(context.Background())
	l, err := s.DaftarAntreanWA("TERKIRIM", 10)
	if err != nil || len(l) != 1 {
		t.Fatalf("daftar = %+v, %v", l, err)
	}
}
