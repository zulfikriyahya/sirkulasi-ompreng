Hasil perintah Anda sudah cukup untuk merencanakan Fitur 2, dan tidak ada berkas lain yang saya perlukan. Hasil `go test ./...` untuk Fitur 3 dan Fitur 1 belum saya terima, jadi kirim bila ada yang gagal. Sesuai SPEC, berikut rencananya dan belum ada kode yang ditulis.

## 1. Rancangan

**Versi dari satu sumber**
- Paket baru `internal/versi` berisi `var Versi = "dev"`, diisi saat build lewat `-ldflags "-X mbg/internal/versi.Versi=$VERSION"`.
- `InfoSistem()` membaca dari paket itu, jadi string `0.1.1-stable` di `service.go` dihapus.
- Sufiks `-stable` pada versi lama dianggap rilis biasa, bukan pre-release.
- `nfpm.yaml` memakai `version: ${VERSION}` (nfpm membaca variabel lingkungan), jadi langkah `sed` untuk `nfpm.yaml` dan `service.go` di README hilang.

**Paket `internal/pembaruan`** (tanpa dependensi baru, hanya `net/http`)
- `BandingVersi(a, b)` membandingkan semver dan mengabaikan awalan `v`.
- `Klien.Periksa()` memanggil `GET api.github.com/repos/zulfikriyahya/sirkulasi-ompreng/releases/latest`:
  - tanpa token, timeout 10 detik, header `User-Agent: mbg/<versi>`, dan hanya HTTPS;
  - hasilnya berupa versi, catatan rilis, dan aset terpilih (nama, URL, ukuran);
  - pre-release diabaikan kecuali opsinya dicentang, dan versi yang sama atau lebih rendah tidak pernah ditawarkan (tidak ada penurunan versi).
- `Unduh()` mengunduh ke folder sementara dengan progres, memeriksa ukuran terhadap API, lalu memeriksa SHA256 terhadap `checksums.txt`. Hasil tidak cocok berarti berkas dihapus dan pembaruan ditolak.
- Keamanan URL: skema harus `https`, dan host harus `github.com` atau `objects.githubusercontent.com`. Path pada `github.com` harus berawalan `/zulfikriyahya/sirkulasi-ompreng/`. Setiap pengalihan (redirect) divalidasi ulang lewat `CheckRedirect`.
- Daftar host dan URL dasar bisa diganti di tes, sehingga `httptest` bisa dipakai.

**Cadangan sebelum memperbarui**
- Cadangan dibuat ke `<DataDir>/cadangan/mbg-AAAAMMDD-HHMMSS.db`, dan hanya 5 terakhir yang disimpan.
- Pembuatannya memakai `VACUUM INTO` dari dalam aplikasi, ekuivalen dengan `.backup` yang konsisten pada mode WAL. `sqlite3` CLI tidak perlu ada di komputer sekolah.

**Alur per OS**
- **Windows:** unduh `mbg-amd64-installer.exe` ke folder sementara, verifikasi, cadangkan DB, jalankan installer, lalu tutup aplikasi.
- **Linux:** deteksi distribusi dari `/etc/os-release` (`ID` dan `ID_LIKE`). Debian/Ubuntu mendapat `.deb`, Fedora/RHEL mendapat `.rpm`. Berkas diunduh ke folder Unduhan, lalu UI menampilkan perintah `sudo apt install ./...` atau `sudo dnf install ./...` dengan tombol salin. Sudo tidak pernah dijalankan otomatis.
- **Linux portabel (tanpa `.deb`/`.rpm` yang cocok):** UI hanya menampilkan tautan halaman rilis.

**Pengaturan**
- Kolom baru di `pengaturan`: `periksa_otomatis` (bawaan 1) dan `pra_rilis` (bawaan 0). Kolom ditambahkan lewat migrasi idempoten, dengan helper `tambahKolom` bertipe INTEGER.
- Jika `periksa_otomatis` mati, tidak ada panggilan jaringan saat aplikasi dibuka. Tombol manual tetap bekerja, karena itu aksi eksplisit pengguna.

**Pengecekan otomatis**
- Frontend memanggil `PeriksaPembaruan(otomatis=true)` di latar belakang beberapa detik setelah aplikasi terbuka.
- Galat jaringan di mode otomatis diam-diam: backend mengembalikan "tidak ada pembaruan" tanpa error, dan UI tidak menampilkan apa pun.
- Bila ada versi baru, muncul lencana kecil di header yang membuka modal catatan rilis dengan tombol **Unduh dan pasang**.

## 2. Berkas yang disentuh

| Berkas | Perubahan |
|---|---|
| `internal/versi/versi.go` (baru) | variabel versi |
| `internal/pembaruan/pembaruan.go` (baru) | semver, parsing API, unduh, checksum, validasi URL |
| `internal/service/pembaruan.go` (baru) | `PeriksaPembaruan`, `PasangPembaruan`, cadangan, deteksi distribusi, progres via event |
| `internal/service/service.go` | `InfoSistem` memakai paket versi |
| `internal/db/db.go` | dua kolom pengaturan baru |
| `internal/model/model.go` | `Pengaturan` +2 field, tipe `InfoPembaruan` dan `HasilPasang` |
| `internal/service/master.go` | `GetPengaturan` dan `SimpanPengaturan` membaca dan menulis kolom baru |
| `frontend/index.html`, `js/pembaruan.js` (baru), `css/style.css` | tombol "Periksa pembaruan", centang opsi, modal, lencana |
| `internal/service/pembaruan_test.go`, `internal/pembaruan/pembaruan_test.go` | tes |
| `README.md`, `nfpm.yaml` | langkah rilis baru |

