package pembaruan

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"testing"
)

func sha(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

type palsu struct {
	srv *httptest.Server
	mux *http.ServeMux
	ua  atomic.Value
}

func newPalsu(t *testing.T) (*palsu, *Klien) {
	t.Helper()
	f := &palsu{mux: http.NewServeMux()}
	f.srv = httptest.NewTLSServer(f.mux)
	t.Cleanup(f.srv.Close)
	k := &Klien{HTTP: f.srv.Client(), APIBase: f.srv.URL, Repo: RepoBawaan, HostRepo: "127.0.0.1", UserAgent: "mbg/uji"}
	return f, k
}

func (f *palsu) rel(tag string, pre, draft bool, nama ...string) map[string]any {
	var as []map[string]any
	for _, n := range nama {
		as = append(as, map[string]any{"name": n, "size": 10, "browser_download_url": f.srv.URL + "/" + RepoBawaan + "/releases/download/" + tag + "/" + n})
	}
	return map[string]any{"tag_name": tag, "body": "catatan", "html_url": "https://github.com/x", "prerelease": pre, "draft": draft, "assets": as}
}

func TestBandingVersi(t *testing.T) {
	urut := []string{"0.1.1", "0.1.2", "0.2.0", "1.0.0"}
	for i := 0; i < len(urut)-1; i++ {
		if c, ok := BandingVersi(urut[i], urut[i+1]); !ok || c != -1 {
			t.Fatalf("%s < %s: %d %v", urut[i], urut[i+1], c, ok)
		}
		if c, _ := BandingVersi(urut[i+1], urut[i]); c != 1 {
			t.Fatalf("%s > %s", urut[i+1], urut[i])
		}
	}
	for _, p := range [][2]string{{"v0.1.1", "0.1.1"}, {"0.1.1-stable", "v0.1.1"}, {"V1.0", "1.0.0"}} {
		if c, ok := BandingVersi(p[0], p[1]); !ok || c != 0 {
			t.Fatalf("%v harus sama: %d %v", p, c, ok)
		}
	}
	if c, _ := BandingVersi("0.3.0-rc1", "0.3.0"); c != -1 {
		t.Fatal("pra-rilis harus lebih rendah dari rilis")
	}
	if _, ok := BandingVersi("dev", "0.1.1"); ok {
		t.Fatal("dev tidak boleh terbaca sebagai versi")
	}
}

func TestLebihBaru(t *testing.T) {
	r := func(tag string, pre bool) Rilis { return Rilis{Tag: tag, PraRilis: pre} }
	cases := []struct {
		saat string
		r    Rilis
		pra  bool
		want bool
	}{
		{"0.1.1", r("v0.2.0", false), false, true},
		{"0.2.0", r("v0.2.0", false), false, false}, // sama
		{"0.3.0", r("v0.2.0", false), false, false}, // tanpa penurunan versi
		{"0.1.1", r("v0.2.0", true), false, false},  // pra-rilis diabaikan
		{"0.1.1", r("v0.2.0-rc1", false), false, false},
		{"0.1.1", r("v0.2.0", true), true, true},
		{"dev", r("v0.2.0", false), false, false},
	}
	for i, c := range cases {
		if got := LebihBaru(c.saat, c.r, c.pra); got != c.want {
			t.Fatalf("kasus %d: %v", i, got)
		}
	}
}

func TestFamiliDistroDanPilihAset(t *testing.T) {
	if FamiliDistro("ID=ubuntu\nID_LIKE=debian\n") != "deb" || FamiliDistro(`ID="linuxmint"`+"\n"+`ID_LIKE="ubuntu debian"`) != "deb" {
		t.Fatal("deb")
	}
	if FamiliDistro("ID=fedora\n") != "rpm" || FamiliDistro(`ID="rocky"`+"\n"+`ID_LIKE="rhel centos fedora"`) != "rpm" {
		t.Fatal("rpm")
	}
	if FamiliDistro("ID=arch\n") != "" {
		t.Fatal("tidak dikenal")
	}
	r := Rilis{Aset: []Aset{{Nama: "mbg-linux-amd64"}, {Nama: "mbg_0.2.0_amd64.deb"}, {Nama: "mbg-0.2.0-1.x86_64.rpm"}, {Nama: NamaInstallerWindows}, {Nama: NamaChecksum}}}
	if a, ok := PilihAset(r, "windows", ""); !ok || a.Nama != NamaInstallerWindows {
		t.Fatalf("windows: %+v", a)
	}
	if a, ok := PilihAset(r, "linux", "deb"); !ok || a.Nama != "mbg_0.2.0_amd64.deb" {
		t.Fatalf("deb: %+v", a)
	}
	if a, ok := PilihAset(r, "linux", "rpm"); !ok || a.Nama != "mbg-0.2.0-1.x86_64.rpm" {
		t.Fatalf("rpm: %+v", a)
	}
	if _, ok := PilihAset(r, "linux", ""); ok {
		t.Fatal("distro tidak dikenal tidak boleh memilih aset")
	}
}

func TestParseChecksums(t *testing.T) {
	h := strings.Repeat("a", 64)
	m := ParseChecksums(h + "  mbg_0.2.0_amd64.deb\n" + strings.ToUpper(h) + " *mbg-amd64-installer.exe\nrusak  x\n")
	if m["mbg_0.2.0_amd64.deb"] != h || m["mbg-amd64-installer.exe"] != h || len(m) != 2 {
		t.Fatalf("%v", m)
	}
}

func TestValidasiURLBawaan(t *testing.T) {
	k := KlienBawaan("x")
	ok := []string{
		"https://github.com/zulfikriyahya/sirkulasi-ompreng/releases/download/v1/a",
		"https://objects.githubusercontent.com/x/y",
		"https://release-assets.githubusercontent.com/github-production-release-asset/1?sig=a",
	}
	tolak := []string{
		"http://github.com/zulfikriyahya/sirkulasi-ompreng/releases/download/v1/a",
		"https://github.com/orang-lain/repo/releases/download/v1/a",
		"https://githubusercontent.com.evil.com/a",
		"https://evil.example/a",
		"ftp://github.com/zulfikriyahya/sirkulasi-ompreng/x",
	}
	for _, s := range ok {
		u, _ := url.Parse(s)
		if err := k.izin(u); err != nil {
			t.Fatalf("%s harus diizinkan: %v", s, err)
		}
	}
	for _, s := range tolak {
		u, _ := url.Parse(s)
		if err := k.izin(u); err == nil {
			t.Fatalf("%s harus ditolak", s)
		}
	}
}

func TestPeriksaLatestDanUserAgent(t *testing.T) {
	f, k := newPalsu(t)
	f.mux.HandleFunc("/repos/"+RepoBawaan+"/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		f.ua.Store(r.Header.Get("User-Agent"))
		json.NewEncoder(w).Encode(f.rel("v0.2.0", false, false, "mbg-amd64-installer.exe", NamaChecksum))
	})
	r, err := k.Periksa(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Versi != "0.2.0" || len(r.Aset) != 2 || r.Catatan != "catatan" || f.ua.Load() != "mbg/uji" {
		t.Fatalf("%+v ua=%v", r, f.ua.Load())
	}
	if !LebihBaru("0.1.1", r, false) || LebihBaru("0.2.0", r, false) || LebihBaru("0.3.0", r, false) {
		t.Fatal("LebihBaru salah")
	}
}

func TestPeriksaPraRilis(t *testing.T) {
	f, k := newPalsu(t)
	f.mux.HandleFunc("/repos/"+RepoBawaan+"/releases", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			f.rel("v9.9.9", false, true, "x"), // draft dilewati
			f.rel("v0.2.0", false, false, "x"),
			f.rel("v0.3.0-rc1", true, false, "x"),
		})
	})
	r, err := k.Periksa(context.Background(), true)
	if err != nil || r.Tag != "v0.3.0-rc1" {
		t.Fatalf("%+v %v", r, err)
	}
	if LebihBaru("0.2.0", r, false) || !LebihBaru("0.2.0", r, true) {
		t.Fatal("pra-rilis hanya ditawarkan bila opsinya aktif")
	}
}

