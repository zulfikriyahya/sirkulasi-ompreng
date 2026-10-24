// Package pembaruan memeriksa dan mengunduh rilis dari GitHub dengan validasi ketat
// (HTTPS, host dan path milik repositori, ukuran, dan checksum SHA256).
package pembaruan

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	RepoBawaan           = "zulfikriyahya/sirkulasi-ompreng"
	NamaChecksum         = "checksums.txt"
	NamaInstallerWindows = "mbg-amd64-installer.exe"
)

var ErrTidakAda = errors.New("tidak ada rilis yang cocok")

// ---------- semver ----------

type Semver struct {
	Maj, Min, Pat int
	Pra           string
}

// ParseVersi: awalan "v" diabaikan, sufiks "-stable" dianggap rilis biasa, "+build" dibuang.
func ParseVersi(s string) (Semver, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(strings.TrimPrefix(s, "v"), "V")
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}
	pra := ""
	if i := strings.IndexByte(s, '-'); i >= 0 {
		pra, s = s[i+1:], s[:i]
	}
	if strings.EqualFold(pra, "stable") {
		pra = ""
	}
	parts := strings.Split(s, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return Semver{}, false
	}
	var n [3]int
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 {
			return Semver{}, false
		}
		n[i] = v
	}
	return Semver{n[0], n[1], n[2], pra}, true
}

// Banding: -1 bila a < b, 0 bila sama, 1 bila a > b.
func (a Semver) Banding(b Semver) int {
	for _, p := range [][2]int{{a.Maj, b.Maj}, {a.Min, b.Min}, {a.Pat, b.Pat}} {
		if p[0] < p[1] {
			return -1
		}
		if p[0] > p[1] {
			return 1
		}
	}
	switch {
	case a.Pra == b.Pra:
		return 0
	case a.Pra == "":
		return 1
	case b.Pra == "":
		return -1
	}
	return strings.Compare(a.Pra, b.Pra)
}

func BandingVersi(a, b string) (int, bool) {
	x, ok1 := ParseVersi(a)
	y, ok2 := ParseVersi(b)
	if !ok1 || !ok2 {
		return 0, false
	}
	return x.Banding(y), true
}

// ---------- rilis ----------

type Aset struct {
	Nama   string
	URL    string
	Ukuran int64
}

type Rilis struct {
	Tag        string
	Versi      string
	Catatan    string
	URLHalaman string
	PraRilis   bool
	Aset       []Aset
}

func (r Rilis) CariAset(nama string) *Aset {
	for i := range r.Aset {
		if r.Aset[i].Nama == nama {
			a := r.Aset[i]
			return &a
		}
	}
	return nil
}

// LebihBaru: true hanya bila rilis benar-benar lebih tinggi dari versi terpasang.
// Tidak pernah menawarkan versi sama atau lebih rendah (tanpa penurunan versi), dan
// pre-release diabaikan kecuali pra = true. Versi terpasang yang tak terbaca ("dev") tidak diperbarui.
func LebihBaru(saat string, r Rilis, pra bool) bool {
	s, ok1 := ParseVersi(saat)
	n, ok2 := ParseVersi(r.Tag)
	if !ok1 || !ok2 {
		return false
	}
	if (r.PraRilis || n.Pra != "") && !pra {
		return false
	}
	return n.Banding(s) > 0
}

// PilihAset: aset pemasang sesuai OS. famili: "deb" | "rpm" | "" (hanya untuk linux).
func PilihAset(r Rilis, goos, famili string) (Aset, bool) {
	for _, a := range r.Aset {
		switch goos {
		case "windows":
			if a.Nama == NamaInstallerWindows {
				return a, true
			}
		case "linux":
			if famili == "deb" && strings.HasPrefix(a.Nama, "mbg_") && strings.HasSuffix(a.Nama, ".deb") {
				return a, true
			}
			if famili == "rpm" && strings.HasPrefix(a.Nama, "mbg-") && strings.HasSuffix(a.Nama, ".rpm") {
				return a, true
			}
		}
	}
	return Aset{}, false
}

