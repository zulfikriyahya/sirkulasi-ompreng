package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"mbg/internal/db"
	"mbg/internal/model"
	"mbg/internal/pembaruan"
	"mbg/internal/versi"
)

type ghPalsu struct {
	srv *httptest.Server
	hit atomic.Int32
}

// newGHPalsu: cs "" = tanpa checksums.txt, "auto" = hitung dari berkas, selain itu isi mentah.
func newGHPalsu(t *testing.T, tag string, berkas map[string]string, cs string) *ghPalsu {
	t.Helper()
	g := &ghPalsu{}
	mux := http.NewServeMux()
	g.srv = httptest.NewTLSServer(mux)
	t.Cleanup(g.srv.Close)
	pref := "/" + pembaruan.RepoBawaan + "/releases/download/" + tag + "/"

	isi := map[string]string{}
	for n, c := range berkas {
		isi[n] = c
	}
	if cs == "auto" {
		var sb strings.Builder
		for n, c := range berkas {
			h := sha256.Sum256([]byte(c))
			fmt.Fprintf(&sb, "%s  %s\n", hex.EncodeToString(h[:]), n)
		}
		cs = sb.String()
	}
	if cs != "" {
		isi[pembaruan.NamaChecksum] = cs
	}
	var aset []map[string]any
	for n, c := range isi {
		aset = append(aset, map[string]any{"name": n, "size": len(c), "browser_download_url": g.srv.URL + pref + n})
	}
	mux.HandleFunc("/repos/"+pembaruan.RepoBawaan+"/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		g.hit.Add(1)
		json.NewEncoder(w).Encode(map[string]any{"tag_name": tag, "body": "Catatan uji", "html_url": "https://github.com/x",
			"draft": false, "prerelease": false, "assets": aset})
	})
	mux.HandleFunc(pref, func(w http.ResponseWriter, r *http.Request) {
		c, ok := isi[strings.TrimPrefix(r.URL.Path, pref)]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(c))
	})
	return g
}

func (g *ghPalsu) pasang(s *Service) {
	s.pemb.klien = &pembaruan.Klien{HTTP: g.srv.Client(), APIBase: g.srv.URL, Repo: pembaruan.RepoBawaan,
		HostRepo: "127.0.0.1", UserAgent: "mbg/uji"}
}

func setVersi(t *testing.T, v string) {
	t.Helper()
	lama := versi.Versi
	versi.Versi = v
	t.Cleanup(func() { versi.Versi = lama })
}

var asetUji = map[string]string{
	"mbg_0.2.0_amd64.deb":     "deb palsu",
	"mbg-0.2.0-1.x86_64.rpm":  "rpm palsu",
	"mbg-amd64-installer.exe": "exe palsu",
}

func TestInfoSistemMemakaiPaketVersi(t *testing.T) {
	setVersi(t, "9.9.9")
	if got := setupT5(t).InfoSistem().Versi; got != "9.9.9" {
		t.Fatalf("versi: %q", got)
	}
}

func TestPembaruanOfflineTidakGalatDiModeOtomatis(t *testing.T) {
	setVersi(t, "0.1.1")
	s := setupT5(t)
	g := newGHPalsu(t, "v0.2.0", asetUji, "auto")
	g.pasang(s)
	g.srv.Close() // server mati = offline

	info, err := s.PeriksaPembaruan(true)
	if err != nil || info.Ada {
		t.Fatalf("otomatis offline: %+v %v", info, err)
	}
	if _, err := s.PeriksaPembaruan(false); err == nil {
		t.Fatal("manual offline harus memberi pesan galat")
	}
}

func TestPembaruanOpsiOtomatisMati(t *testing.T) {
	setVersi(t, "0.1.1")
	s := setupT5(t)
	g := newGHPalsu(t, "v0.2.0", asetUji, "auto")
	g.pasang(s)

	harusOK(t, s.SimpanOpsiPembaruan(model.OpsiPembaruan{Otomatis: false}))
	if info, err := s.PeriksaPembaruan(true); err != nil || info.Ada || g.hit.Load() != 0 {
		t.Fatalf("otomatis mati tidak boleh memanggil jaringan: %+v %v hit=%d", info, err, g.hit.Load())
	}
	info, err := s.PeriksaPembaruan(false)
	if err != nil || !info.Ada || g.hit.Load() != 1 {
		t.Fatalf("manual tetap bekerja: %+v %v hit=%d", info, err, g.hit.Load())
	}
	o, _ := s.GetOpsiPembaruan()
	if o.Otomatis || o.PraRilis {
		t.Fatalf("opsi: %+v", o)
	}
}

func TestPembaruanTidakMenurunkanVersi(t *testing.T) {
	setVersi(t, "0.1.1")
	for _, tag := range []string{"v0.1.0", "v0.1.1"} {
		s := setupT5(t)
		g := newGHPalsu(t, tag, asetUji, "auto")
		g.pasang(s)
		if info, err := s.PeriksaPembaruan(false); err != nil || info.Ada {
			t.Fatalf("%s tidak boleh ditawarkan: %+v %v", tag, info, err)
		}
	}
}