func TestPeriksaServerGalat(t *testing.T) {
	f, k := newPalsu(t)
	f.mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "x", 500) })
	if _, err := k.Periksa(context.Background(), false); err == nil {
		t.Fatal("status 500 harus galat")
	}
}

func jalankanUnduh(t *testing.T, csBody string, ukuran int64) (Hasil, string, error) {
	t.Helper()
	isi := []byte("isi installer palsu")
	f, k := newPalsu(t)
	pref := "/" + RepoBawaan + "/releases/download/v0.2.0/"
	f.mux.HandleFunc(pref, func(w http.ResponseWriter, r *http.Request) {
		switch path.Base(r.URL.Path) {
		case NamaInstallerWindows:
			w.Write(isi)
		case NamaChecksum:
			fmt.Fprint(w, strings.ReplaceAll(csBody, "{HASH}", sha(isi)))
		default:
			http.NotFound(w, r)
		}
	})
	if ukuran == 0 {
		ukuran = int64(len(isi))
	}
	a := Aset{Nama: NamaInstallerWindows, URL: f.srv.URL + pref + NamaInstallerWindows, Ukuran: ukuran}
	var csAset *Aset
	if csBody != "" {
		csAset = &Aset{Nama: NamaChecksum, URL: f.srv.URL + pref + NamaChecksum}
	}
	dir := t.TempDir()
	h, err := k.Unduh(context.Background(), a, csAset, dir, nil)
	return h, dir, err
}