// FamiliDistro membaca isi /etc/os-release: "deb", "rpm", atau "" bila tidak dikenal.
func FamiliDistro(osRelease string) string {
	var token []string
	for _, ln := range strings.Split(osRelease, "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(ln), "=")
		if !ok || (k != "ID" && k != "ID_LIKE") {
			continue
		}
		token = append(token, strings.Fields(strings.ToLower(strings.Trim(v, `"'`)))...)
	}
	has := func(daftar ...string) bool {
		for _, t := range token {
			for _, d := range daftar {
				if t == d {
					return true
				}
			}
		}
		return false
	}
	switch {
	case has("debian", "ubuntu"):
		return "deb"
	case has("fedora", "rhel", "centos", "rocky", "almalinux"):
		return "rpm"
	}
	return ""
}

// ---------- klien ----------

type Klien struct {
	HTTP      *http.Client
	APIBase   string   // mis. https://api.github.com
	Repo      string   // pemilik/nama
	HostRepo  string   // github.com; path wajib berawalan /<Repo>/
	HostLain  []string // host CDN aset
	UserAgent string
}

func KlienBawaan(versi string) *Klien {
	return &Klien{
		HTTP:      &http.Client{},
		APIBase:   "https://api.github.com",
		Repo:      RepoBawaan,
		HostRepo:  "github.com",
		HostLain:  []string{"objects.githubusercontent.com", "release-assets.githubusercontent.com"},
		UserAgent: "mbg/" + versi,
	}
}

func (k *Klien) izinAPI(u *url.URL) error {
	if u.Scheme != "https" {
		return errors.New("hanya HTTPS yang diizinkan")
	}
	a, err := url.Parse(k.APIBase)
	if err != nil || !strings.EqualFold(u.Hostname(), a.Hostname()) {
		return fmt.Errorf("host API %q tidak diizinkan", u.Hostname())
	}
	return nil
}

// izin: validasi alamat unduhan (dipakai juga untuk setiap pengalihan).
func (k *Klien) izin(u *url.URL) error {
	if u.Scheme != "https" {
		return errors.New("hanya HTTPS yang diizinkan")
	}
	h := strings.ToLower(u.Hostname())
	if h == strings.ToLower(k.HostRepo) {
		p, repo := strings.ToLower(u.Path), strings.ToLower(k.Repo)
		if strings.HasPrefix(p, "/"+repo+"/") || strings.HasPrefix(p, "/repos/"+repo+"/") {
			return nil
		}
		return errors.New("alamat bukan milik repositori rilis")
	}
	for _, x := range k.HostLain {
		if h == strings.ToLower(x) {
			return nil
		}
	}
	return fmt.Errorf("host %q tidak diizinkan", h)
}

func (k *Klien) klien() *http.Client {
	c := *k.HTTP
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("terlalu banyak pengalihan")
		}
		if k.izin(req.URL) != nil && k.izinAPI(req.URL) != nil {
			return fmt.Errorf("pengalihan ke %s ditolak", req.URL.Host)
		}
		return nil
	}
	return &c
}

func (k *Klien) get(ctx context.Context, raw string, api bool) (*http.Response, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if api {
		err = k.izinAPI(u)
	} else {
		err = k.izin(u)
	}
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", k.UserAgent)
	if api {
		req.Header.Set("Accept", "application/vnd.github+json")
	}
	resp, err := k.klien().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("server menjawab status %d", resp.StatusCode)
	}
	return resp, nil
}

