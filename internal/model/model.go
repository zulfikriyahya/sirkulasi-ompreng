package model

type Kelas struct {
	ID        int64  `json:"id"`
	Nama      string `json:"nama"`
	NamaKetua string `json:"nama_ketua"`
	Urutan    int    `json:"urutan"`
	Aktif     bool   `json:"aktif"`
}

type Kartu struct {
	UID     string `json:"uid"`
	KelasID int64  `json:"kelas_id"`
	Aktif   bool   `json:"aktif"`
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
// Aksi: AMBIL | KEMBALI | KONFIRMASI_KEMBALI | PESAN | ABAIKAN
// Tipe (untuk warna/bunyi): ok | err | info
type TapResult struct {
	Aksi        string `json:"aksi"`
	Tipe        string `json:"tipe"`
	Pesan       string `json:"pesan"`
	KelasID     int64  `json:"kelas_id"`
	NamaKelas   string `json:"nama_kelas"`
	NamaKetua   string `json:"nama_ketua"`
	JumlahAmbil int    `json:"jumlah_ambil"`
	MenitSejak  int    `json:"menit_sejak"`
}

type BarisHari struct {
	KelasID       int64  `json:"kelas_id"`
	Nama          string `json:"nama"`
	NamaKetua     string `json:"nama_ketua"`
	Status        string `json:"status"` // BELUM_AMBIL | AMBIL | SELESAI | BELUM_KEMBALI | TIDAK_AMBIL
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

type Transaksi struct {
	ID            int64  `json:"id"`
	Tanggal       string `json:"tanggal"`
	KelasID       int64  `json:"kelas_id"`
	NamaKelas     string `json:"nama_kelas"`
	NamaKetua     string `json:"nama_ketua"`
	WaktuAmbil    string `json:"waktu_ambil"`
	WaktuKembali  string `json:"waktu_kembali"`
	JumlahAmbil   *int   `json:"jumlah_ambil"`
	JumlahKembali *int   `json:"jumlah_kembali"`
	Status        string `json:"status"`
	Catatan       string `json:"catatan"`
	Keterangan    string `json:"keterangan"`
	Merah         bool   `json:"merah"`
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
