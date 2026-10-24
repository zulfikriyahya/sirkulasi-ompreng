// Package sinkron membaca dan menulis berkas impor/ekspor data master (xlsx/csv)
// tanpa menyentuh database.
package sinkron

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	SheetPengaturan = "Pengaturan"
	SheetKelas      = "Kelas"
	SheetKartu      = "Kartu"
	SheetWali       = "Wali"
	SheetPetunjuk   = "Petunjuk"
	maksBerkas      = 5 << 20
)

// KunciPengaturan: pengaturan yang disinkronkan (port serial dan ambang HID khusus perangkat, tidak ikut).
var KunciPengaturan = []string{
	"nama_instansi", "alamat_instansi", "penanggung_jawab",
	"nip_penanggung_jawab", "jabatan_penanggung_jawab", "jeda_minimum_menit",
}

var (
	headKelas = []string{"nama", "nama_ketua", "urutan", "aktif", "no_wa_wali"}
	headKartu = []string{"uid", "kelas", "nisn", "nama_siswa", "aktif"}
	headWali  = []string{"nama", "no_wa", "aktif"}
)

var PetunjukTemplate = []string{
	"PETUNJUK IMPOR DATA MASTER MBG",
	"",
	"Sheet Kelas: nama (wajib, kunci pencocokan), nama_ketua (wajib), urutan (angka, bawaan 0), aktif (1/0, bawaan 1).",
	"Sheet Kartu: uid (wajib, angka 1-10 digit, kunci pencocokan), kelas (wajib, harus sama dengan nama kelas), nisn (opsional, 10 digit), nama_siswa (opsional), aktif (1/0, bawaan 1).",
	"Kolom sudah berformat TEKS agar angka nol di depan UID dan NISN tidak hilang. Jangan ubah formatnya menjadi angka.",
	"Mode Gabung hanya menambah dan memperbarui. Mode Ganti juga menghapus kartu dan menonaktifkan kelas yang tidak ada di berkas.",
	"Berkas berisi data pribadi siswa (NISN dan nama): jangan dimasukkan ke repositori atau dibagikan sembarangan.",
	"Sheet Kelas juga punya kolom no_wa_wali (opsional): nomor WhatsApp wali kelas, harus ada di sheet Wali atau sudah terdaftar di aplikasi. Kosongkan untuk melepas wali.",
	"Sheet Wali: nama (wajib), no_wa (wajib, kunci pencocokan; 08xx/62xx diterima), aktif (1/0, bawaan 1).",
}

type KelasBaris struct {
	Baris                      int
	Nama, Ketua, Urutan, Aktif string
	Wali                       string // no_wa wali
	AdaWali                    bool   // kolom no_wa_wali ada di berkas
}

type WaliBaris struct {
	Baris             int
	Nama, NoWA, Aktif string
}

type KartuBaris struct {
	Baris                         int
	UID, Kelas, NISN, Nama, Aktif string
}

type Data struct {
	SHA        string            // sha256 isi berkas (hanya hasil Baca)
	Pengaturan map[string]string // nil = berkas tidak memuat pengaturan
	AdaKelas   bool
	AdaKartu   bool
	AdaWali    bool
	Kelas      []KelasBaris
	Kartu      []KartuBaris
	Wali       []WaliBaris
}

// ---------- Baca ----------

func Baca(path string) (Data, error) {
	var d Data
	st, err := os.Stat(path)
	if err != nil {
		return d, err
	}
	if st.Size() > maksBerkas {
		return d, errors.New("Berkas terlalu besar (maksimal 5 MB)")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return d, err
	}
	sum := sha256.Sum256(raw)
	d.SHA = hex.EncodeToString(sum[:])
	switch strings.ToLower(filepath.Ext(path)) {
	case ".xlsx":
		err = bacaXLSX(raw, &d)
	case ".csv":
		err = bacaCSV(raw, &d)
	default:
		err = errors.New("Format berkas harus .xlsx atau .csv")
	}
	return d, err
}

func kosong(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func petakan(rows [][]string, sheet string, wajib []string) (map[string]int, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("Sheet %q kosong: baris header tidak ada", sheet)
	}
	idx := map[string]int{}
	for i, h := range rows[0] {
		h = strings.ToLower(strings.TrimSpace(h))
		if _, ada := idx[h]; h != "" && !ada {
			idx[h] = i
		}
	}
	for _, w := range wajib {
		if _, ok := idx[w]; !ok {
			return nil, fmt.Errorf("Kolom %q tidak ada pada %q", w, sheet)
		}
	}
	return idx, nil
}

func sel(row []string, idx map[string]int, nama string) string {
	if j, ok := idx[nama]; ok && j < len(row) {
		return strings.TrimSpace(row[j])
	}
	return ""
}

