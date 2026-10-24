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
