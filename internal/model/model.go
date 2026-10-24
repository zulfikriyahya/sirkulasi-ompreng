package model

type Kelas struct {
	ID          int64  `json:"id"`
	Nama        string `json:"nama"`
	NamaKetua   string `json:"nama_ketua"`
	Urutan      int    `json:"urutan"`
	Aktif       bool   `json:"aktif"`
	WaliKelasID int64  `json:"wali_kelas_id"` // 0 = belum ada
	NamaWali    string `json:"nama_wali"`
}

type WaliKelas struct {
	ID          int64  `json:"id"`
	Nama        string `json:"nama"`
	NoWA        string `json:"no_wa"` // format 62xxxxxxxxxx
	Aktif       bool   `json:"aktif"`
	DaftarKelas string `json:"daftar_kelas"` // kelas aktif yang diampu, dipisah koma
}

type Kartu struct {
	UID       string `json:"uid"`
	KelasID   int64  `json:"kelas_id"`
	Aktif     bool   `json:"aktif"`
	NISN      string `json:"nisn"`
	NamaSiswa string `json:"nama_siswa"`
}

type Pengaturan struct {
	NamaInstansi           string `json:"nama_instansi"`
	AlamatInstansi         string `json:"alamat_instansi"`
	PenanggungJawab        string `json:"penanggung_jawab"`
	NIPPenanggungJawab     string `json:"nip_penanggung_jawab"`
	JabatanPenanggungJawab string `json:"jabatan_penanggung_jawab"`
	JedaMinimumMenit       int    `json:"jeda_minimum_menit"`
	HIDAmbangMs            int    `json:"hid_ambang_ms"`
	PortSerial             string `json:"port_serial"`
}

// TapResult: apa yang harus dilakukan UI setelah tap.
// Aksi: AMBIL | KEMBALI | KEMBALI_SUSULAN | KONFIRMASI_KEMBALI | PESAN | ABAIKAN
// Tipe (untuk warna/bunyi): ok | err | info
// Sisa hanya terisi pada KEMBALI_SUSULAN (jumlah_ambil - total sudah kembali).
type TapResult struct {
	Aksi        string `json:"aksi"`
	Tipe        string `json:"tipe"`
	Pesan       string `json:"pesan"`
	KelasID     int64  `json:"kelas_id"`
	NamaKelas   string `json:"nama_kelas"`
	NamaKetua   string `json:"nama_ketua"`
	JumlahAmbil int    `json:"jumlah_ambil"`
	Sisa        int    `json:"sisa"`
	MenitSejak  int    `json:"menit_sejak"`
	NISN        string `json:"nisn"`
	NamaSiswa   string `json:"nama_siswa"`
}

type BarisHari struct {
	KelasID       int64  `json:"kelas_id"`
	Nama          string `json:"nama"`
	NamaKetua     string `json:"nama_ketua"`
	Status        string `json:"status"` // BELUM_AMBIL | AMBIL | KURANG | SELESAI | BELUM_KEMBALI | TIDAK_AMBIL
	JumlahAmbil   *int   `json:"jumlah_ambil"`
	JumlahKembali *int   `json:"jumlah_kembali"`
	WaktuAmbil    string `json:"waktu_ambil"`
	WaktuKembali  string `json:"waktu_kembali"`
	Keterangan    string `json:"keterangan"`
	Merah         bool   `json:"merah"`
}

type StatusHari struct {
	Tanggal      string      `json:"tanggal"`
	TanggalIndo  string      `json:"tanggal_indo"`
	Ditutup      bool        `json:"ditutup"`
	SemuaSelesai bool        `json:"semua_selesai"`
	BolehCetak   bool        `json:"boleh_cetak"`
	Baris        []BarisHari `json:"baris"`
}

type FilterTransaksi struct {
	Dari    string `json:"dari"`
	Sampai  string `json:"sampai"`
	KelasID int64  `json:"kelas_id"` // 0 = semua
	Status  string `json:"status"`   // "" = semua
}

// Pengembalian: satu kali pengembalian (awal atau susulan) pada sebuah transaksi.
type Pengembalian struct {
	Waktu     string `json:"waktu"`
	Jumlah    int    `json:"jumlah"`
	NISN      string `json:"nisn"`
	NamaSiswa string `json:"nama_siswa"`
}

type Transaksi struct {
	ID            int64          `json:"id"`
	Tanggal       string         `json:"tanggal"`
	KelasID       int64          `json:"kelas_id"`
	NamaKelas     string         `json:"nama_kelas"`
	NamaKetua     string         `json:"nama_ketua"`
	WaktuAmbil    string         `json:"waktu_ambil"`
	WaktuKembali  string         `json:"waktu_kembali"`
	JumlahAmbil   *int           `json:"jumlah_ambil"`
	JumlahKembali *int           `json:"jumlah_kembali"`
	Status        string         `json:"status"`
	Catatan       string         `json:"catatan"`
	Keterangan    string         `json:"keterangan"`
	Merah         bool           `json:"merah"`
	NamaAmbil     string         `json:"nama_ambil"`
	NamaKembali   string         `json:"nama_kembali"`
	Susulan       bool           `json:"susulan"`
	Pengembalian  []Pengembalian `json:"pengembalian"`
}

