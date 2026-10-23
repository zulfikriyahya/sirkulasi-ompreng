package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"mbg/internal/export"
	"mbg/internal/timeutil"
)

type appCtx struct{ ctx context.Context }

// SetContext dipanggil dari main.go saat Wails start (dibutuhkan untuk dialog file).
func (s *Service) SetContext(ctx context.Context) { s.app.ctx = ctx }

func periodeTeks(dari, sampai time.Time) string {
	if dari.Equal(sampai) {
		return "Hari/Tanggal: " + timeutil.HariTanggalIndo(dari)
	}
	if dari.Year() == sampai.Year() && dari.Month() == sampai.Month() {
		return fmt.Sprintf("Periode: %d - %s", dari.Day(), timeutil.TanggalIndo(sampai))
	}
	return "Periode: " + timeutil.TanggalIndo(dari) + " - " + timeutil.TanggalIndo(sampai)
}

func angka(n sql.NullInt64) string {
	if !n.Valid {
		return "-"
	}
	return fmt.Sprint(n.Int64)
}

func wib(j string) string {
	if j == "" {
		return "-"
	}
	return j + " WIB"
}

// susunLaporan: baris TIDAK_AMBIL dikecualikan. Status AMBIL (belum kembali, hari belum ditutup)
// diperlakukan sebagai "Belum kembali".
func (s *Service) susunLaporan(dari, sampai string) (export.Laporan, error) {
	var l export.Laporan
	d, err1 := time.ParseInLocation("2006-01-02", dari, timeutil.WIB)
	e, err2 := time.ParseInLocation("2006-01-02", sampai, timeutil.WIB)
	if err1 != nil || err2 != nil {
		return l, errors.New("Tanggal tidak valid")
	}
	if e.Before(d) {
		return l, errors.New("Tanggal akhir tidak boleh sebelum tanggal awal")
	}
	p, err := s.GetPengaturan()
	if err != nil {
		return l, err
	}
	var mime sql.NullString
	if err := s.db.QueryRow(`SELECT logo, logo_mime FROM pengaturan WHERE id=1`).Scan(&l.Logo, &mime); err != nil {
		return l, err
	}
	l.LogoMime = mime.String
	l.NamaInstansi, l.Alamat = p.NamaInstansi, p.AlamatInstansi
	l.Jabatan, l.Nama, l.NIP = p.JabatanPenanggungJawab, p.PenanggungJawab, p.NIPPenanggungJawab
	l.Periode = periodeTeks(d, e)
	l.TanggalCetak = timeutil.TanggalIndo(s.now())

	rows, err := s.db.Query(`SELECT tanggal,nama_kelas,nama_ketua,COALESCE(waktu_ambil,''),COALESCE(waktu_kembali,''),
		jumlah_ambil,jumlah_kembali,status,COALESCE(catatan,''),COALESCE(nama_ambil,''),COALESCE(nama_kembali,'') FROM transaksi
		WHERE tanggal>=? AND tanggal<=? AND status!='TIDAK_AMBIL'
		ORDER BY tanggal,(SELECT urutan FROM kelas k WHERE k.id=transaksi.kelas_id),nama_kelas`, dari, sampai)
	if err != nil {
		return l, err
	}
	defer rows.Close()
	for rows.Next() {
		var tgl, kelas, ketua, wa, wk, status, cat, pengambil, pengembali string
		var ja, jk sql.NullInt64
		if err := rows.Scan(&tgl, &kelas, &ketua, &wa, &wk, &ja, &jk, &status, &cat, &pengambil, &pengembali); err != nil {
			return l, err
		}
		var pja, pjk *int
		if ja.Valid {
			v := int(ja.Int64)
			pja = &v
		}
		if jk.Valid {
			v := int(jk.Int64)
			pjk = &v
		}
		if status == "AMBIL" {
			status = "BELUM_KEMBALI"
		}
		ket, merah := Keterangan(status, pja, pjk, cat)
		t, _ := time.ParseInLocation("2006-01-02", tgl, timeutil.WIB)
		l.Baris = append(l.Baris, export.Baris{
			Kelas: kelas, Ketua: ketua, Tanggal: timeutil.TanggalIndo(t),
			WaktuAmbil: wib(wa), WaktuKembali: wib(wk),
			JumlahAmbil: angka(ja), JumlahKembali: angka(jk), Keterangan: ket, Merah: merah,
			Pengambil: strip(pengambil), Pengembali: strip(pengembali),
		})
	}
	if err := rows.Err(); err != nil {
		return l, err
	}
	if len(l.Baris) == 0 {
		return l, errors.New("Tidak ada data pada periode ini")
	}
	return l, nil
}