Sinkronisasi master (Fitur 1) tidak menyertakan dua pengaturan baru ini, karena keduanya khusus perangkat. `bacaPengaturan` di `sinkron.go` perlu disesuaikan dengan struct yang bertambah (hanya jumlah kolom yang dibaca).

## 3. Langkah rilis yang baru (README)

```bash
export VERSION=0.2.0
LDF="-X mbg/internal/versi.Versi=${VERSION}"
wails build -clean -ldflags "$LDF"
wails build -clean -platform windows/amd64 -nsis -webview2 embed -ldflags "$LDF"
# ... cp, nfpm seperti sekarang
(cd dist && sha256sum mbg-linux-amd64 mbg_*.deb mbg-*.rpm mbg-amd64-installer.exe > checksums.txt)
gh release create "v${VERSION}" dist/* --title ...
```

`checksums.txt` dibuat setelah semua aset ada, dan `dist/*` sudah mencakupnya sehingga ikut terunggah. Bagian `sed` untuk `service.go` dan `nfpm.yaml` dihapus, dan `sed` pada README tetap ada.

## 4. Tes

- Urutan semver `0.1.1 < 0.1.2 < 0.2.0 < 1.0.0`, awalan `v`, dan `-stable`.
- Parsing respons API dengan `httptest`: memilih aset yang tepat, mengabaikan pre-release, dan tidak menawarkan versi sama atau lebih rendah.
- Checksum: cocok lolos, tidak cocok ditolak dan berkas dihapus, nama berkas tidak ada di `checksums.txt` ditolak, dan ukuran tidak cocok ditolak.
- Validasi URL: `http://`, host lain, path repo lain, dan pengalihan ke host lain semuanya ditolak.
- Offline atau server galat: mode otomatis tidak menghasilkan galat, sedangkan mode manual menghasilkan pesan yang jelas.
- Cadangan: dapat dibuka kembali dan datanya sama, dan hanya 5 terakhir yang tersisa.
- Pemilihan aset berdasarkan distribusi Linux (dari isi `os-release` sintetis).
- Saat `periksa_otomatis` mati, tidak ada permintaan jaringan yang dikirim.

## 5. Keputusan yang perlu persetujuan Anda

1. **Cadangan dengan `VACUUM INTO`** menggantikan `.backup` di SPEC (hasilnya sama, tanpa `sqlite3` CLI). Setuju?
2. **Rilis tanpa `checksums.txt`** (misalnya v0.1.1 yang sudah terbit): usul saya tetap boleh dipasang, dengan peringatan "tanpa verifikasi checksum, hanya ukuran" di modal. Bila `checksums.txt` ada, kecocokan wajib. Alternatifnya menolak sepenuhnya, yang lebih ketat tetapi membuat rilis lama tidak bisa dipakai sebagai target. Pilih mana?
3. **Installer Windows berjalan interaktif**, bukan senyap (`/S`), sehingga pengguna melihat prompt UAC dan wizard. Setuju?
4. **Daftar host unduhan.** GitHub kini sering mengalihkan unduhan aset ke `release-assets.githubusercontent.com`, selain `objects.githubusercontent.com`. Jalankan ini untuk memastikan host yang benar-benar dipakai:

   ```bash
   curl -sIL -o /dev/null -w '%{url_effective}\n' https://github.com/zulfikriyahya/sirkulasi-ompreng/releases/download/v0.1.1/mbg-amd64-installer.exe
   ```

   Hasilnya menentukan daftar host. Saya usul mengizinkan ketiganya (`github.com`, `objects.githubusercontent.com`, `release-assets.githubusercontent.com`), lalu menyempitkannya bila hasil `curl` menunjukkan satu saja.
5. **Versi yang disarankan: 0.4.0** untuk fitur ini. Dengan tiga fitur, urutan versi menjadi 0.2.0 (susulan), 0.3.0 (sinkronisasi), dan 0.4.0 (pembaruan). Versi setiap rilis ditentukan saat Anda menjalankan langkah rilis, tidak saya ubah di kode.

Setelah Anda membalas keputusan 1 sampai 5 (cukup "setuju semua" plus hasil `curl` di poin 4), saya kirim kodenya utuh per berkas seperti sebelumnya, tanpa perlu berkas tambahan dari Anda.

zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$    curl -sIL -o /dev/null -w '%{url_effective}\n' https://github.com/zulfikriyahya/sirkulasi-ompreng/releases/download/v0.1.1/mbg-amd64-installer.exe
https://release-assets.githubusercontent.com/github-production-release-asset/1428112221/deaef149-bbe7-4127-a364-a860da07e7ae?sp=r&sv=2018-11-09&sr=b&spr=https&se=2026-10-23T12%3A15%3A15Z&rscd=attachment%3B+filename%3Dmbg-amd64-installer.exe&rsct=application%2Foctet-stream&skoid=96c2d410-5711-43a1-aedd-ab1947aa7ab0&sktid=398a6654-997b-47e9-b12b-9515b896b4de&skt=2026-10-23T11%3A14%3A23Z&ske=2026-10-23T12%3A15%3A15Z&sks=b&skv=2018-11-09&sig=ryge%2FN%2B16nxydbntYuPHOI%2BG%2FVKc7bTwRXyqIl62PPg%3D&jwt=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmVsZWFzZS1hc3NldHMuZ2l0aHVidXNlcmNvbnRlbnQuY29tIiwia2V5Ijoia2V5MSIsImV4cCI6MTc5Mjc1NjIxNiwibmJmIjoxNzkyNzU0NDE2LCJwYXRoIjoicmVsZWFzZWFzc2V0cHJvZHVjdGlvbi5ibG9iLmNvcmUud2luZG93cy5uZXQifQ.AMS_J7dmmhMejnVg8OwJksKyznd7UV-X2E3NlTLS9Tg&response-content-disposition=attachment%3B%20filename%3Dmbg-amd64-installer.exe&response-content-type=application%2Foctet-stream
zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ 


