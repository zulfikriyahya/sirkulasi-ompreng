package whatsapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func ttd(secret string, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

func klien(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := New(Config{BaseURL: srv.URL + "/", KeyID: "key_x", Secret: "rahasia"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestKirimTandaTangan(t *testing.T) {
	c := klien(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != "POST" || r.URL.Path != "/api/v1/messages" {
			t.Errorf("rute salah: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "key_x" {
			t.Error("X-API-Key salah")
		}
		if r.Header.Get("X-Signature") != ttd("rahasia", body) {
			t.Error("X-Signature tidak cocok dengan body mentah")
		}
		ts, _ := strconv.ParseInt(r.Header.Get("X-Timestamp"), 10, 64)
		if d := time.Since(time.Unix(ts, 0)); d > time.Minute || d < -time.Minute {
			t.Errorf("timestamp melenceng: %v", d)
		}
		w.Write([]byte(`{"data":{"job_id":"job_1","status":"queued"}}`))
	})
	h, err := c.Kirim(context.Background(), Pesan{TemplateCode: "presensi_masuk", Recipient: "6281234567890",
		Variables: map[string]string{"nama": "Yahya"}, ReferenceID: "A-1"})
	if err != nil || h.JobID != "job_1" || h.Status != "queued" {
		t.Fatalf("hasil=%+v err=%v", h, err)
	}
}

func TestStatusBodyKosong(t *testing.T) {
	c := klien(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/messages/job_9" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Signature") != ttd("rahasia", nil) {
			t.Error("signature GET harus dari body kosong")
		}
		w.Write([]byte(`{"job_id":"job_9","status":"delivered"}`))
	})
	h, err := c.Status(context.Background(), "job_9")
	if err != nil || h.Status != "delivered" {
		t.Fatalf("hasil=%+v err=%v", h, err)
	}
}

func TestGalat(t *testing.T) {
	kasus := []struct {
		kode      int
		is        error
		sementara bool
	}{
		{401, ErrAutentikasi, false},
		{429, ErrDibatasi, false},
		{422, nil, false},
		{503, nil, true},
	}
	for _, k := range kasus {
		c := klien(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(k.kode)
			w.Write([]byte(`{"message":"ditolak"}`))
		})
		_, err := c.Kirim(context.Background(), Pesan{TemplateCode: "t", Recipient: "62811"})
		if err == nil {
			t.Fatalf("%d: seharusnya galat", k.kode)
		}
		if k.is != nil && !errors.Is(err, k.is) {
			t.Errorf("%d: bukan %v", k.kode, k.is)
		}
		if IsSementara(err) != k.sementara {
			t.Errorf("%d: IsSementara salah", k.kode)
		}
	}
}

func TestKonfigurasiTidakValid(t *testing.T) {
	for _, c := range []Config{
		{BaseURL: "", KeyID: "a", Secret: "b"},
		{BaseURL: "ftp://x", KeyID: "a", Secret: "b"},
		{BaseURL: "https://x.id", KeyID: "", Secret: "b"},
		{BaseURL: "https://x.id", KeyID: "a", Secret: ""},
	} {
		if _, err := New(c); err == nil {
			t.Errorf("%+v seharusnya ditolak", c)
		}
	}
}
