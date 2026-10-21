package export

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/go-pdf/fpdf"
)

func PDF(l Laporan, path string) error {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AliasNbPages("")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.CellFormat(0, 5, fmt.Sprintf("Halaman %d dari {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	const bawah = 195.0
	const lh = 4.5
	lebar := []float64{10, 22, 60, 34, 28, 30, 22, 22, 49} // total 277
	heads := []string{"No.", "Kelas", "Nama Ketua Murid", "Tanggal", "Waktu Pengambilan", "Waktu Pengembalian",
		"Jumlah Ambil", "Jumlah Kembali", "Keterangan"}

	pdf.AddPage()

	// Kop surat: logo di tengah, lalu nama, alamat, garis, judul
	y := 8.0
	if len(l.Logo) > 0 {
		t := "PNG"
		if strings.Contains(l.LogoMime, "jpeg") || strings.Contains(l.LogoMime, "jpg") {
			t = "JPG"
		}
		logoH, logoW := 18.0, 18.0
		if cfg, _, err := image.DecodeConfig(bytes.NewReader(l.Logo)); err == nil && cfg.Height > 0 {
			logoW = logoH * float64(cfg.Width) / float64(cfg.Height)
			if logoW > 60 {
				logoH = logoH * 60 / logoW
				logoW = 60
			}
		}
		opt := fpdf.ImageOptions{ImageType: t}
		pdf.RegisterImageOptionsReader("logo", opt, bytes.NewReader(l.Logo))
		pdf.ImageOptions("logo", (297-logoW)/2, y, logoW, logoH, false, opt, 0, "")
		y += logoH + 2
	}
	pdf.SetFont("Helvetica", "B", 15)
	pdf.SetXY(10, y)
	pdf.CellFormat(277, 8, tr(l.NamaInstansi), "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetX(10)
	pdf.MultiCell(277, 5, tr(l.Alamat), "", "C", false)
	yGaris := pdf.GetY() + 1.5
	pdf.SetLineWidth(0.6)
	pdf.Line(10, yGaris, 287, yGaris)
	pdf.SetLineWidth(0.2)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetXY(10, yGaris+4)
	pdf.CellFormat(277, 6, tr(JudulLaporan), "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetX(10)
	pdf.CellFormat(277, 6, tr(l.Periode), "", 1, "C", false, 0, "")
	pdf.SetY(pdf.GetY() + 4)

	kotak := func(x, y, w, h float64, teks, align, gaya string) {
		pdf.Rect(x, y, w, h, gaya)
		ls := pdf.SplitLines([]byte(tr(teks)), w-2)
		ty := y + (h-float64(len(ls))*lh)/2
		for _, ln := range ls {
			pdf.SetXY(x+1, ty)
			pdf.CellFormat(w-2, lh, string(ln), "", 0, align, false, 0, "")
			ty += lh
		}
	}
	head := func() {
		pdf.SetFont("Helvetica", "B", 8.5)
		pdf.SetFillColor(217, 217, 217)
		pdf.SetTextColor(0, 0, 0)
		x, y := 10.0, pdf.GetY()
		for i, h := range heads {
			kotak(x, y, lebar[i], 12, h, "C", "FD")
			x += lebar[i]
		}
		pdf.SetY(y + 12)
	}
	head()

	for i, b := range l.Baris {
		pdf.SetFont("Helvetica", "", 9)
		h := 9.0
		for _, c := range []struct {
			t string
			w float64
		}{{b.Ketua, lebar[2]}, {b.Keterangan, lebar[8]}} {
			n := len(pdf.SplitLines([]byte(tr(c.t)), c.w-2))
			if v := float64(n)*lh + 3; v > h {
				h = v
			}
		}
		if pdf.GetY()+h > bawah {
			pdf.AddPage()
			head()
			pdf.SetFont("Helvetica", "", 9)
		}
		gaya := "D"
		if b.Merah {
			pdf.SetFillColor(255, 228, 230)
			pdf.SetTextColor(159, 18, 57)
			gaya = "FD"
		} else {
			pdf.SetTextColor(0, 0, 0)
		}
		vals := []string{fmt.Sprint(i + 1), b.Kelas, b.Ketua, b.Tanggal, b.WaktuAmbil, b.WaktuKembali,
			b.JumlahAmbil, b.JumlahKembali, b.Keterangan}
		x, y := 10.0, pdf.GetY()
		for c, v := range vals {
			al := "C"
			if c == 2 || c == 8 {
				al = "L"
			}
			kotak(x, y, lebar[c], h, v, al, gaya)
			x += lebar[c]
		}
		pdf.SetY(y + h)
	}
	pdf.SetTextColor(0, 0, 0)

	// Blok tanda tangan
	if pdf.GetY()+55 > bawah {
		pdf.AddPage()
	}
	y = pdf.GetY() + 8
	x, w := 287.0-80, 80.0
	teks := func(s, gaya string, tinggi float64) {
		pdf.SetFont("Helvetica", gaya, 10)
		pdf.SetXY(x, y)
		pdf.CellFormat(w, 5, tr(s), "", 0, "C", false, 0, "")
		y += tinggi
	}
	teks(l.TanggalCetak, "", 6)
	teks(l.Jabatan, "", 24)
	teks(l.Nama, "BU", 6)
	teks(nipTeks(l.NIP), "", 6)

	return pdf.OutputFileAndClose(path)
}