type apiRilis struct {
	TagName    string `json:"tag_name"`
	Body       string `json:"body"`
	HTMLURL    string `json:"html_url"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func (a apiRilis) ke() Rilis {
	r := Rilis{Tag: a.TagName, Versi: strings.TrimPrefix(a.TagName, "v"), Catatan: a.Body, URLHalaman: a.HTMLURL, PraRilis: a.Prerelease}
	for _, x := range a.Assets {
		r.Aset = append(r.Aset, Aset{Nama: x.Name, URL: x.URL, Ukuran: x.Size})
	}
	return r
}

// Periksa mengambil rilis terbaru. pra = false memakai releases/latest (tanpa pre-release);
// pra = true memilih versi tertinggi dari 20 rilis terakhir (draft dilewati).
func (k *Klien) Periksa(ctx context.Context, pra bool) (Rilis, error) {
	base := strings.TrimRight(k.APIBase, "/") + "/repos/" + k.Repo + "/releases"
	if !pra {
		resp, err := k.get(ctx, base+"/latest", true)
		if err != nil {
			return Rilis{}, err
		}
		defer resp.Body.Close()
		var a apiRilis
		if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&a); err != nil {
			return Rilis{}, fmt.Errorf("respons tidak terbaca: %w", err)
		}
		if a.TagName == "" {
			return Rilis{}, ErrTidakAda
		}
		return a.ke(), nil
	}
	resp, err := k.get(ctx, base+"?per_page=20", true)
	if err != nil {
		return Rilis{}, err
	}
	defer resp.Body.Close()
	var list []apiRilis
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&list); err != nil {
		return Rilis{}, fmt.Errorf("respons tidak terbaca: %w", err)
	}
	var best Rilis
	var bestV Semver
	found := false
	for _, a := range list {
		if a.Draft {
			continue
		}
		v, ok := ParseVersi(a.TagName)
		if !ok {
			continue
		}
		if !found || v.Banding(bestV) > 0 {
			best, bestV, found = a.ke(), v, true
		}
	}
	if !found {
		return Rilis{}, ErrTidakAda
	}
	return best, nil
}

// ---------- unduh ----------

// ParseChecksums membaca format keluaran sha256sum: "<hash>  <nama>".
func ParseChecksums(teks string) map[string]string {
	out := map[string]string{}
	for _, ln := range strings.Split(teks, "\n") {
		ln = strings.TrimSpace(ln)
		f := strings.Fields(ln)
		if len(f) < 2 || len(f[0]) != 64 {
			continue
		}
		if _, err := hex.DecodeString(f[0]); err != nil {
			continue
		}
		nama := strings.TrimPrefix(strings.TrimSpace(ln[len(f[0]):]), "*")
		out[nama] = strings.ToLower(f[0])
	}
	return out
}

func (k *Klien) ambilChecksum(ctx context.Context, raw string) (map[string]string, error) {
	resp, err := k.get(ctx, raw, false)
	if err != nil {
		return nil, fmt.Errorf("checksums.txt tidak dapat diunduh: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	return ParseChecksums(string(b)), nil
}

type penghitung struct {
	n, total int64
	fn       func(selesai, total int64)
}

func (p *penghitung) Write(b []byte) (int, error) {
	p.n += int64(len(b))
	if p.fn != nil {
		p.fn(p.n, p.total)
	}
	return len(b), nil
}

type Hasil struct {
	Path          string
	Terverifikasi bool // checksum SHA256 dicocokkan
}

// Unduh: checksum (bila ada) diambil lebih dulu. Berkas ditulis ke .part, diverifikasi
// ukuran dan SHA256-nya, baru diganti nama. Gagal verifikasi berarti berkas dihapus.
func (k *Klien) Unduh(ctx context.Context, a Aset, checksum *Aset, dir string, progres func(selesai, total int64)) (Hasil, error) {
	nama := filepath.Base(a.Nama)
	if nama == "." || nama == "" || nama == string(filepath.Separator) {
		return Hasil{}, errors.New("nama berkas tidak valid")
	}
	want := ""
	if checksum != nil {
		daftar, err := k.ambilChecksum(ctx, checksum.URL)
		if err != nil {
			return Hasil{}, err
		}
		h, ok := daftar[nama]
		if !ok {
			return Hasil{}, fmt.Errorf("%s tidak memuat checksum untuk %s", NamaChecksum, nama)
		}
		want = h
	}
	resp, err := k.get(ctx, a.URL, false)
	if err != nil {
		return Hasil{}, err
	}
	defer resp.Body.Close()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Hasil{}, err
	}
	tmp, final := filepath.Join(dir, nama+".part"), filepath.Join(dir, nama)
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return Hasil{}, err
	}
	total := a.Ukuran
	if total <= 0 {
		total = resp.ContentLength
	}
	hsh := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, hsh, &penghitung{total: total, fn: progres}), resp.Body)
	cerr := f.Close()
	if err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return Hasil{}, err
	}
	if a.Ukuran > 0 && n != a.Ukuran {
		os.Remove(tmp)
		return Hasil{}, fmt.Errorf("ukuran berkas tidak cocok (%d dari %d byte)", n, a.Ukuran)
	}
	if want != "" && hex.EncodeToString(hsh.Sum(nil)) != want {
		os.Remove(tmp)
		return Hasil{}, errors.New("checksum SHA256 tidak cocok: berkas ditolak")
	}
	os.Remove(final)
	if err := os.Rename(tmp, final); err != nil {
		os.Remove(tmp)
		return Hasil{}, err
	}
	return Hasil{Path: final, Terverifikasi: want != ""}, nil
}