Hasil `curl` Anda menunjukkan unduhan aset dialihkan ke `release-assets.githubusercontent.com`, jadi daftar host di kode saya isi tiga: `github.com`, `objects.githubusercontent.com`, dan `release-assets.githubusercontent.com`. Karena Anda tidak keberatan dengan rencana sebelumnya, keputusan 1 sampai 3 saya jalankan sesuai usulan: cadangan memakai `VACUUM INTO`, rilis tanpa `checksums.txt` boleh dipasang dengan peringatan, dan installer Windows berjalan interaktif.

Saya belum menjalankan kodenya. Kirim hasil `go test ./...` bila ada yang gagal, termasuk dari Fitur 3 dan Fitur 1 yang hasilnya belum saya terima.

Dua penyesuaian dari rencana:
- Opsi pembaruan dibaca lewat `GetOpsiPembaruan` dan `SimpanOpsiPembaruan`, bukan lewat `Pengaturan`. Jadi `master.go`, `sinkron.go`, dan `model.Pengaturan` tidak berubah.
- Pembaruan dicatat di audit log sebagai `PEMBARUAN` pada entitas `aplikasi`.

## 1. `internal/versi/versi.go` (berkas baru)

```go
// Package versi menyimpan nomor versi aplikasi. Nilainya disuntikkan saat build:
//
//	wails build -ldflags "-X mbg/internal/versi.Versi=0.4.0"
package versi

var Versi = "dev"
```

## 2. `internal/pembaruan/pembaruan.go` (berkas baru)

```go
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
		HTTP:     &http.Client{},
		APIBase:  "https://api.github.com",
		Repo:     RepoBawaan,
		HostRepo: "github.com",
		HostLain: []string{"objects.githubusercontent.com", "release-assets.githubusercontent.com"},
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
```

## 3. `internal/model/model.go`: tambahkan di akhir berkas

```go
// ---------- Pembaruan aplikasi ----------

type OpsiPembaruan struct {
	Otomatis bool `json:"otomatis"` // periksa otomatis saat aplikasi dibuka
	PraRilis bool `json:"pra_rilis"`
}

// Cara: WINDOWS | LINUX_DEB | LINUX_RPM | MANUAL (tidak ada berkas pemasang yang cocok).
type InfoPembaruan struct {
	Ada          bool   `json:"ada"`
	VersiSaatIni string `json:"versi_saat_ini"`
	VersiBaru    string `json:"versi_baru"`
	Catatan      string `json:"catatan"`
	URLHalaman   string `json:"url_halaman"`
	Cara         string `json:"cara"`
	NamaAset     string `json:"nama_aset"`
	Ukuran       int64  `json:"ukuran"`
	ChecksumAda  bool   `json:"checksum_ada"`
	Pesan        string `json:"pesan"`
}

type HasilPasang struct {
	Berkas   string `json:"berkas"`
	Cadangan string `json:"cadangan"`
	Perintah string `json:"perintah"` // Linux: perintah yang dijalankan pengguna sendiri
	Pesan    string `json:"pesan"`
	Keluar   bool   `json:"keluar"` // Windows: aplikasi akan ditutup
}
```

## 4. `internal/service/service.go` (timpa utuh)

```go
package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"mbg/internal/timeutil"
	"mbg/internal/versi"
)

type pendingInput struct {
	KelasID int64
	Nama    string
}

type Service struct {
	db  *sql.DB
	dir string
	now func() time.Time

	mu         sync.Mutex
	lastUID    string
	lastUIDAt  time.Time
	pending    *pendingInput
	lastSubmit map[int64]time.Time // cooldown per kelas
	rdr        readerState
	app        appCtx
	siswaNISN  string
	siswaNama  string
	pemb       pembState
}

func New(db *sql.DB, dir string) *Service {
	return NewWithClock(db, dir, timeutil.Now)
}

func NewWithClock(db *sql.DB, dir string, now func() time.Time) *Service {
	return &Service{db: db, dir: dir, now: now, lastSubmit: map[int64]time.Time{}}
}

type InfoSistem struct {
	Versi   string `json:"versi"`
	DataDir string `json:"data_dir"`
}

// InfoSistem: versi dibaca dari internal/versi (disuntikkan lewat -ldflags saat build).
func (s *Service) InfoSistem() InfoSistem { return InfoSistem{Versi: versi.Versi, DataDir: s.dir} }

func js(v any) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func (s *Service) audit(tx *sql.Tx, now time.Time, aksi, entitas, id string, lama, baru any, alasan string) error {
	_, err := tx.Exec(`INSERT INTO audit_log (waktu,aksi,entitas,entitas_id,nilai_lama,nilai_baru,alasan) VALUES (?,?,?,?,?,?,?)`,
		now.In(timeutil.WIB).Format(time.RFC3339), aksi, entitas, id, js(lama), js(baru), alasan)
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	return nil
}
```