func parseKelas(rows [][]string) ([]KelasBaris, error) {
	idx, err := petakan(rows, SheetKelas, []string{"nama", "nama_ketua"})
	if err != nil {
		return nil, err
	}
	_, adaWali := idx["no_wa_wali"]
	var out []KelasBaris
	for i := 1; i < len(rows); i++ {
		if kosong(rows[i]) {
			continue
		}
		r := rows[i]
		out = append(out, KelasBaris{Baris: i + 1, Nama: sel(r, idx, "nama"), Ketua: sel(r, idx, "nama_ketua"),
			Urutan: sel(r, idx, "urutan"), Aktif: sel(r, idx, "aktif"),
			Wali: sel(r, idx, "no_wa_wali"), AdaWali: adaWali})
	}
	return out, nil
}

func parseWali(rows [][]string) ([]WaliBaris, error) {
	idx, err := petakan(rows, SheetWali, []string{"nama", "no_wa"})
	if err != nil {
		return nil, err
	}
	var out []WaliBaris
	for i := 1; i < len(rows); i++ {
		if kosong(rows[i]) {
			continue
		}
		r := rows[i]
		out = append(out, WaliBaris{Baris: i + 1, Nama: sel(r, idx, "nama"),
			NoWA: sel(r, idx, "no_wa"), Aktif: sel(r, idx, "aktif")})
	}
	return out, nil
}

func parseKartu(rows [][]string) ([]KartuBaris, error) {
	idx, err := petakan(rows, SheetKartu, []string{"uid", "kelas"})
	if err != nil {
		return nil, err
	}
	var out []KartuBaris
	for i := 1; i < len(rows); i++ {
		if kosong(rows[i]) {
			continue
		}
		r := rows[i]
		out = append(out, KartuBaris{Baris: i + 1, UID: sel(r, idx, "uid"), Kelas: sel(r, idx, "kelas"),
			NISN: sel(r, idx, "nisn"), Nama: sel(r, idx, "nama_siswa"), Aktif: sel(r, idx, "aktif")})
	}
	return out, nil
}

func parsePengaturan(rows [][]string) (map[string]string, error) {
	idx, err := petakan(rows, SheetPengaturan, []string{"kunci", "nilai"})
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	for i := 1; i < len(rows); i++ {
		if kosong(rows[i]) {
			continue
		}
		k := strings.ToLower(sel(rows[i], idx, "kunci"))
		if k != "" {
			m[k] = sel(rows[i], idx, "nilai")
		}
	}
	return m, nil
}

func bacaXLSX(raw []byte, d *Data) error {
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return errors.New("Berkas Excel tidak dapat dibaca: " + err.Error())
	}
	defer f.Close()
	for _, nama := range f.GetSheetList() {
		rows, err := f.GetRows(nama)
		if err != nil {
			return err
		}
		switch strings.ToLower(strings.TrimSpace(nama)) {
		case "pengaturan":
			m, err := parsePengaturan(rows)
			if err != nil {
				return err
			}
			d.Pengaturan = m
		case "kelas":
			k, err := parseKelas(rows)
			if err != nil {
				return err
			}
			d.AdaKelas, d.Kelas = true, k
		case "kartu":
			k, err := parseKartu(rows)
			if err != nil {
				return err
			}
			d.AdaKartu, d.Kartu = true, k
		case "wali":
			k, err := parseWali(rows)
			if err != nil {
				return err
			}
			d.AdaWali, d.Wali = true, k
		}
	}
	if d.Pengaturan == nil && !d.AdaKelas && !d.AdaKartu && !d.AdaWali {
		return errors.New("Berkas tidak memuat sheet Pengaturan, Kelas, Kartu, maupun Wali")
	}
	return nil
}

func bacaCSV(raw []byte, d *Data) error {
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	baris1 := raw
	if i := bytes.IndexByte(raw, '\n'); i >= 0 {
		baris1 = raw[:i]
	}
	koma := ','
	if bytes.Count(baris1, []byte(";")) > bytes.Count(baris1, []byte(",")) {
		koma = ';'
	}
	r := csv.NewReader(bytes.NewReader(raw))
	r.Comma = koma
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	rows, err := r.ReadAll()
	if err != nil {
		return errors.New("Berkas CSV tidak dapat dibaca: " + err.Error())
	}
	if len(rows) == 0 {
		return errors.New("Berkas CSV kosong")
	}
	head := map[string]bool{}
	for _, h := range rows[0] {
		head[strings.ToLower(strings.TrimSpace(h))] = true
	}
	switch {
	case head["uid"]:
		k, err := parseKartu(rows)
		if err != nil {
			return err
		}
		d.AdaKartu, d.Kartu = true, k
	case head["nama_ketua"]:
		k, err := parseKelas(rows)
		if err != nil {
			return err
		}
		d.AdaKelas, d.Kelas = true, k
	case head["no_wa"]:
		k, err := parseWali(rows)
		if err != nil {
			return err
		}
		d.AdaWali, d.Wali = true, k
	default:
		return errors.New("Header CSV tidak dikenali: gunakan kolom kelas (nama, nama_ketua, urutan, aktif, no_wa_wali), kartu (uid, kelas, nisn, nama_siswa, aktif), atau wali (nama, no_wa, aktif)")
	}
	return nil
}

// ---------- Tulis ----------