type EditTransaksi struct {
	ID            int64  `json:"id"`
	JumlahAmbil   *int   `json:"jumlah_ambil"`
	JumlahKembali *int   `json:"jumlah_kembali"`
	WaktuAmbil    string `json:"waktu_ambil"`
	WaktuKembali  string `json:"waktu_kembali"`
	Status        string `json:"status"`
	Catatan       string `json:"catatan"`
	Alasan        string `json:"alasan"`
}

type FilterAudit struct {
	Dari    string `json:"dari"`
	Sampai  string `json:"sampai"`
	Entitas string `json:"entitas"`
	Limit   int    `json:"limit"`
}

type Audit struct {
	ID        int64  `json:"id"`
	Waktu     string `json:"waktu"`
	Aksi      string `json:"aksi"`
	Entitas   string `json:"entitas"`
	EntitasID string `json:"entitas_id"`
	NilaiLama string `json:"nilai_lama"`
	NilaiBaru string `json:"nilai_baru"`
	Alasan    string `json:"alasan"`
}

// ---------- Sinkronisasi master data ----------

// OpsiImpor: Mode GABUNG (bawaan) atau GANTI. Hash diisi dari HasilPratinjau.Hash saat menerapkan.
type OpsiImpor struct {
	Mode              string `json:"mode"`
	BuatKelasOtomatis bool   `json:"buat_kelas_otomatis"`
	KonfirmasiGanti   bool   `json:"konfirmasi_ganti"`
	Alasan            string `json:"alasan"`
	Hash              string `json:"hash"`
}

type BarisPratinjau struct {
	Entitas string `json:"entitas"` // kelas | kartu | pengaturan
	Baris   int    `json:"baris"`   // nomor baris di berkas (0 = bukan dari baris berkas)
	Kunci   string `json:"kunci"`
	Aksi    string `json:"aksi"` // BARU | UBAH | HAPUS | NONAKTIF | OTOMATIS | GALAT | PERINGATAN
	Pesan   string `json:"pesan"`
}

type RingkasanImpor struct {
	KelasBaru      int `json:"kelas_baru"`
	KelasUbah      int `json:"kelas_ubah"`
	KelasSama      int `json:"kelas_sama"`
	KelasNonaktif  int `json:"kelas_nonaktif"`
	KelasOtomatis  int `json:"kelas_otomatis"`
	KartuBaru      int `json:"kartu_baru"`
	KartuUbah      int `json:"kartu_ubah"`
	KartuSama      int `json:"kartu_sama"`
	KartuHapus     int `json:"kartu_hapus"`
	WaliBaru       int `json:"wali_baru"`
	WaliUbah       int `json:"wali_ubah"`
	WaliSama       int `json:"wali_sama"`
	WaliNonaktif   int `json:"wali_nonaktif"`
	PengaturanUbah int `json:"pengaturan_ubah"`
	Galat          int `json:"galat"`
	Peringatan     int `json:"peringatan"`
}

type HasilPratinjau struct {
	Berkas    string           `json:"berkas"`
	Mode      string           `json:"mode"`
	Ringkasan RingkasanImpor   `json:"ringkasan"`
	Baris     []BarisPratinjau `json:"baris"`
	Terpotong bool             `json:"terpotong"`
	Hash      string           `json:"hash"`
}

type HasilImpor struct {
	Ringkasan RingkasanImpor `json:"ringkasan"`
	Pesan     string         `json:"pesan"`
}

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

type WAConfig struct {
	Aktif           bool   `json:"aktif"`
	BaseURL         string `json:"base_url"`
	KeyID           string `json:"key_id"`
	Secret          string `json:"secret"`     // hanya masuk; tidak pernah dikirim balik ke UI
	SecretAda       bool   `json:"secret_ada"` // true bila secret sudah tersimpan
	TemplateAmbil   string `json:"template_ambil"`
	TemplateKembali string `json:"template_kembali"`
}

type WAAntrean struct {
	ID            int64  `json:"id"`
	Jenis         string `json:"jenis"`
	NamaWali      string `json:"nama_wali"`
	NoWA          string `json:"no_wa"`
	Status        string `json:"status"` // ANTRE | TERKIRIM | GAGAL
	StatusLayanan string `json:"status_layanan"`
	Percobaan     int    `json:"percobaan"`
	Galat         string `json:"galat"`
	Dibuat        string `json:"dibuat"`
	CobaLagi      string `json:"coba_lagi"`
	Terkirim      string `json:"terkirim"`
}

type WARingkasan struct {
	Antre    int `json:"antre"`
	Terkirim int `json:"terkirim"`
	Gagal    int `json:"gagal"`
}

type WAUji struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}
