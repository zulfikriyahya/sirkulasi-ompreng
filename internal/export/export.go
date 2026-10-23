package export

const JudulLaporan = "LAPORAN PENGAMBILAN DAN PENGEMBALIAN OMPRENG MAKANAN BERGIZI GRATIS (MBG)"

type Baris struct {
	Kelas         string `json:"kelas"`
	Ketua         string `json:"ketua"`
	Pengambil     string `json:"pengambil"`
	Pengembali    string `json:"pengembali"`
	Tanggal       string `json:"tanggal"`
	WaktuAmbil    string `json:"waktu_ambil"`
	WaktuKembali  string `json:"waktu_kembali"`
	JumlahAmbil   string `json:"jumlah_ambil"`
	JumlahKembali string `json:"jumlah_kembali"`
	Keterangan    string `json:"keterangan"`
	Merah         bool   `json:"merah"`
}

type Laporan struct {
	NamaInstansi string
	Alamat       string
	Logo         []byte
	LogoMime     string
	Periode      string
	Baris        []Baris
	Jabatan      string
	Nama         string
	NIP          string
	TanggalCetak string // "12 Januari 2026"
}