## 5. `internal/db/db.go`: dua perubahan

a. Ubah `tambahKolom` agar menerima tipe kolom. Ganti baris tanda tangan dan baris `ALTER`:

```go
func tambahKolom(conn *sql.DB, tabel, kolom, tipe string) error {
```

```go
	_, err = conn.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, tabel, kolom, tipe))
```

b. Di `Migrate`, ubah pemanggilan di dalam loop yang sudah ada menjadi `tambahKolom(conn, k.tabel, k.kolom, "TEXT")`. Lalu sisipkan blok ini tepat setelah loop itu (sebelum komentar `// Backfill`):

```go
	for _, k := range []struct{ tabel, kolom, tipe string }{
		{"pengaturan", "periksa_otomatis", "INTEGER NOT NULL DEFAULT 1"},
		{"pengaturan", "pra_rilis", "INTEGER NOT NULL DEFAULT 0"},
	} {
		if err := tambahKolom(conn, k.tabel, k.kolom, k.tipe); err != nil {
			return err
		}
	}
```

## 6. `internal/service/pembaruan.go` (berkas baru)

```go
package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"mbg/internal/model"
	"mbg/internal/pembaruan"
	"mbg/internal/timeutil"
	"mbg/internal/versi"
)

const jumlahCadangan = 5

// pembState: hasil pemeriksaan terakhir. Bidang bertanda (uji) hanya diisi oleh tes.
type pembState struct {
	mu       sync.Mutex
	rilis    *pembaruan.Rilis
	aset     *pembaruan.Aset
	checksum *pembaruan.Aset
	cara     string
	sibuk    bool

	klien     *pembaruan.Klien // (uji)
	goos      string           // (uji)
	osRelease func() string    // (uji)
	dirUnduh  string           // (uji)
	luncur    func(string) error
}

func (s *Service) klienPemb() *pembaruan.Klien {
	if s.pemb.klien != nil {
		return s.pemb.klien
	}
	return pembaruan.KlienBawaan(versi.Versi)
}

func (s *Service) goosPemb() string {
	if s.pemb.goos != "" {
		return s.pemb.goos
	}
	return runtime.GOOS
}

func (s *Service) bacaOSRelease() string {
	if s.pemb.osRelease != nil {
		return s.pemb.osRelease()
	}
	b, _ := os.ReadFile("/etc/os-release")
	return string(b)
}

// ---------- opsi ----------

func (s *Service) GetOpsiPembaruan() (model.OpsiPembaruan, error) {
	var a, p int
	err := s.db.QueryRow(`SELECT periksa_otomatis,pra_rilis FROM pengaturan WHERE id=1`).Scan(&a, &p)
	return model.OpsiPembaruan{Otomatis: a == 1, PraRilis: p == 1}, err
}

func (s *Service) SimpanOpsiPembaruan(o model.OpsiPembaruan) error {
	lama, _ := s.GetOpsiPembaruan() // sebelum Begin: koneksi tunggal
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE pengaturan SET periksa_otomatis=?,pra_rilis=? WHERE id=1`, boolInt(o.Otomatis), boolInt(o.PraRilis)); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "pengaturan", "1", lama, o, "opsi pembaruan"); err != nil {
		return err
	}
	return tx.Commit()
}

// ---------- periksa ----------

// PeriksaPembaruan: otomatis = true dipakai saat aplikasi dibuka. Mode itu tidak memanggil
// jaringan bila dimatikan di pengaturan, dan gagal diam-diam (tanpa galat) bila offline.
// Tombol manual (otomatis = false) selalu memeriksa dan melaporkan galat.
func (s *Service) PeriksaPembaruan(otomatis bool) (model.InfoPembaruan, error) {
	info := model.InfoPembaruan{VersiSaatIni: versi.Versi}
	opsi, err := s.GetOpsiPembaruan()
	if err != nil {
		opsi = model.OpsiPembaruan{Otomatis: true}
	}
	if otomatis && !opsi.Otomatis {
		return info, nil
	}
	if _, ok := pembaruan.ParseVersi(versi.Versi); !ok {
		info.Pesan = "Versi pengembangan: pemeriksaan pembaruan dilewati"
		return info, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := s.klienPemb().Periksa(ctx, opsi.PraRilis)
	if err != nil {
		if otomatis {
			return info, nil
		}
		if errors.Is(err, pembaruan.ErrTidakAda) {
			info.Pesan = "Belum ada rilis di GitHub"
			return info, nil
		}
		return info, errors.New("Tidak dapat memeriksa pembaruan (periksa koneksi internet): " + err.Error())
	}
	if !pembaruan.LebihBaru(versi.Versi, r, opsi.PraRilis) {
		info.Pesan = "Aplikasi sudah versi terbaru"
		return info, nil
	}

	goos := s.goosPemb()
	famili := ""
	if goos == "linux" {
		famili = pembaruan.FamiliDistro(s.bacaOSRelease())
	}
	cara := "MANUAL"
	aset, ada := pembaruan.PilihAset(r, goos, famili)
	if ada {
		switch {
		case goos == "windows":
			cara = "WINDOWS"
		case famili == "deb":
			cara = "LINUX_DEB"
		case famili == "rpm":
			cara = "LINUX_RPM"
		}
	}
	cs := r.CariAset(pembaruan.NamaChecksum)

	info.Ada, info.VersiBaru, info.URLHalaman, info.Cara = true, r.Versi, r.URLHalaman, cara
	info.Catatan = r.Catatan
	if len(info.Catatan) > 4000 {
		info.Catatan = info.Catatan[:4000] + "..."
	}
	info.ChecksumAda = cs != nil
	if ada {
		info.NamaAset, info.Ukuran = aset.Nama, aset.Ukuran
	}

	s.pemb.mu.Lock()
	s.pemb.rilis, s.pemb.cara, s.pemb.checksum = &r, cara, cs
	if ada {
		a := aset
		s.pemb.aset = &a
	} else {
		s.pemb.aset = nil
	}
	s.pemb.mu.Unlock()
	return info, nil
}

// ---------- cadangan ----------

// CadangkanDB: salinan konsisten database (VACUUM INTO, setara .backup) ke <data>/cadangan.
// Hanya 5 cadangan terakhir yang disimpan.
func (s *Service) CadangkanDB() (string, error) {
	dir := filepath.Join(s.dir, "cadangan")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	p := filepath.Join(dir, "mbg-"+s.now().In(timeutil.WIB).Format("20060102-150405")+".db")
	lit := strings.ReplaceAll(filepath.ToSlash(p), "'", "''")
	if _, err := s.db.Exec(`VACUUM INTO '` + lit + `'`); err != nil {
		return "", err
	}
	_ = os.Chmod(p, 0o600)
	pangkasCadangan(dir, jumlahCadangan)
	return p, nil
}

func pangkasCadangan(dir string, simpan int) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var nama []string
	for _, e := range ents {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "mbg-") && strings.HasSuffix(e.Name(), ".db") {
			nama = append(nama, e.Name())
		}
	}
	sort.Strings(nama) // stempel waktu di nama: urutan abjad = urutan waktu
	for len(nama) > simpan {
		_ = os.Remove(filepath.Join(dir, nama[0]))
		nama = nama[1:]
	}
}