func (d Data) barisKelas() [][]string {
	out := make([][]string, 0, len(d.Kelas))
	for _, k := range d.Kelas {
		out = append(out, []string{k.Nama, k.Ketua, k.Urutan, k.Aktif, k.Wali})
	}
	return out
}

func (d Data) barisWali() [][]string {
	out := make([][]string, 0, len(d.Wali))
	for _, w := range d.Wali {
		out = append(out, []string{w.Nama, w.NoWA, w.Aktif})
	}
	return out
}

func (d Data) barisKartu() [][]string {
	out := make([][]string, 0, len(d.Kartu))
	for _, k := range d.Kartu {
		out = append(out, []string{k.UID, k.Kelas, k.NISN, k.Nama, k.Aktif})
	}
	return out
}

// TulisXLSX: semua sel ditulis sebagai teks berformat "@" agar nol di depan UID/NISN tidak hilang.
// petunjuk (opsional) ditulis sebagai sheet pertama.
func TulisXLSX(path string, d Data, petunjuk []string) error {
	f := excelize.NewFile()
	defer f.Close()
	teks, err := f.NewStyle(&excelize.Style{NumFmt: 49})
	if err != nil {
		return err
	}
	tebal, err := f.NewStyle(&excelize.Style{NumFmt: 49, Font: &excelize.Font{Bold: true}})
	if err != nil {
		return err
	}
	pertama := true
	baru := func(nama string) error {
		if pertama {
			pertama = false
			return f.SetSheetName("Sheet1", nama)
		}
		_, err := f.NewSheet(nama)
		return err
	}
	if len(petunjuk) > 0 {
		if err := baru(SheetPetunjuk); err != nil {
			return err
		}
		_ = f.SetColWidth(SheetPetunjuk, "A", "A", 120)
		for i, ln := range petunjuk {
			cell, _ := excelize.CoordinatesToCellName(1, i+1)
			if err := f.SetCellStr(SheetPetunjuk, cell, ln); err != nil {
				return err
			}
		}
	}
	tabel := func(nama string, head []string, rows [][]string, lebar []float64) error {
		if err := baru(nama); err != nil {
			return err
		}
		akhir, _ := excelize.ColumnNumberToName(len(head))
		if err := f.SetColStyle(nama, "A:"+akhir, teks); err != nil {
			return err
		}
		for c, h := range head {
			cell, _ := excelize.CoordinatesToCellName(c+1, 1)
			if err := f.SetCellStr(nama, cell, h); err != nil {
				return err
			}
			if err := f.SetCellStyle(nama, cell, cell, tebal); err != nil {
				return err
			}
		}
		for r, row := range rows {
			for c, v := range row {
				cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
				if err := f.SetCellStr(nama, cell, v); err != nil {
					return err
				}
			}
		}
		for c, w := range lebar {
			col, _ := excelize.ColumnNumberToName(c + 1)
			_ = f.SetColWidth(nama, col, col, w)
		}
		return nil
	}
	if d.Pengaturan != nil {
		var rows [][]string
		for _, k := range KunciPengaturan {
			rows = append(rows, []string{k, d.Pengaturan[k]})
		}
		if err := tabel(SheetPengaturan, []string{"kunci", "nilai"}, rows, []float64{30, 50}); err != nil {
			return err
		}
	}
	if d.AdaKelas {
		if err := tabel(SheetKelas, headKelas, d.barisKelas(), []float64{18, 30, 10, 8, 18}); err != nil {
			return err
		}
	}
	if d.AdaKartu {
		if err := tabel(SheetKartu, headKartu, d.barisKartu(), []float64{16, 18, 16, 32, 8}); err != nil {
			return err
		}
	}
	if d.AdaWali {
		if err := tabel(SheetWali, headWali, d.barisWali(), []float64{30, 18, 8}); err != nil {
			return err
		}
	}
	if pertama {
		return errors.New("Tidak ada data untuk ditulis")
	}
	if err := f.SaveAs(path); err != nil {
		return err
	}
	_ = os.Chmod(path, 0o600)
	return nil
}

// tulisCSV: UTF-8 dengan BOM agar terbuka benar di Excel.
func tulisCSV(path string, head []string, rows [][]string) error {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	if err := w.Write(head); err != nil {
		return err
	}
	if err := w.WriteAll(rows); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o600)
}

func TulisCSVKelas(path string, d Data) error { return tulisCSV(path, headKelas, d.barisKelas()) }
func TulisCSVKartu(path string, d Data) error { return tulisCSV(path, headKartu, d.barisKartu()) }
func TulisCSVWali(path string, d Data) error  { return tulisCSV(path, headWali, d.barisWali()) }

func TemplateXLSX(path string) error {
	return TulisXLSX(path, Data{AdaKelas: true, AdaKartu: true, AdaWali: true}, PetunjukTemplate)
}

func TemplateCSV(path, jenis string) error {
	switch jenis {
	case "kelas":
		return tulisCSV(path, headKelas, nil)
	case "kartu":
		return tulisCSV(path, headKartu, nil)
	case "wali":
		return tulisCSV(path, headWali, nil)
	}
	return errors.New("Jenis template harus kelas, kartu, atau wali")
}