func TestPasangLinuxDeb(t *testing.T) {
	setVersi(t, "0.1.1")
	s := setupT5(t)
	_, err := s.TambahKelas("VII A", "Ani", 1)
	harusOK(t, err)
	g := newGHPalsu(t, "v0.2.0", asetUji, "auto")
	g.pasang(s)
	s.pemb.goos = "linux"
	s.pemb.osRelease = func() string { return "ID=ubuntu\nID_LIKE=debian\n" }
	unduh := t.TempDir()
	s.pemb.dirUnduh = unduh

	info, err := s.PeriksaPembaruan(false)
	harusOK(t, err)
	if !info.Ada || info.Cara != "LINUX_DEB" || !info.ChecksumAda || info.VersiBaru != "0.2.0" || info.NamaAset != "mbg_0.2.0_amd64.deb" {
		t.Fatalf("info: %+v", info)
	}
	h, err := s.PasangPembaruan()
	harusOK(t, err)
	if h.Keluar || !strings.HasPrefix(h.Perintah, "sudo apt install ") || !strings.Contains(h.Perintah, "mbg_0.2.0_amd64.deb") {
		t.Fatalf("hasil: %+v", h)
	}
	if b, _ := os.ReadFile(filepath.Join(unduh, "mbg_0.2.0_amd64.deb")); string(b) != "deb palsu" {
		t.Fatal("berkas unduhan salah")
	}

	// cadangan dapat dibuka dan datanya sama
	cad, err := db.Open(h.Cadangan)
	harusOK(t, err)
	defer cad.Close()
	var n int
	harusOK(t, cad.QueryRow(`SELECT COUNT(*) FROM kelas`).Scan(&n))
	if n != 1 {
		t.Fatalf("cadangan: %d kelas", n)
	}
	au, err := s.ListAudit(model.FilterAudit{Entitas: "aplikasi"})
	harusOK(t, err)
	if len(au) != 1 || au[0].Aksi != "PEMBARUAN" {
		t.Fatalf("audit: %+v", au)
	}
}

func TestPasangChecksumSalahDitolak(t *testing.T) {
	setVersi(t, "0.1.1")
	s := setupT5(t)
	g := newGHPalsu(t, "v0.2.0", asetUji, strings.Repeat("0", 64)+"  mbg_0.2.0_amd64.deb\n")
	g.pasang(s)
	s.pemb.goos = "linux"
	s.pemb.osRelease = func() string { return "ID=debian\n" }
	unduh := t.TempDir()
	s.pemb.dirUnduh = unduh

	_, err := s.PeriksaPembaruan(false)
	harusOK(t, err)
	if _, err := s.PasangPembaruan(); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("checksum salah harus ditolak: %v", err)
	}
	if _, err := os.Stat(filepath.Join(unduh, "mbg_0.2.0_amd64.deb")); err == nil {
		t.Fatal("berkas yang gagal verifikasi harus dihapus")
	}
}

func TestPasangTanpaChecksumBolehDenganPeringatan(t *testing.T) {
	setVersi(t, "0.1.1")
	s := setupT5(t)
	g := newGHPalsu(t, "v0.2.0", asetUji, "")
	g.pasang(s)
	s.pemb.goos = "linux"
	s.pemb.osRelease = func() string { return "ID=fedora\n" }
	s.pemb.dirUnduh = t.TempDir()

	info, err := s.PeriksaPembaruan(false)
	harusOK(t, err)
	if info.ChecksumAda || info.Cara != "LINUX_RPM" {
		t.Fatalf("info: %+v", info)
	}
	h, err := s.PasangPembaruan()
	harusOK(t, err)
	if !strings.HasPrefix(h.Perintah, "sudo dnf install ") || !strings.Contains(h.Pesan, "tanpa checksums.txt") {
		t.Fatalf("hasil: %+v", h)
	}
}

func TestPasangWindows(t *testing.T) {
	setVersi(t, "0.1.1")
	s := setupT5(t)
	g := newGHPalsu(t, "v0.2.0", asetUji, "auto")
	g.pasang(s)
	s.pemb.goos = "windows"
	var dijalankan string
	s.pemb.luncur = func(p string) error { dijalankan = p; return nil }

	info, err := s.PeriksaPembaruan(false)
	harusOK(t, err)
	if info.Cara != "WINDOWS" {
		t.Fatalf("info: %+v", info)
	}
	h, err := s.PasangPembaruan()
	harusOK(t, err)
	if dijalankan != "" {
		t.Cleanup(func() { os.RemoveAll(filepath.Dir(dijalankan)) })
	}
	if !h.Keluar || !strings.HasSuffix(dijalankan, "mbg-amd64-installer.exe") || h.Perintah != "" {
		t.Fatalf("hasil: %+v dijalankan=%q", h, dijalankan)
	}
}

func TestCadanganDipangkasLimaTerakhir(t *testing.T) {
	s := setupT5(t)
	dir := filepath.Join(s.dir, "cadangan")
	harusOK(t, os.MkdirAll(dir, 0o700))
	for d := 1; d <= 7; d++ {
		harusOK(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("mbg-2025010%d-000000.db", d)), []byte("x"), 0o600))
	}
	baru, err := s.CadangkanDB()
	harusOK(t, err)
	ents, _ := os.ReadDir(dir)
	if len(ents) != 5 {
		t.Fatalf("mau 5 cadangan, dapat %d", len(ents))
	}
	if _, err := os.Stat(baru); err != nil {
		t.Fatal("cadangan baru harus tersisa")
	}
	if _, err := os.Stat(filepath.Join(dir, "mbg-20250101-000000.db")); err == nil {
		t.Fatal("cadangan terlama harus dihapus")
	}
}