func (s *Service) simpanDialog(nama, ext, label string) (string, error) {
	if s.app.ctx == nil {
		return "", errors.New("Dialog file belum siap")
	}
	p, err := wruntime.SaveFileDialog(s.app.ctx, wruntime.SaveDialogOptions{
		Title: "Simpan laporan", DefaultFilename: nama,
		Filters: []wruntime.FileFilter{{DisplayName: label, Pattern: "*." + ext}},
	})
	if err != nil || p == "" {
		return "", err
	}
	if !strings.HasSuffix(strings.ToLower(p), "."+ext) {
		p += "." + ext
	}
	return p, nil
}

func namaBerkas(dari, sampai, ext string) string {
	if dari == sampai {
		return fmt.Sprintf("laporan-mbg-%s.%s", dari, ext)
	}
	return fmt.Sprintf("laporan-mbg-%s_sd_%s.%s", dari, sampai, ext)
}

// EksporExcel / EksporPDF: mengembalikan path berkas, atau "" jika dialog dibatalkan.
func (s *Service) EksporExcel(dari, sampai string) (string, error) {
	l, err := s.susunLaporan(dari, sampai)
	if err != nil {
		return "", err
	}
	p, err := s.simpanDialog(namaBerkas(dari, sampai, "xlsx"), "xlsx", "Excel (*.xlsx)")
	if err != nil || p == "" {
		return "", err
	}
	if err := export.Excel(l, p); err != nil {
		log.Printf("[ekspor] excel gagal: %v", err)
		return "", err
	}
	return p, nil
}

func (s *Service) EksporPDF(dari, sampai string) (string, error) {
	l, err := s.susunLaporan(dari, sampai)
	if err != nil {
		return "", err
	}
	p, err := s.simpanDialog(namaBerkas(dari, sampai, "pdf"), "pdf", "PDF (*.pdf)")
	if err != nil || p == "" {
		return "", err
	}
	if err := export.PDF(l, p); err != nil {
		log.Printf("[ekspor] pdf gagal: %v", err)
		return "", err
	}
	return p, nil
}

// ---------- Logo ----------

func (s *Service) AdaLogo() bool {
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM pengaturan WHERE id=1 AND logo IS NOT NULL AND length(logo)>0`).Scan(&n)
	return n > 0
}

// PilihLogo membuka dialog pilih gambar (PNG/JPEG, maks 1 MB). Mengembalikan true jika logo tersimpan.
func (s *Service) PilihLogo() (bool, error) {
	if s.app.ctx == nil {
		return false, errors.New("Dialog file belum siap")
	}
	path, err := wruntime.OpenFileDialog(s.app.ctx, wruntime.OpenDialogOptions{
		Title:   "Pilih logo instansi",
		Filters: []wruntime.FileFilter{{DisplayName: "Gambar (PNG/JPEG)", Pattern: "*.png;*.jpg;*.jpeg"}},
	})
	if err != nil || path == "" {
		return false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if len(data) > 1<<20 {
		return false, errors.New("Ukuran logo maksimal 1 MB")
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg") {
		return false, errors.New("Logo harus berupa gambar PNG atau JPEG")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE pengaturan SET logo=?,logo_mime=? WHERE id=1`, data, "image/"+format); err != nil {
		return false, err
	}
	if err := s.audit(tx, s.now(), "UBAH", "pengaturan", "1", nil, map[string]any{"logo": format, "ukuran": len(data)}, "logo instansi"); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (s *Service) HapusLogo() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE pengaturan SET logo=NULL,logo_mime=NULL WHERE id=1`); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "pengaturan", "1", map[string]any{"logo": "ada"}, map[string]any{"logo": "dihapus"}, "logo instansi"); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) PratinjauLaporan(dari, sampai string) ([]export.Baris, error) {
	l, err := s.susunLaporan(dari, sampai)
	if err != nil {
		return []export.Baris{}, err
	}
	return l.Baris, nil
}

func strip(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}