func sisaBerkas(t *testing.T, dir string) int {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	return len(ents)
}

func TestUnduhChecksum(t *testing.T) {
	h, dir, err := jalankanUnduh(t, "{HASH}  "+NamaInstallerWindows+"\n", 0)
	if err != nil || !h.Terverifikasi {
		t.Fatalf("cocok: %+v %v", h, err)
	}
	if b, _ := os.ReadFile(h.Path); string(b) != "isi installer palsu" {
		t.Fatal("isi berkas salah")
	}
	_ = dir

	_, dir, err = jalankanUnduh(t, strings.Repeat("0", 64)+"  "+NamaInstallerWindows+"\n", 0)
	if err == nil || !strings.Contains(err.Error(), "checksum") || sisaBerkas(t, dir) != 0 {
		t.Fatalf("checksum salah harus ditolak dan berkas dihapus: %v sisa=%d", err, sisaBerkas(t, dir))
	}

	if _, _, err = jalankanUnduh(t, "{HASH}  lain.bin\n", 0); err == nil {
		t.Fatal("nama tidak ada di checksums.txt harus ditolak")
	}

	_, dir, err = jalankanUnduh(t, "{HASH}  "+NamaInstallerWindows+"\n", 999)
	if err == nil || !strings.Contains(err.Error(), "ukuran") || sisaBerkas(t, dir) != 0 {
		t.Fatalf("ukuran salah: %v", err)
	}
}

func TestUnduhTanpaChecksum(t *testing.T) {
	h, _, err := jalankanUnduh(t, "", 0)
	if err != nil || h.Terverifikasi {
		t.Fatalf("%+v %v", h, err)
	}
}

func TestUnduhTolakHostDanPengalihan(t *testing.T) {
	f, k := newPalsu(t)
	for _, u := range []string{
		"http://127.0.0.1/x",
		"https://evil.example/x",
		"https://127.0.0.1/orang-lain/repo/releases/download/v1/x",
	} {
		if _, err := k.Unduh(context.Background(), Aset{Nama: "x", URL: u}, nil, t.TempDir(), nil); err == nil {
			t.Fatalf("%s harus ditolak", u)
		}
	}
	pref := "/" + RepoBawaan + "/releases/download/v1/"
	f.mux.HandleFunc(pref, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://evil.example/x", http.StatusFound)
	})
	if _, err := k.Unduh(context.Background(), Aset{Nama: "x", URL: f.srv.URL + pref + "x"}, nil, t.TempDir(), nil); err == nil {
		t.Fatal("pengalihan ke host lain harus ditolak")
	}
}