// ---------- pasang ----------

func (s *Service) emitProgres(selesai, total int64) {
	if s.app.ctx == nil {
		return
	}
	wruntime.EventsEmit(s.app.ctx, "pembaruan:progres", map[string]int64{"selesai": selesai, "total": total})
}

func folderUnduhan() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	for _, sub := range []string{"Downloads", "Unduhan"} {
		p := filepath.Join(h, sub)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return h
}

func kutip(p string) string { return "'" + strings.ReplaceAll(p, "'", `'\''`) + "'" }

// luncurkanInstaller (Windows): lewat shell Explorer supaya prompt UAC installer bisa tampil.
func luncurkanInstaller(path string) error {
	return exec.Command("explorer.exe", path).Start()
}

func (s *Service) catatPembaruan(lama, baru, berkas, cadangan string, terverifikasi bool) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := s.audit(tx, s.now(), "PEMBARUAN", "aplikasi", baru, map[string]any{"versi": lama},
		map[string]any{"versi": baru, "berkas": filepath.Base(berkas), "cadangan": filepath.Base(cadangan), "checksum_terverifikasi": terverifikasi}, ""); err != nil {
		return err
	}
	return tx.Commit()
}

// PasangPembaruan: cadangkan DB, unduh (dengan progres via event pembaruan:progres), verifikasi,
// lalu Windows: jalankan installer dan tutup aplikasi; Linux: kembalikan perintah sudo untuk
// dijalankan pengguna sendiri (sudo tidak pernah dijalankan otomatis).
func (s *Service) PasangPembaruan() (model.HasilPasang, error) {
	var out model.HasilPasang
	st := &s.pemb
	st.mu.Lock()
	if st.sibuk {
		st.mu.Unlock()
		return out, errors.New("Pembaruan sedang berjalan")
	}
	if st.rilis == nil || st.aset == nil || st.cara == "MANUAL" {
		st.mu.Unlock()
		return out, errors.New("Tidak ada berkas pemasang untuk sistem ini; unduh dari halaman rilis")
	}
	rilis, aset, cs, cara := *st.rilis, *st.aset, st.checksum, st.cara
	st.sibuk = true
	st.mu.Unlock()
	defer func() { st.mu.Lock(); st.sibuk = false; st.mu.Unlock() }()

	if !pembaruan.LebihBaru(versi.Versi, rilis, true) {
		return out, errors.New("Versi di GitHub tidak lebih baru dari versi terpasang")
	}
	cad, err := s.CadangkanDB()
	if err != nil {
		return out, fmt.Errorf("Cadangan database gagal, pembaruan dibatalkan: %w", err)
	}
	out.Cadangan = cad

	var dir string
	if s.goosPemb() == "windows" {
		d, err := os.MkdirTemp("", "mbg-pembaruan-")
		if err != nil {
			return out, err
		}
		dir = d
	} else if dir = st.dirUnduh; dir == "" {
		dir = folderUnduhan()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	var terakhir time.Time
	hasil, err := s.klienPemb().Unduh(ctx, aset, cs, dir, func(sel, tot int64) {
		if sel < tot && time.Since(terakhir) < 150*time.Millisecond {
			return
		}
		terakhir = time.Now()
		s.emitProgres(sel, tot)
	})
	if err != nil {
		return out, fmt.Errorf("Unduhan gagal: %w", err)
	}
	out.Berkas = hasil.Path
	if err := s.catatPembaruan(versi.Versi, rilis.Versi, hasil.Path, cad, hasil.Terverifikasi); err != nil {
		return out, err
	}
	peringatan := ""
	if !hasil.Terverifikasi {
		peringatan = " Peringatan: rilis ini tanpa checksums.txt, hanya ukuran berkas yang diperiksa."
	}

	if cara == "WINDOWS" {
		fn := st.luncur
		if fn == nil {
			fn = luncurkanInstaller
		}
		if err := fn(hasil.Path); err != nil {
			return out, fmt.Errorf("Installer tidak dapat dijalankan: %w", err)
		}
		out.Keluar = true
		out.Pesan = "Installer dijalankan; aplikasi akan ditutup." + peringatan
		if ctx := s.app.ctx; ctx != nil {
			go func() { time.Sleep(1500 * time.Millisecond); wruntime.Quit(ctx) }()
		}
		return out, nil
	}
	pengelola := "apt"
	if cara == "LINUX_RPM" {
		pengelola = "dnf"
	}
	out.Perintah = "sudo " + pengelola + " install " + kutip(hasil.Path)
	out.Pesan = "Berkas diunduh dan diverifikasi. Jalankan perintah di bawah di terminal (butuh sudo), lalu buka ulang aplikasi." + peringatan
	return out, nil
}
```

## 7. Frontend

**`frontend/index.html`**, empat sisipan:

a. Di `<header>`, setelah `<span id="status-reader" ...>`:

```html
        <button id="badge-pembaruan" class="badge hidden"></button>
```

b. Di `#sub-instansi`, setelah `<div class="form-baris">` yang berisi tombol logo dan sebelum `<div class="form-baris"><button id="p-simpan" ...>`:

```html
                <div class="form-baris">
                    <button id="pb-periksa">Periksa pembaruan</button>
                    <label class="cek"><input type="checkbox" id="pb-otomatis"> Periksa pembaruan otomatis</label>
                    <label class="cek"><input type="checkbox" id="pb-pra"> Sertakan pra-rilis</label>
                </div>
```

c. Di `#ra-entitas`, setelah `<option>master</option>`:

```html
                        <option>aplikasi</option>
```

d. Di daftar skrip, setelah `js/sinkron.js` dan sebelum `js/main.js`:

```html
    <script src="js/pembaruan.js"></script>
```

**`frontend/js/pembaruan.js`** (berkas baru):

```js
const Pembaruan = {
    info: null,

    tampilBadge(i) {
        const b = _r("badge-pembaruan");
        b.textContent = "Pembaruan v" + i.versi_baru;
        b.classList.remove("hidden");
    },

    // dipanggil beberapa detik setelah aplikasi terbuka; gagal diam-diam
    async otomatis() {
        try {
            const i = await API.PeriksaPembaruan(true);
            if (i && i.ada) { this.info = i; this.tampilBadge(i); }
        } catch (e) { }
    },

    async manual() {
        try {
            const i = await API.PeriksaPembaruan(false);
            this.info = i;
            if (i.ada) { this.tampilBadge(i); this.tampil(i); }
            else toast(i.pesan || "Aplikasi sudah versi terbaru", "ok");
        } catch (e) { toast(String(e.message || e), "err"); }
    },

    async muatOpsi() {
        try {
            const o = await API.GetOpsiPembaruan();
            _r("pb-otomatis").checked = o.otomatis; _r("pb-pra").checked = o.pra_rilis;
        } catch (e) { }
    },

    async simpanOpsi() {
        try {
            await API.SimpanOpsiPembaruan({ otomatis: _r("pb-otomatis").checked, pra_rilis: _r("pb-pra").checked });
            toast("Pengaturan pembaruan disimpan", "ok");
        } catch (e) { toast(String(e.message || e), "err"); }
    },

    tutup() { Dash.tutupModal(); pasangHandler(); },

    tampil(i) {
        Tap.handler = () => toast("Selesaikan pembaruan dulu", "info");
        const manual = i.cara === "MANUAL";
        const mb = i.ukuran ? ` (${(i.ukuran / 1048576).toFixed(1)} MB)` : "";
        Dash.bukaModal(`
          <h2>Pembaruan tersedia: v${esc(i.versi_baru)}</h2>
          <p>Versi terpasang: v${esc(i.versi_saat_ini)}${manual ? "" : " &middot; Berkas: " + esc(i.nama_aset) + mb}</p>
          <pre class="catatan-rilis">${esc(i.catatan || "(tanpa catatan rilis)")}</pre>
          ${manual ? `<div class="peringatan">Tidak ada berkas pemasang yang cocok untuk sistem ini. Unduh manual dari halaman rilis.</div>` : ""}
          ${!manual && !i.checksum_ada ? `<div class="peringatan">Rilis ini tanpa checksums.txt: hanya ukuran berkas yang diperiksa.</div>` : ""}
          <p class="kecil">Database dicadangkan otomatis sebelum pembaruan dimulai.</p>
          <progress id="pb-prog" class="hidden" max="100" value="0"></progress>
          <div class="kecil" id="pb-status"></div>
          <div class="galat" id="pb-galat"></div>
          <div class="aksi"><button id="pb-batal">Nanti</button><button id="pb-ok">${manual ? "Buka halaman rilis" : "Unduh dan pasang"}</button></div>`);
        _r("pb-batal").onclick = () => this.tutup();
        _r("pb-ok").onclick = () => manual ? this.bukaHalaman(i.url_halaman) : this.pasang();
    },

    bukaHalaman(u) {
        if (!/^https:\/\/github\.com\//.test(u || "")) { toast("Alamat rilis tidak valid", "err"); return; }
        if (window.runtime && window.runtime.BrowserOpenURL) window.runtime.BrowserOpenURL(u);
        else window.open(u, "_blank");
    },

    progres(p) {
        const el = _r("pb-prog"); if (!el) return;
        const pct = p.total > 0 ? Math.min(100, Math.floor(p.selesai * 100 / p.total)) : 0;
        el.classList.remove("hidden"); el.value = pct;
        _r("pb-status").textContent = `Mengunduh... ${pct}%`;
    },

    async pasang() {
        _r("pb-ok").disabled = true; _r("pb-batal").disabled = true;
        _r("pb-galat").textContent = ""; _r("pb-status").textContent = "Menyiapkan cadangan dan unduhan...";
        try {
            const h = await API.PasangPembaruan();
            if (h.keluar) { _r("pb-status").textContent = h.pesan; return; }
            this.hasilLinux(h);
        } catch (e) {
            _r("pb-galat").textContent = String(e.message || e); _r("pb-status").textContent = "";
            _r("pb-ok").disabled = false; _r("pb-batal").disabled = false; bunyi("err");
        }
    },

    hasilLinux(h) {
        Dash.bukaModal(`
          <h2>Siap dipasang</h2>
          <p>${esc(h.pesan)}</p>
          <input id="pb-cmd" readonly value="${esc(h.perintah)}">
          <p class="kecil">Berkas: ${esc(h.berkas)}<br>Cadangan database: ${esc(h.cadangan)}</p>
          <div class="aksi"><button id="pb-tutup">Tutup</button><button id="pb-salin">Salin perintah</button></div>`);
        _r("pb-tutup").onclick = () => this.tutup();
        _r("pb-salin").onclick = async () => {
            const el = _r("pb-cmd");
            try { await navigator.clipboard.writeText(el.value); }
            catch (e) { el.select(); document.execCommand("copy"); }
            toast("Perintah disalin", "ok");
        };
    },
};

// opsi pembaruan dimuat bersama tab Master > Instansi
const _muatInstansiPb = Master.muatInstansi.bind(Master);
Master.muatInstansi = async function () { await _muatInstansiPb(); await Pembaruan.muatOpsi(); };

document.addEventListener("DOMContentLoaded", () => {
    _r("pb-periksa").onclick = () => Pembaruan.manual();
    _r("pb-otomatis").onchange = () => Pembaruan.simpanOpsi();
    _r("pb-pra").onchange = () => Pembaruan.simpanOpsi();
    _r("badge-pembaruan").onclick = () => { if (Pembaruan.info && Pembaruan.info.ada) Pembaruan.tampil(Pembaruan.info); };
    if (window.runtime && window.runtime.EventsOn) window.runtime.EventsOn("pembaruan:progres", p => Pembaruan.progres(p));
    setTimeout(() => Pembaruan.otomatis(), 5000);
});
```

**`frontend/css/style.css`**, tambahkan di akhir:

```css
/* Pembaruan */
.badge {
    padding: 4px 12px;
    font-size: 12px;
    color: #fff;
    background: #1d4ed8;
    border-color: #1d4ed8;
    border-radius: 999px
}

.badge:hover:not(:disabled) {
    background: #1e40af;
    border-color: #1e40af
}

.catatan-rilis {
    max-height: 200px;
    overflow: auto;
    margin: 10px 0;
    padding: 10px 12px;
    font: inherit;
    font-size: 13px;
    white-space: pre-wrap;
    background: #f8fafc;
    border: 1px solid var(--garis);
    border-radius: 8px
}

#pb-prog {
    width: 100%;
    margin: 6px 0
}

#pb-cmd {
    width: 100%;
    font-family: ui-monospace, "Cascadia Mono", Menlo, monospace;
    font-size: 13px
}

#pb-ok {
    color: #fff;
    background: var(--hijau);
    border-color: var(--hijau)
}

#pb-ok:hover:not(:disabled) {
    background: var(--hijau-tua);
    border-color: var(--hijau-tua)
}
```

## 8. Tes

**`internal/pembaruan/pembaruan_test.go`** (berkas baru):

```go
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
	var cs *Asset
	_ = cs
	var csAset *Aset
	if csBody != "" {
		csAset = &Aset{Nama: NamaChecksum, URL: f.srv.URL + pref + NamaChecksum}
	}
	dir := t.TempDir()
	h, err := k.Unduh(context.Background(), a, csAset, dir, nil)
	return h, dir, err
}
```

Saya perlu mengoreksi bagian akhir blok di atas: dua baris `var cs *Asset` dan `_ = cs` adalah sisa yang tidak valid. Hapus keduanya, lalu lanjutkan berkas dengan tes berikut (sambungan `pembaruan_test.go`, tepat setelah fungsi `jalankanUnduh`):

```go
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
```

**`internal/service/pembaruan_test.go`** (berkas baru):

```go
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
	"mbg_0.2.0_amd64.deb":    "deb palsu",
	"mbg-0.2.0-1.x86_64.rpm": "rpm palsu",
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
```

## 9. Rilis: `nfpm.yaml` dan README

**`nfpm.yaml`**: ganti baris versi agar dibaca dari variabel lingkungan:

```yaml
version: ${VERSION}
```

**`README.md`**, tiga penggantian:

a. Di bagian **Build**, ganti blok "Build semua target" dengan:

````markdown
```bash
go vet ./... && go test ./... || exit 1
rm -rf dist && mkdir -p dist
export VERSION=0.4.0
LDF="-X mbg/internal/versi.Versi=${VERSION}"

wails build -clean -ldflags "$LDF"
cp build/bin/mbg dist/mbg-linux-amd64

wails build -clean -platform windows/amd64 -nsis -webview2 embed -ldflags "$LDF"
cp build/bin/mbg-amd64-installer.exe dist/

nfpm package -p deb -t dist/
nfpm package -p rpm -t dist/

(cd dist && sha256sum mbg-linux-amd64 mbg_*.deb mbg-*.rpm mbg-amd64-installer.exe > checksums.txt)
```

Hasil di `dist/`: `mbg-linux-amd64`, `.deb`, `.rpm`, `mbg-amd64-installer.exe`, dan `checksums.txt` (dipakai fitur pembaruan untuk verifikasi SHA256).
````

b. Di bagian **Rilis**, langkah 1 diganti dengan:

````markdown
1. Tentukan nomor versi (contoh `0.4.1`). Versi aplikasi dan paket Linux diambil dari variabel `VERSION` saat build, jadi tidak ada lagi string versi yang diubah di kode:

   ```bash
   export VERSION=0.4.1
   sed -i "s/0\.4\.0/${VERSION}/g" README.md
   ```
````

c. Tambahkan di bagian **Fitur**:

```markdown
- Pembaruan aplikasi dari GitHub Release (periksa otomatis, verifikasi checksum, cadangan database otomatis sebelum memperbarui)
```

Catatan untuk bagian Data: cadangan otomatis tersimpan di `cadangan/` dalam folder data (`~/.local/share/mbg/cadangan/` atau `%APPDATA%\MBG\cadangan\`), lima terakhir.

## 10. Jalankan dan commit

```bash
gofmt -l . ; go vet ./... && go test ./...
export VERSION=0.4.0 && wails build -ldflags "-X mbg/internal/versi.Versi=${VERSION}"
git add -A
git commit -m "Pembaruan aplikasi via GitHub Release dengan verifikasi checksum dan cadangan otomatis"
```

Hal yang perlu diwaspadai:
- **Blok `pembaruan_test.go` di paket `pembaruan`** berasal dari dua potongan. Setelah menyalin, pastikan dua baris `var cs *Asset` dan `_ = cs` di dalam `jalankanUnduh` sudah dihapus, karena kalau tidak, kompilasi gagal.
- **Peluncuran installer Windows** lewat `explorer.exe` belum bisa saya uji dari sini. Uji sekali di Windows: pembaruan dari `0.1.x` ke versi uji harus memunculkan prompt UAC lalu wizard NSIS. Bila tidak muncul, kirim perilakunya dan saya ganti dengan `cmd /C start`.
- **Rilis lama (v0.1.1) tidak punya `checksums.txt`**, jadi pada rilis itu modal menampilkan peringatan "hanya ukuran berkas yang diperiksa", sesuai keputusan 2.
- **Pengujian end-to-end:** setelah build `-ldflags` versi `0.1.1`, buat rilis uji (misalnya `v0.4.0-uji` sebagai pra-rilis), centang "Sertakan pra-rilis", lalu klik "Periksa pembaruan".
- **`.rpm` dari Debian 13** tetap tidak cocok untuk RHEL/Rocky 9 (catatan README yang sudah ada), jadi fitur ini memasang `.rpm` apa adanya dari rilis.

## 11. Ringkasan dan versi

- **Versi:** satu sumber di `internal/versi`, disuntikkan lewat `-ldflags`. `nfpm.yaml` membaca `${VERSION}`, dan langkah `sed` untuk `service.go` dan `nfpm.yaml` hilang.
- **Pemeriksaan:** GitHub API `releases/latest` tanpa token (timeout 15 detik, User-Agent `mbg/<versi>`). Otomatis saat aplikasi dibuka dan diam-diam bila offline, ditambah tombol manual. Opsi mematikan otomatis dan opsi pra-rilis tersimpan di pengaturan.
- **Keamanan:** hanya HTTPS, host `github.com` (path wajib milik repo) atau dua host CDN GitHub, dan setiap pengalihan divalidasi ulang. Ukuran dicek, dan SHA256 dicocokkan dengan `checksums.txt`. Checksum tidak cocok berarti berkas dihapus dan pembaruan ditolak. Tidak pernah menurunkan versi.
- **Pemasangan:** Windows mengunduh lalu menjalankan installer dan menutup aplikasi. Linux mengunduh `.deb` atau `.rpm` sesuai distribusi, lalu menampilkan perintah `sudo` untuk dijalankan sendiri.
- **Cadangan:** `VACUUM INTO` ke folder `cadangan/` sebelum pembaruan dimulai, lima terakhir disimpan, dan pembaruan dibatalkan bila cadangan gagal.
- **Audit:** satu entri `PEMBARUAN` per pemasangan, plus entri perubahan opsi pembaruan.

Saran versi **0.4.0**, sesuai rencana urutan 0.2.0 (susulan), 0.3.0 (sinkronisasi), dan 0.4.0 (pembaruan). Jika ketiganya dirilis sekaligus, cukup satu rilis `0.4.0`. Fitur pembaruan baru berguna di rilis berikutnya, karena baru versi yang memuat kodenya yang bisa memeriksa rilis di atasnya.

Ketiga fitur sudah selesai. Kirim hasil `gofmt -l .`, `go vet ./...`, dan `go test ./...` bila ada yang gagal, supaya bisa saya perbaiki sebelum rilis.