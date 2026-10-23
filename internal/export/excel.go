package export

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func ptr[T any](v T) *T { return &v }

func garis() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1},
	}
}

func Excel(l Laporan, path string) error {
	f := excelize.NewFile()
	defer f.Close()
	const sh = "Laporan"
	if err := f.SetSheetName("Sheet1", sh); err != nil {
		return err
	}
	lebar := []float64{6, 10, 28, 28, 28, 13, 14, 14, 10, 10, 32}
	for i, w := range lebar {
		c, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sh, c, c, w)
	}
	st := func(s *excelize.Style) int { id, _ := f.NewStyle(s); return id }
	al := func(h string) *excelize.Alignment {
		return &excelize.Alignment{Horizontal: h, Vertical: "center", WrapText: true}
	}
	abu := excelize.Fill{Type: "pattern", Color: []string{"D9D9D9"}, Pattern: 1}
	pink := excelize.Fill{Type: "pattern", Color: []string{"FFE4E6"}, Pattern: 1}
	merahFont := &excelize.Font{Color: "9F1239"}

	sNama := st(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}, Alignment: al("center")})
	sAlamat := st(&excelize.Style{Font: &excelize.Font{Size: 10}, Alignment: al("center")})
	sGaris := st(&excelize.Style{Border: []excelize.Border{{Type: "bottom", Color: "000000", Style: 5}}})
	sJudul := st(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 12}, Alignment: al("center")})
	sPeriode := st(&excelize.Style{Font: &excelize.Font{Size: 11}, Alignment: al("center")})
	sHead := st(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: abu, Alignment: al("center"), Border: garis()})
	sTengah := st(&excelize.Style{Alignment: al("center"), Border: garis()})
	sKiri := st(&excelize.Style{Alignment: al("left"), Border: garis()})
	sTengahM := st(&excelize.Style{Alignment: al("center"), Border: garis(), Fill: pink, Font: merahFont})
	sKiriM := st(&excelize.Style{Alignment: al("left"), Border: garis(), Fill: pink, Font: merahFont})
	sTtd := st(&excelize.Style{Alignment: al("center")})
	sTtdNama := st(&excelize.Style{Font: &excelize.Font{Bold: true, Underline: "single"}, Alignment: al("center")})

	// Kop surat: baris 1 logo (tengah), 2 nama, 3 alamat, 4 garis, 5 judul, 6 periode
	_ = f.MergeCell(sh, "A1", "K1")
	_ = f.MergeCell(sh, "A2", "K2")
	_ = f.SetCellValue(sh, "A2", l.NamaInstansi)
	_ = f.SetCellStyle(sh, "A2", "K2", sNama)
	_ = f.MergeCell(sh, "A3", "K3")
	_ = f.SetCellValue(sh, "A3", l.Alamat)
	_ = f.SetCellStyle(sh, "A3", "K3", sAlamat)
	_ = f.MergeCell(sh, "A4", "K4")
	_ = f.SetCellStyle(sh, "A4", "K4", sGaris)
	_ = f.MergeCell(sh, "A5", "K5")
	_ = f.SetCellValue(sh, "A5", JudulLaporan)
	_ = f.SetCellStyle(sh, "A5", "K5", sJudul)
	_ = f.MergeCell(sh, "A6", "K6")
	_ = f.SetCellValue(sh, "A6", l.Periode)
	_ = f.SetCellStyle(sh, "A6", "K6", sPeriode)
	_ = f.SetRowHeight(sh, 1, 58)
	_ = f.SetRowHeight(sh, 2, 24)
	_ = f.SetRowHeight(sh, 3, 30)
	_ = f.SetRowHeight(sh, 4, 6)
	_ = f.SetRowHeight(sh, 5, 22)

	if len(l.Logo) > 0 {
		ext := ".png"
		if strings.Contains(l.LogoMime, "jpeg") || strings.Contains(l.LogoMime, "jpg") {
			ext = ".jpg"
		}
		totalPx := 0.0
		for _, w := range lebar {
			totalPx += w*7 + 5
		}
		sc, logoPx := 1.0, 70.0
		if cfg, _, err := image.DecodeConfig(bytes.NewReader(l.Logo)); err == nil && cfg.Width > 0 && cfg.Height > 0 {
			sc = 70.0 / float64(cfg.Height)
			if w := 150.0 / float64(cfg.Width); w < sc {
				sc = w
			}
			logoPx = float64(cfg.Width) * sc
		}
		_ = f.AddPictureFromBytes(sh, "A1", &excelize.Picture{
			Extension: ext, File: l.Logo,
			Format: &excelize.GraphicOptions{ScaleX: sc, ScaleY: sc, Positioning: "oneCell",
				OffsetX: int((totalPx - logoPx) / 2), OffsetY: 2},
		})
	}

	// Header tabel (baris 8)
	const hr = 8
	heads := []string{"No.", "Kelas", "Nama Ketua Murid", "Nama Pengambil", "Nama Pengembali", "Tanggal",
		"Waktu Pengambilan", "Waktu Pengembalian", "Jumlah Ambil", "Jumlah Kembali", "Keterangan"}
	for i, h := range heads {
		c, _ := excelize.CoordinatesToCellName(i+1, hr)
		_ = f.SetCellValue(sh, c, h)
		_ = f.SetCellStyle(sh, c, c, sHead)
	}
	_ = f.SetRowHeight(sh, hr, 32)

	for i, b := range l.Baris {
		r := hr + 1 + i
		vals := []string{strconv.Itoa(i + 1), b.Kelas, b.Ketua, b.Pengambil, b.Pengembali, b.Tanggal, b.WaktuAmbil, b.WaktuKembali,
			b.JumlahAmbil, b.JumlahKembali, b.Keterangan}
		for c, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, r)
			if n, err := strconv.Atoi(v); err == nil && (c == 0 || c == 8 || c == 9) {
				_ = f.SetCellValue(sh, cell, n)
			} else {
				_ = f.SetCellValue(sh, cell, v)
			}
			kiri := c == 2 || c == 3 || c == 4 || c == 10
			s := sTengah
			switch {
			case kiri && b.Merah:
				s = sKiriM
			case kiri:
				s = sKiri
			case b.Merah:
				s = sTengahM
			}
			_ = f.SetCellStyle(sh, cell, cell, s)
		}
		_ = f.SetRowHeight(sh, r, 26)
	}

	// Blok tanda tangan (kolom G-I)
	r0 := hr + len(l.Baris) + 3
	ttd := []struct {
		teks string
		s    int
		h    float64
	}{
		{l.TanggalCetak, sTtd, 18}, {l.Jabatan, sTtd, 18}, {"", sTtd, 60},
		{l.Nama, sTtdNama, 18}, {nipTeks(l.NIP), sTtd, 18},
	}
	for i, t := range ttd {
		r := r0 + i
		a, _ := excelize.CoordinatesToCellName(9, r)
		z, _ := excelize.CoordinatesToCellName(11, r)
		_ = f.MergeCell(sh, a, z)
		_ = f.SetCellValue(sh, a, t.teks)
		_ = f.SetCellStyle(sh, a, z, t.s)
		_ = f.SetRowHeight(sh, r, t.h)
	}

	_ = f.SetPanes(sh, &excelize.Panes{Freeze: true, YSplit: hr, TopLeftCell: "A9", ActivePane: "bottomLeft",
		Selection: []excelize.Selection{{SQRef: "A9", ActiveCell: "A9", Pane: "bottomLeft"}}})
	_ = f.SetSheetProps(sh, &excelize.SheetPropsOptions{FitToPage: ptr(true)})
	_ = f.SetPageLayout(sh, &excelize.PageLayoutOptions{Size: ptr(9), Orientation: ptr("landscape"),
		FitToHeight: ptr(0), FitToWidth: ptr(1)})
	_ = f.SetDefinedName(&excelize.DefinedName{Name: "_xlnm.Print_Titles", RefersTo: fmt.Sprintf("%s!$%d:$%d", sh, hr, hr), Scope: sh})
	return f.SaveAs(path)
}

func nipTeks(n string) string {
	if strings.TrimSpace(n) == "" {
		return ""
	}
	return "NIP. " + n
}
