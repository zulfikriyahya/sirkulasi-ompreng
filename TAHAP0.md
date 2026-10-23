zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ tree
.
├── build
│   ├── appicon.png
│   ├── bin
│   │   ├── mbg-amd64-installer.exe
│   │   └── mbg.exe
│   ├── linux
│   │   ├── 99-mbg-reader.rules
│   │   ├── mbg.desktop
│   │   ├── mbg.png
│   │   └── postinst.sh
│   └── windows
│       ├── icon.ico
│       ├── info.json
│       ├── installer
│       │   ├── project.nsi
│       │   ├── tmp
│       │   │   └── MicrosoftEdgeWebview2Setup.exe
│       │   └── wails_tools.nsh
│       └── wails.exe.manifest
├── dist
│   ├── mbg-0.1.1-1.x86_64.rpm
│   ├── mbg_0.1.1_amd64.deb
│   ├── mbg-amd64-installer.exe
│   └── mbg-linux-amd64
├── draft.md
├── firmware
│   └── esp32c3_rc522
│       └── esp32c3_rc522.ino
├── frontend
│   ├── css
│   │   └── style.css
│   ├── index.html
│   ├── js
│   │   ├── api.js
│   │   ├── dashboard.js
│   │   ├── ekspor.js
│   │   ├── main.js
│   │   ├── master.js
│   │   ├── riwayat.js
│   │   └── tap.js
│   └── wailsjs
│       ├── go
│       │   ├── models.ts
│       │   └── service
│       │       ├── Service.d.ts
│       │       └── Service.js
│       └── runtime
│           ├── package.json
│           ├── runtime.d.ts
│           └── runtime.js
├── go.mod
├── go.sum
├── icon.jpeg
├── internal
│   ├── db
│   │   └── db.go
│   ├── export
│   │   ├── excel.go
│   │   ├── export.go
│   │   └── pdf.go
│   ├── model
│   │   └── model.go
│   ├── reader
│   │   ├── serial.go
│   │   ├── serial_test.go
│   │   ├── uid.go
│   │   └── uid_test.go
│   ├── service
│   │   ├── ekspor.go
│   │   ├── ekspor_test.go
│   │   ├── hari.go
│   │   ├── master.go
│   │   ├── reader.go
│   │   ├── riwayat.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── tahap5_test.go
│   │   └── transaksi.go
│   └── timeutil
│       └── timeutil.go
├── LICENSE
├── main.go
├── mbg.db
├── nfpm.yaml
├── README.md
├── SPEC.md
├── TAHAP0.md
├── TAHAP1.md
├── TAHAP2.md
└── wails.json

24 directories, 64 files
zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ cat README.md 
# MBG - Pencatatan Ompreng

Aplikasi desktop offline untuk mencatat pengambilan dan pengembalian ompreng Makanan Bergizi Gratis (MBG) memakai kartu RFID. Siswa tap kartu, operator mengisi jumlah, nama pengambil dan pengembali tercatat otomatis, dan laporan diekspor ke Excel dan PDF.

Dibuat dengan Go, SQLite, dan Wails v2. Berjalan di Windows 10+ dan Linux (Debian/Ubuntu, Red Hat/Fedora).

![Icon](icon.jpeg)

## Fitur

- Tap kartu: pengambilan pada tap pertama, pengembalian pada tap berikutnya
- Nama siswa pemilik kartu tampil di popup dan tercatat sebagai pengambil/pengembali
- Reader serial (ESP32-C3 + RC522) dan reader HID (keyboard emulator), terdeteksi otomatis
- Perlindungan tap ganda: penguncian popup, cooldown, jeda minimum ambil ke kembali
- Validasi: jumlah kembali tidak boleh melebihi jumlah ambil
- Tutup hari dengan catatan, dan buka kembali bila perlu
- Edit dan hapus transaksi dengan alasan wajib, tercatat di audit log
- Ekspor laporan formal ke Excel dan PDF (A4 landscape, kop dan logo instansi, kolom Pengambil dan Pengembali, baris merah untuk kekurangan)
- Data lokal di SQLite, tanpa login dan tanpa internet

## Instalasi

Unduh dari [halaman Releases](https://github.com/zulfikriyahya/sirkulasi-ompreng/releases/latest).

| OS | Berkas | Cara pasang |
|---|---|---|
| Windows 10/11 | `mbg-amd64-installer.exe` | Jalankan installer |
| Debian/Ubuntu | `mbg_0.1.1_amd64.deb` | `sudo apt install ./mbg_0.1.1_amd64.deb` |
| Fedora/RHEL 10 | `mbg-0.1.1-1.x86_64.rpm` | `sudo dnf install ./mbg-0.1.1-1.x86_64.rpm` |
| Linux (portabel) | `mbg-linux-amd64` | `chmod +x mbg-linux-amd64 && ./mbg-linux-amd64` |

Paket Linux memasang dependensi (GTK3 dan WebKitGTK 4.1) otomatis dan menyertakan aturan udev agar reader bisa diakses tanpa grup `dialout`. Cabut dan colok ulang reader setelah instalasi.

Windows memakai WebView2 (sudah ada di Windows 10/11 terbaru; installer menyertakannya bila belum ada).

## Penggunaan

1. **Master Data > Instansi**: isi nama, alamat, penanggung jawab, NIP, jabatan, dan unggah logo.
2. **Master Data > Kelas & Ketua**: tambahkan semua kelas.
3. **Master Data > Kartu RFID**: pilih kelas, tap kartu (atau ketik UID), isi NISN dan nama siswa (opsional), lalu Daftarkan.
4. **Beranda**: siswa tap kartu, operator mengisi jumlah, Submit. Tap lagi saat mengembalikan. Pengambil dan pengembali boleh berbeda orang.
5. Setelah semua kelas selesai (atau hari ditutup), klik **Cetak PDF**. Rentang tanggal tersedia di tab **Ekspor**.

Kartu tanpa NISN dan nama (kartu kelas) tetap berfungsi; kolom pengambil/pengembali menampilkan "-".

## Data

| OS | Lokasi |
|---|---|
| Linux | `~/.local/share/mbg/mbg.db` |
| Windows | `%APPDATA%\MBG\mbg.db` |

Data tidak terhapus saat update atau uninstall. Kolom baru ditambahkan otomatis saat aplikasi dibuka, jadi database versi lama tetap kompatibel.

Untuk mengosongkan, tutup aplikasi lalu hapus `mbg.db`, `mbg.db-wal`, dan `mbg.db-shm`.

### Cadangan dan pemindahan database

Database memakai mode WAL, jadi jangan menyalin dengan `cp` saat aplikasi terbuka. Tutup aplikasi, lalu:

```bash
sqlite3 ~/.local/share/mbg/mbg.db ".backup /lokasi/cadangan/mbg.db"
```

Saat menimpa database di komputer lain, hapus dulu `mbg.db-wal` dan `mbg.db-shm` di folder tujuan.

### Privasi

`mbg.db` berisi NISN dan nama siswa. Jangan memasukkannya ke repositori atau rilis publik. File `.db`, `.csv`, dan `.tsv` data siswa sebaiknya masuk `.gitignore`.

## Firmware reader

Sketch ada di `firmware/esp32c3_rc522/`. Wiring ESP32-C3 Super Mini ke RC522:

| RC522 | ESP32-C3 |
|---|---|
| SDA (SS) | GPIO 7 |
| SCK | GPIO 4 |
| MOSI | GPIO 6 |
| MISO | GPIO 5 |
| RST | GPIO 3 |
| 3.3V | 3V3 (jangan 5V) |
| GND | GND |

Upload lewat Arduino IDE: board **ESP32C3 Dev Module**, **USB CDC On Boot: Enabled**, library **MFRC522**. Setiap tap mengirim 10 digit desimal (4 byte pertama UID, little-endian) dengan baud 115200.

## Pengembangan

Prasyarat: Go, Node.js/npm, Wails CLI, dan di Linux `libgtk-3-dev` serta `libwebkit2gtk-4.1-dev`.

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
go vet ./... && go test ./...
wails dev
```

## Build

Linux dibangun di Linux (CGO). Windows bisa dibangun dari Linux (tanpa CGO).

Prasyarat build (Debian/Ubuntu):

```bash
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev nsis imagemagick
go install github.com/wailsapp/wails/v2/cmd/wails@latest
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
```

Ikon (sekali saja, dari `icon.jpeg`):

```bash
mkdir -p build/linux
convert icon.jpeg -gravity center -background white -resize 1024x1024 -extent 1024x1024 build/appicon.png
convert build/appicon.png -resize 512x512 build/linux/mbg.png
```

Build semua target:

```bash
go vet ./... && go test ./... || exit 1
rm -rf dist && mkdir -p dist
export VERSION=0.1.1

wails build -clean
cp build/bin/mbg dist/mbg-linux-amd64

wails build -clean -platform windows/amd64 -nsis -webview2 embed
cp build/bin/mbg-amd64-installer.exe dist/

nfpm package -p deb -t dist/
nfpm package -p rpm -t dist/
```

Hasil di `dist/`: `mbg-linux-amd64`, `mbg_0.1.1_amd64.deb`, `mbg-0.1.1-1.x86_64.rpm`, `mbg-amd64-installer.exe`.

Catatan: `nfpm.yaml` mengambil binary Linux dari `dist/mbg-linux-amd64`, jadi build Linux harus dijalankan lebih dulu. `.rpm` dari Debian 13 menautkan glibc baru, sehingga tidak cocok untuk RHEL/Rocky 9; bangun di container Rocky 9 untuk target itu.

## Rilis

Prasyarat: [GitHub CLI](https://cli.github.com) (`sudo apt install gh`, lalu `gh auth login`).

1. Naikkan nomor versi (contoh ke `0.1.2`):

   ```bash
   export VERSION=0.1.2
   sed -i "s/0\.1\.1-stable/${VERSION}-stable/" internal/service/service.go
   sed -i "s/^version: .*/version: ${VERSION}/" nfpm.yaml
   sed -i "s/0\.1\.1/${VERSION}/g" README.md
   ```

2. Jalankan perintah di bagian **Build** (semua target) dengan `VERSION` yang sama.

3. Commit, beri tag, dan push:

   ```bash
   git add -A
   git commit -m "Rilis v${VERSION}"
   git tag -a "v${VERSION}" -m "Versi ${VERSION}"
   git push origin main --tags
   ```

4. Buat release di GitHub dan unggah semua berkas:

   ```bash
   gh release create "v${VERSION}" dist/* \
     --title "v${VERSION}" \
     --notes "Catatan perubahan versi ${VERSION}."
   ```

   Untuk catatan otomatis dari riwayat commit, ganti `--notes "..."` dengan `--generate-notes`.

### Merebuild versi yang sama

Jika perlu membangun ulang tanpa menaikkan versi:

```bash
git add -A && git commit -m "Perbaikan v${VERSION}"
git push origin main
git tag -fa "v${VERSION}" -m "Versi ${VERSION}"
git push origin "v${VERSION}" --force
gh release upload "v${VERSION}" dist/* --clobber
```

Verifikasi:

```bash
gh release view "v${VERSION}" --json assets --jq '.assets[] | "\(.name) \(.size)"'
git ls-remote --tags origin "v${VERSION}*"
```

Hasil `v${VERSION}^{}` harus menunjuk ke commit `HEAD`, dan ukuran aset sama dengan file di `dist/`.

## Struktur

```
main.go            entry Wails
internal/          db, model, service, reader, export, timeutil
frontend/          HTML/CSS/JS vanilla
firmware/          sketch ESP32-C3 + RC522
build/             ikon, aturan udev, berkas paket
dist/              hasil build (tidak masuk git)
```

## Lisensi

[MIT](LICENSE)

---

built with ♥ by [Yahya Zulfikri](https://zulfikriyahya.my.id)zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ cat SPEC.md 
# KONTEKS PROYEK

Aplikasi desktop "MBG - Pencatatan Ompreng" (Go + Wails v2, SQLite via modernc.org/sqlite, frontend vanilla JS). Repo: github.com/zulfikriyahya/sirkulasi-ompreng, versi saat ini v0.1.1. Struktur: internal/{db,model,service,export,reader,timeutil}, frontend/{index.html,js/*.js}.

Skema relevan:
- kelas(id, nama, nama_ketua, urutan, aktif)
- kartu(uid PK, kelas_id, aktif, nisn, nama_siswa)  -- kartu RFID terikat ke kelas; nisn/nama_siswa menandai pemilik kartu
- transaksi(id, tanggal, kelas_id, nama_kelas, nama_ketua, waktu_ambil, waktu_kembali, jumlah_ambil, jumlah_kembali, status, catatan, nisn_ambil, nama_ambil, nisn_kembali, nama_kembali, UNIQUE(tanggal, kelas_id))
- status: AMBIL | SELESAI | BELUM_KEMBALI | TIDAK_AMBIL
- Alur tap: ProsesTap -> aksi AMBIL | KEMBALI | KONFIRMASI_KEMBALI | PESAN | ABAIKAN; lalu SubmitAmbil/SubmitKembali.
- Migrasi kolom otomatis lewat tambahKolom() di internal/db/db.go. Semua perubahan data wajib masuk audit_log lewat s.audit().

ATURAN UMUM
- Pertahankan gaya kode, nama Indonesia, dan pola transaksi/audit yang sudah ada.
- Jangan merusak DB lama: semua perubahan skema lewat migrasi idempoten.
- Setiap fitur wajib punya tes di internal/service/*_test.go (go test ./... harus lulus, go vet bersih, gofmt bersih).
- Binding Wails berubah bila signature berubah; jalankan wails build untuk regenerasi.
- Kerjakan per fitur, satu commit per fitur. Sebelum menulis kode, tunjukkan rencana singkat (file yang diubah, perubahan skema) lalu tunggu persetujuan.

---

# FITUR 1: Import dan Export data master (Excel/CSV)

Tujuan: data master bisa dipindah/disinkronkan antar komputer.

Cakupan data: pengaturan instansi (tanpa logo opsional), kelas (nama, ketua, urutan, aktif), kartu (uid, kelas, nisn, nama_siswa, aktif).

Kebutuhan:
1. Tab Master Data mendapat sub-tab "Sinkronisasi" dengan tombol: Ekspor Master (Excel .xlsx), Ekspor Master (CSV), Impor Master.
2. Ekspor Excel: satu berkas dengan sheet "Kelas" dan "Kartu" (dan "Pengaturan"). CSV: pilihan jenis data (kelas atau kartu), UTF-8 dengan BOM agar terbuka benar di Excel.
3. Impor: dialog pilih berkas .xlsx/.csv, lalu TAHAP PRATINJAU sebelum menulis: tampilkan jumlah baris baru, diubah, dilewati, dan galat per baris (UID bukan angka 1-10 digit, kelas tidak dikenal, UID ganda dalam berkas, dst).
4. Mode impor dipilih pengguna: "Gabung" (tambah/perbarui, tidak menghapus) atau "Ganti" (hapus data kartu/kelas yang tidak ada di berkas, dengan konfirmasi eksplisit dan alasan).
5. Kunci pencocokan: kelas = nama; kartu = uid. Kartu yang kelasnya belum ada: tawarkan membuat kelas otomatis atau tolak.
6. Seluruh impor dalam satu transaksi database (gagal sebagian = batal semua) dan satu entri audit_log ringkasan plus entri per perubahan kartu.
7. Jangan mengimpor/mengekspor riwayat transaksi pada fitur ini (opsional fase 2: ekspor transaksi untuk cadangan).
8. Keamanan data pribadi: peringatan di UI bahwa berkas berisi NISN dan nama siswa; jangan menulis berkas ke lokasi repo.
9. Berikan juga template impor kosong yang bisa diunduh.

Tes: round-trip (ekspor lalu impor ke DB kosong menghasilkan data identik), UID ganda, kelas tidak dikenal, nol di depan NISN/UID tidak hilang (simpan sebagai teks di Excel), mode Ganti.

---

# FITUR 2: Pembaruan aplikasi via GitHub Release

Tujuan: aplikasi dapat memeriksa dan memasang versi baru dari github.com/zulfikriyahya/sirkulasi-ompreng/releases.

Kebutuhan:
1. Versi aplikasi dibaca dari satu sumber (variabel build via -ldflags "-X"), bukan string tertulis di service.go. Perbarui skrip rilis agar menyuntikkan VERSION.
2. Cek pembaruan: panggil GitHub API releases/latest (tanpa token, dengan timeout dan User-Agent), bandingkan semver dengan versi terpasang. Dijalankan saat aplikasi dibuka (di latar belakang, gagal diam-diam bila offline) dan lewat tombol "Periksa pembaruan" di Master Data > Instansi.
3. Tampilkan catatan rilis dan tombol "Unduh dan pasang".
4. Windows: unduh aset mbg-amd64-installer.exe ke folder sementara, verifikasi ukuran (dan checksum SHA256 bila rilis menyertakan berkas checksums.txt), tampilkan progres, lalu jalankan installer dan tutup aplikasi. Linux: unduh .deb/.rpm sesuai distribusi dan tampilkan instruksi (pemasangan butuh sudo; jangan menjalankan sudo otomatis).
5. Cadangkan mbg.db otomatis (dengan .backup, bukan cp) ke folder data sebelum memulai pembaruan; simpan 5 cadangan terakhir.
6. Jangan pernah menurunkan versi otomatis; abaikan pre-release kecuali opsi dicentang.
7. Tambahkan langkah rilis: hasilkan checksums.txt (sha256sum dist/*) dan unggah bersama aset.
8. Semua panggilan jaringan dapat dimatikan lewat pengaturan ("Periksa pembaruan otomatis").

Keamanan: hanya unduh dari domain github.com/objects.githubusercontent.com milik repo tersebut, wajib HTTPS, tolak bila checksum tidak cocok.

Tes: perbandingan semver (0.1.1 < 0.1.2 < 0.2.0 < 1.0.0, abaikan awalan "v"), parsing respons API memakai httptest, checksum cocok/tidak cocok, offline tidak menimbulkan galat di UI.

---

# FITUR 3: Tap kedua untuk pengembalian susulan

Aturan bisnis:
- Saat pengembalian, jika jumlah_kembali < jumlah_ambil, ompreng dianggap MENYUSUL dikembalikan: transaksi tetap dapat di-tap lagi pada hari yang sama untuk pengembalian susulan.
- Jika jumlah_kembali == jumlah_ambil (lengkap), transaksi SELESAI final dan tap berikutnya DITOLAK dengan pesan "Kelas X sudah mengembalikan ompreng dengan lengkap".

Kebutuhan:
1. Status baru "KURANG" (atau pakai SELESAI + flag), yaitu sudah ada pengembalian tetapi belum lengkap. Pilih desain yang paling sedikit memecah logika lama (Keterangan(), BarisHari, dashboard, ekspor, validasiEdit, tutup hari/buka hari) dan jelaskan alasannya sebelum mengerjakan.
2. Skema: simpan riwayat pengembalian parsial dalam tabel baru pengembalian(id, transaksi_id, waktu, jumlah, nisn, nama_siswa); jumlah_kembali di transaksi = total akumulasi. Migrasi data lama: satu baris pengembalian per transaksi SELESAI yang sudah ada.
3. ProsesTap: bila status KURANG, tampilkan popup "Pengembalian susulan" dengan info sisa (jumlah_ambil - total kembali), input jumlah susulan (1..sisa), nama pengembali dari kartu. Tidak ada peringatan jeda minimum untuk susulan.
4. SubmitKembali: tambah pengembalian; bila total == jumlah_ambil ubah status menjadi SELESAI; bila masih kurang tetap KURANG. Total tidak boleh melebihi jumlah_ambil.
5. Tap pada transaksi SELESAI (lengkap): tolak dengan pesan jelas, tanpa membuka popup. Perlu berlaku juga untuk kartu siswa berbeda di kelas yang sama.
6. Cooldown 30 detik per kelas dan debounce 2 detik tetap berlaku; pastikan tap susulan tidak terblokir cooldown dari tap sebelumnya secara tidak wajar (dokumentasikan keputusan).
7. Keterangan laporan: tampil "Kurang N Ompreng" selama belum lengkap; setelah susulan lengkap tampil "Lengkap (susulan)" dan waktu kembali memakai waktu pengembalian terakhir. Kolom pengembali menampilkan semua nama pengembali (dipisah koma) bila lebih dari satu.
8. Tutup hari: transaksi KURANG yang belum lengkap menjadi BELUM_KEMBALI dengan keterangan sisa; Buka hari mengembalikannya ke KURANG. Hari yang sudah ditutup tetap menolak tap.
9. UI: dashboard menampilkan status "Kurang" (merah) dengan sisa; riwayat menampilkan rincian pengembalian parsial; dialog Ubah transaksi tetap valid (validasiEdit disesuaikan).
10. Audit log: setiap pengembalian susulan tercatat dengan nilai lama/baru.

Tes: ambil 10 -> kembali 7 -> tap lagi (popup susulan, sisa 3) -> kembali 3 -> SELESAI; tap lagi ditolak; susulan melebihi sisa ditolak; kembali lengkap langsung lalu tap ditolak; tutup hari saat KURANG; buka hari; ekspor menampilkan keterangan dan pengembali ganda dengan benar; migrasi DB lama aman.

---

URUTAN KERJA YANG DIMINTA
1. Fitur 3 (paling dekat ke logika inti), 2. Fitur 1, 3. Fitur 2.
Setiap fitur: rencana -> persetujuan -> implementasi -> tes -> ringkasan perubahan -> saran nomor versi (0.1.2, 0.2.0, ...).zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$

selesaikan dan finalkan pertahap, jika kamu perlu melihat file berikan dalam kelompok perintah cat

Siap, saya kerjakan bertahap mulai dari Fitur 3. Sebelum menulis rencana, saya perlu melihat kode yang akan disentuh. Jalankan tiga kelompok perintah ini dari `~/ZEDLABS/transaksi-mbg/mbg` dan tempel hasilnya (boleh sekaligus atau per kelompok).

**Kelompok A: logika inti (paling penting)**

```bash
for f in internal/model/model.go internal/db/db.go internal/service/service.go internal/service/transaksi.go internal/service/hari.go internal/service/riwayat.go; do echo "=== $f ==="; cat "$f"; done
```

**Kelompok B: ekspor, reader, dan frontend**

```bash
for f in internal/service/ekspor.go internal/service/reader.go internal/export/export.go internal/export/excel.go internal/export/pdf.go frontend/js/tap.js frontend/js/dashboard.js frontend/js/riwayat.js; do echo "=== $f ==="; cat "$f"; done
```

**Kelompok C: pola tes yang sudah ada**

```bash
for f in internal/service/service_test.go internal/service/tahap5_test.go; do echo "=== $f ==="; cat "$f"; done
wc -l internal/service/*.go internal/export/*.go frontend/js/*.js
```

Kalau ada yang terlalu panjang, kirim Kelompok A dulu. Dari situ saya sudah bisa menyusun rencana Fitur 3, yaitu berkas yang diubah, migrasi skema, pilihan desain status `KURANG` atau `SELESAI` + flag beserta alasannya, dan keputusan cooldown. Kelompok B dan C saya butuhkan sebelum implementasi.

Sesuai SPEC, saya akan menunggu persetujuan Anda atas rencana sebelum menulis kode.

zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ for f in internal/model/model.go internal/db/db.go internal/service/service.go internal/service/transaksi.go internal/service/hari.go internal/service/riwayat.go; do echo "=== $f ==="; cat "$f"; done
=== internal/model/model.go ===
package model

type Kelas struct {
	ID        int64  `json:"id"`
	Nama      string `json:"nama"`
	NamaKetua string `json:"nama_ketua"`
	Urutan    int    `json:"urutan"`
	Aktif     bool   `json:"aktif"`
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
	NISN        string `json:"nisn"`
	NamaSiswa   string `json:"nama_siswa"`
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
	NamaAmbil     string `json:"nama_ambil"`
	NamaKembali   string `json:"nama_kembali"`
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
=== internal/db/db.go ===
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite"
)

// DataDir: %APPDATA%\MBG (Windows), $XDG_DATA_HOME/mbg atau ~/.local/share/mbg (Linux).
func DataDir() string {
	if runtime.GOOS == "windows" {
		if a := os.Getenv("APPDATA"); a != "" {
			return filepath.Join(a, "MBG")
		}
	}
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "mbg")
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".local", "share", "mbg")
}

func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", filepath.ToSlash(path))
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1) // satu penulis, hindari SQLITE_BUSY
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	if err := Migrate(conn); err != nil {
		return nil, fmt.Errorf("migrasi: %w", err)
	}
	return conn, nil
}

func Migrate(conn *sql.DB) error {
	if _, err := conn.Exec(schema); err != nil {
		return err
	}
	for _, k := range []struct{ tabel, kolom string }{
		{"kartu", "nisn"}, {"kartu", "nama_siswa"},
		{"transaksi", "nisn_ambil"}, {"transaksi", "nama_ambil"},
		{"transaksi", "nisn_kembali"}, {"transaksi", "nama_kembali"},
	} {
		if err := tambahKolom(conn, k.tabel, k.kolom); err != nil {
			return err
		}
	}
	_, err := conn.Exec(`INSERT OR IGNORE INTO pengaturan (id, jabatan_penanggung_jawab) VALUES (1, 'Penanggung Jawab MBG')`)
	return err
}

func tambahKolom(conn *sql.DB, tabel, kolom string) error {
	rows, err := conn.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, tabel))
	if err != nil {
		return err
	}
	ada := false
	for rows.Next() {
		var cid, notnull, pk int
		var nama, tipe string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &nama, &tipe, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		if nama == kolom {
			ada = true
		}
	}
	rows.Close()
	if ada {
		return nil
	}
	_, err = conn.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s TEXT`, tabel, kolom))
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS pengaturan (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  nama_instansi TEXT, alamat_instansi TEXT,
  logo BLOB, logo_mime TEXT,
  penanggung_jawab TEXT, nip_penanggung_jawab TEXT,
  jabatan_penanggung_jawab TEXT,
  jeda_minimum_menit INTEGER NOT NULL DEFAULT 10,
  hid_ambang_ms INTEGER NOT NULL DEFAULT 50,
  port_serial TEXT
);
CREATE TABLE IF NOT EXISTS kelas (
  id INTEGER PRIMARY KEY,
  nama TEXT NOT NULL UNIQUE,
  nama_ketua TEXT NOT NULL,
  urutan INTEGER NOT NULL DEFAULT 0,
  aktif INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS kartu (
  uid TEXT PRIMARY KEY,
  kelas_id INTEGER NOT NULL REFERENCES kelas(id),
  aktif INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS transaksi (
  id INTEGER PRIMARY KEY,
  tanggal TEXT NOT NULL,
  kelas_id INTEGER NOT NULL REFERENCES kelas(id),
  nama_kelas TEXT NOT NULL,
  nama_ketua TEXT NOT NULL,
  waktu_ambil TEXT,
  waktu_kembali TEXT,
  jumlah_ambil INTEGER,
  jumlah_kembali INTEGER,
  status TEXT NOT NULL,
  catatan TEXT,
  UNIQUE (tanggal, kelas_id)
);
CREATE TABLE IF NOT EXISTS hari_ditutup (
  tanggal TEXT PRIMARY KEY,
  waktu_tutup TEXT NOT NULL,
  catatan TEXT
);
CREATE TABLE IF NOT EXISTS audit_log (
  id INTEGER PRIMARY KEY,
  waktu TEXT NOT NULL,
  aksi TEXT NOT NULL,
  entitas TEXT NOT NULL,
  entitas_id TEXT,
  nilai_lama TEXT,
  nilai_baru TEXT,
  alasan TEXT
);
`
=== internal/service/service.go ===
package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"mbg/internal/timeutil"
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

func (s *Service) InfoSistem() InfoSistem { return InfoSistem{Versi: "0.1.1-stable", DataDir: s.dir} }

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
=== internal/service/transaksi.go ===
package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mbg/internal/model"
	"mbg/internal/reader"
	"mbg/internal/timeutil"
)

func itoa(i int64) string { return strconv.FormatInt(i, 10) }

const (
	debounceTap   = 2 * time.Second
	cooldownKelas = 30 * time.Second
)

func (s *Service) jedaMenit() int {
	var m int
	if err := s.db.QueryRow(`SELECT jeda_minimum_menit FROM pengaturan WHERE id=1`).Scan(&m); err != nil {
		return 10
	}
	return m
}

func pesan(tipe, teks string) model.TapResult {
	return model.TapResult{Aksi: "PESAN", Tipe: tipe, Pesan: teks}
}

// ProsesTap: satu pintu untuk semua sumber (serial, HID, manual, simulator).
func (s *Service) ProsesTap(rawUID string) model.TapResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	uid, ok := reader.NormalisasiUID(rawUID)
	if !ok {
		return pesan("err", "UID tidak valid")
	}
	now := s.now()

	if s.pending != nil {
		return pesan("info", "Selesaikan input kelas "+s.pending.Nama+" dulu")
	}
	if uid == s.lastUID && now.Sub(s.lastUIDAt) < debounceTap {
		return model.TapResult{Aksi: "ABAIKAN"}
	}
	s.lastUID, s.lastUIDAt = uid, now

	var k model.Kelas
	var kartuAktif, kelasAktif int
	var nisn, namaSiswa string
	err := s.db.QueryRow(`SELECT k.id,k.nama,k.nama_ketua,c.aktif,k.aktif,COALESCE(c.nisn,''),COALESCE(c.nama_siswa,'') FROM kartu c JOIN kelas k ON k.id=c.kelas_id WHERE c.uid=?`, uid).
		Scan(&k.ID, &k.Nama, &k.NamaKetua, &kartuAktif, &kelasAktif, &nisn, &namaSiswa)
	if err != nil || kartuAktif != 1 || kelasAktif != 1 {
		return pesan("err", "Kartu tidak terdaftar")
	}

	tgl := timeutil.Tanggal(now)
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n)
	if n > 0 {
		return pesan("err", "Hari ini sudah ditutup")
	}

	if t, ok := s.lastSubmit[k.ID]; ok && now.Sub(t) < cooldownKelas {
		return pesan("info", "Baru saja tercatat")
	}

	base := model.TapResult{KelasID: k.ID, NamaKelas: k.Nama, NamaKetua: k.NamaKetua, Tipe: "ok", NISN: nisn, NamaSiswa: namaSiswa}
	var status, waktuAmbil string
	var jAmbil int
	err = s.db.QueryRow(`SELECT status,COALESCE(waktu_ambil,''),COALESCE(jumlah_ambil,0) FROM transaksi WHERE tanggal=? AND kelas_id=?`, tgl, k.ID).
		Scan(&status, &waktuAmbil, &jAmbil)
	if errors.Is(err, sql.ErrNoRows) {
		s.pending = &pendingInput{k.ID, k.Nama}
		s.siswaNISN, s.siswaNama = nisn, namaSiswa
		base.Aksi = "AMBIL"
		return base
	}
	if err != nil {
		return pesan("err", "Kesalahan database")
	}

	if status != "AMBIL" {
		return pesan("info", "Kelas "+k.Nama+" sudah selesai hari ini")
	}

	base.JumlahAmbil = jAmbil
	s.pending = &pendingInput{k.ID, k.Nama}
	s.siswaNISN, s.siswaNama = nisn, namaSiswa
	if t, err := time.ParseInLocation("2006-01-02 15:04", tgl+" "+waktuAmbil, timeutil.WIB); err == nil {
		menit := int(now.Sub(t).Minutes())
		if menit < s.jedaMenit() {
			base.Aksi = "KONFIRMASI_KEMBALI"
			base.MenitSejak = menit
			return base
		}
	}
	base.Aksi = "KEMBALI"
	return base
}

// BatalInput: dipanggil UI saat popup/konfirmasi dibatalkan, melepas penguncian.
func (s *Service) BatalInput() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending = nil
	s.siswaNISN, s.siswaNama = "", ""
}

func (s *Service) SubmitAmbil(kelasID int64, jumlah int) error {
	if jumlah < 1 {
		return errors.New("Jumlah ambil harus bilangan bulat minimal 1")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var nama, ketua string
	if err := tx.QueryRow(`SELECT nama,nama_ketua FROM kelas WHERE id=?`, kelasID).Scan(&nama, &ketua); err != nil {
		return errors.New("Kelas tidak ditemukan")
	}
	res, err := tx.Exec(`INSERT INTO transaksi (tanggal,kelas_id,nama_kelas,nama_ketua,waktu_ambil,jumlah_ambil,status,nisn_ambil,nama_ambil) VALUES (?,?,?,?,?,?,'AMBIL',?,?)`,
		timeutil.Tanggal(now), kelasID, nama, ketua, timeutil.Jam(now), jumlah, s.siswaNISN, s.siswaNama)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("Kelas ini sudah tercatat mengambil hari ini")
		}
		return err
	}
	id, _ := res.LastInsertId()
	if err := s.audit(tx, now, "TAMBAH", "transaksi", itoa(id), nil,
		map[string]any{"kelas": nama, "jumlah_ambil": jumlah, "status": "AMBIL", "pengambil": s.siswaNama}, ""); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.lastSubmit[kelasID] = now
	s.pending = nil
	s.siswaNISN, s.siswaNama = "", ""
	return nil
}

func (s *Service) SubmitKembali(kelasID int64, jumlah int) error {
	if jumlah < 0 {
		return errors.New("Jumlah kembali tidak boleh negatif")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int64
	var status, nama string
	var jAmbil int
	err = tx.QueryRow(`SELECT id,status,nama_kelas,COALESCE(jumlah_ambil,0) FROM transaksi WHERE tanggal=? AND kelas_id=?`,
		timeutil.Tanggal(now), kelasID).Scan(&id, &status, &nama, &jAmbil)
	if err != nil {
		return errors.New("Belum ada pengambilan hari ini untuk kelas ini")
	}
	if status != "AMBIL" {
		return errors.New("Transaksi kelas ini sudah selesai")
	}
	if jumlah > jAmbil {
		return fmt.Errorf("Jumlah kembali (%d) tidak boleh melebihi jumlah ambil (%d)", jumlah, jAmbil)
	}
	if _, err := tx.Exec(`UPDATE transaksi SET waktu_kembali=?,jumlah_kembali=?,status='SELESAI',nisn_kembali=?,nama_kembali=? WHERE id=?`,
		timeutil.Jam(now), jumlah, s.siswaNISN, s.siswaNama, id); err != nil {
		return err
	}
	if err := s.audit(tx, now, "UBAH", "transaksi", itoa(id),
		map[string]any{"status": "AMBIL"},
		map[string]any{"status": "SELESAI", "jumlah_kembali": jumlah, "pengembali": s.siswaNama}, ""); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.lastSubmit[kelasID] = now
	s.pending = nil
	s.siswaNISN, s.siswaNama = "", ""
	return nil
}

// StatusHariIni: data dashboard (keterangan dihitung saat tampil).
func (s *Service) StatusHariIni() (model.StatusHari, error) {
	now := s.now()
	tgl := timeutil.Tanggal(now)
	out := model.StatusHari{Tanggal: tgl, TanggalIndo: timeutil.HariTanggalIndo(now), Baris: []model.BarisHari{}}

	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n)
	out.Ditutup = n > 0

	rows, err := s.db.Query(`SELECT k.id,k.nama,k.nama_ketua,t.status,t.jumlah_ambil,t.jumlah_kembali,
		COALESCE(t.waktu_ambil,''),COALESCE(t.waktu_kembali,'')
		FROM kelas k LEFT JOIN transaksi t ON t.kelas_id=k.id AND t.tanggal=?
		WHERE k.aktif=1 ORDER BY k.urutan,k.nama`, tgl)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	semua := true
	for rows.Next() {
		var b model.BarisHari
		var st sql.NullString
		var ja, jk sql.NullInt64
		if err := rows.Scan(&b.KelasID, &b.Nama, &b.NamaKetua, &st, &ja, &jk, &b.WaktuAmbil, &b.WaktuKembali); err != nil {
			return out, err
		}
		b.Status = "BELUM_AMBIL"
		if st.Valid {
			b.Status = st.String
		}
		if ja.Valid {
			v := int(ja.Int64)
			b.JumlahAmbil = &v
		}
		if jk.Valid {
			v := int(jk.Int64)
			b.JumlahKembali = &v
		}
		b.Keterangan, b.Merah = Keterangan(b.Status, b.JumlahAmbil, b.JumlahKembali, "")
		if b.Status != "SELESAI" {
			semua = false
		}
		out.Baris = append(out.Baris, b)
	}
	out.SemuaSelesai = semua && len(out.Baris) > 0
	out.BolehCetak = out.Ditutup || out.SemuaSelesai
	return out, rows.Err()
}

// Keterangan: aturan 4.3. Dipakai dashboard dan nanti ekspor.
func Keterangan(status string, ambil, kembali *int, catatan string) (string, bool) {
	var ket string
	merah := false
	switch status {
	case "SELESAI":
		if ambil != nil && kembali != nil && *kembali < *ambil {
			ket, merah = fmt.Sprintf("Kurang %d Ompreng", *ambil-*kembali), true
		} else {
			ket = "Lengkap"
		}
	case "BELUM_KEMBALI":
		ket, merah = "Belum kembali", true
	case "TIDAK_AMBIL":
		ket, merah = "Tidak mengambil", true
	}
	if catatan != "" {
		if ket != "" {
			ket += " - "
		}
		ket += catatan
	}
	return ket, merah
}
=== internal/service/hari.go ===
package service

import (
	"errors"
	"strings"

	"mbg/internal/timeutil"
)

// TutupHari: kelas AMBIL -> BELUM_KEMBALI, kelas aktif tanpa transaksi -> TIDAK_AMBIL.
func (s *Service) TutupHari(catatan string) error {
	catatan = strings.TrimSpace(catatan)
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tgl := timeutil.Tanggal(now)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return errors.New("Hari ini sudah ditutup")
	}
	r1, err := tx.Exec(`UPDATE transaksi SET status='BELUM_KEMBALI' WHERE tanggal=? AND status='AMBIL'`, tgl)
	if err != nil {
		return err
	}
	belum, _ := r1.RowsAffected()
	r2, err := tx.Exec(`INSERT INTO transaksi (tanggal,kelas_id,nama_kelas,nama_ketua,status)
		SELECT ?,k.id,k.nama,k.nama_ketua,'TIDAK_AMBIL' FROM kelas k
		WHERE k.aktif=1 AND NOT EXISTS (SELECT 1 FROM transaksi t WHERE t.tanggal=? AND t.kelas_id=k.id)`, tgl, tgl)
	if err != nil {
		return err
	}
	tidak, _ := r2.RowsAffected()
	if _, err := tx.Exec(`INSERT INTO hari_ditutup (tanggal,waktu_tutup,catatan) VALUES (?,?,?)`,
		tgl, timeutil.Jam(now), catatan); err != nil {
		return err
	}
	if err := s.audit(tx, now, "TUTUP_HARI", "hari", tgl, nil,
		map[string]any{"belum_kembali": belum, "tidak_ambil": tidak}, catatan); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.pending = nil
	return nil
}

// BukaHari: hanya hari ini. Status otomatis dikembalikan (BELUM_KEMBALI -> AMBIL, baris TIDAK_AMBIL otomatis dihapus)
// supaya tap bisa dilanjutkan. Koreksi hari lampau dilakukan lewat Riwayat (edit/hapus).
func (s *Service) BukaHari(tanggal, alasan string) error {
	if strings.TrimSpace(alasan) == "" {
		return errors.New("Alasan wajib diisi")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if tanggal != timeutil.Tanggal(now) {
		return errors.New("Hanya hari ini yang bisa dibuka kembali; koreksi hari sebelumnya lewat menu Riwayat")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tanggal).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return errors.New("Hari ini tidak dalam keadaan ditutup")
	}
	r1, err := tx.Exec(`UPDATE transaksi SET status='AMBIL' WHERE tanggal=? AND status='BELUM_KEMBALI' AND jumlah_kembali IS NULL`, tanggal)
	if err != nil {
		return err
	}
	r2, err := tx.Exec(`DELETE FROM transaksi WHERE tanggal=? AND status='TIDAK_AMBIL' AND jumlah_ambil IS NULL`, tanggal)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM hari_ditutup WHERE tanggal=?`, tanggal); err != nil {
		return err
	}
	a, _ := r1.RowsAffected()
	b, _ := r2.RowsAffected()
	if err := s.audit(tx, now, "BUKA_HARI", "hari", tanggal, nil,
		map[string]any{"dikembalikan_ke_ambil": a, "baris_tidak_ambil_dihapus": b}, alasan); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// ApakahBolehCetak: hari ditutup, atau semua kelas aktif berstatus SELESAI pada tanggal itu.
func (s *Service) ApakahBolehCetak(tanggal string) (bool, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tanggal).Scan(&n); err != nil {
		return false, err
	}
	if n > 0 {
		return true, nil
	}
	var aktif, selesai int
	err := s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(CASE WHEN t.status='SELESAI' THEN 1 ELSE 0 END),0)
		FROM kelas k LEFT JOIN transaksi t ON t.kelas_id=k.id AND t.tanggal=? WHERE k.aktif=1`, tanggal).Scan(&aktif, &selesai)
	if err != nil {
		return false, err
	}
	return aktif > 0 && aktif == selesai, nil
}
=== internal/service/riwayat.go ===
package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"mbg/internal/model"
)

const kolomTrx = `id,tanggal,kelas_id,nama_kelas,nama_ketua,COALESCE(waktu_ambil,''),COALESCE(waktu_kembali,''),jumlah_ambil,jumlah_kembali,status,COALESCE(catatan,''),COALESCE(nama_ambil,''),COALESCE(nama_kembali,'')`

type pemindai interface{ Scan(dest ...any) error }

func scanTrx(r pemindai) (model.Transaksi, error) {
	var t model.Transaksi
	var ja, jk sql.NullInt64
	if err := r.Scan(&t.ID, &t.Tanggal, &t.KelasID, &t.NamaKelas, &t.NamaKetua, &t.WaktuAmbil, &t.WaktuKembali,
		&ja, &jk, &t.Status, &t.Catatan, &t.NamaAmbil, &t.NamaKembali); err != nil {
		return t, err
	}
	if ja.Valid {
		v := int(ja.Int64)
		t.JumlahAmbil = &v
	}
	if jk.Valid {
		v := int(jk.Int64)
		t.JumlahKembali = &v
	}
	t.Keterangan, t.Merah = Keterangan(t.Status, t.JumlahAmbil, t.JumlahKembali, t.Catatan)
	return t, nil
}

func nilInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func nilStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func jamValid(s string) bool { _, err := time.Parse("15:04", s); return err == nil }

func validasiEdit(e model.EditTransaksi) error {
	switch e.Status {
	case "AMBIL", "SELESAI", "BELUM_KEMBALI", "TIDAK_AMBIL":
	default:
		return errors.New("Status tidak valid")
	}
	if e.Status == "TIDAK_AMBIL" {
		return nil
	}
	if e.JumlahAmbil == nil || *e.JumlahAmbil < 1 {
		return errors.New("Jumlah ambil harus bilangan bulat minimal 1")
	}
	if !jamValid(e.WaktuAmbil) {
		return errors.New("Waktu ambil harus berformat HH:MM")
	}
	if e.Status == "SELESAI" {
		if e.JumlahKembali == nil {
			return errors.New("Status Selesai wajib punya jumlah kembali")
		}
		if *e.JumlahKembali < 0 {
			return errors.New("Jumlah kembali tidak boleh negatif")
		}
		if *e.JumlahKembali > *e.JumlahAmbil {
			return fmt.Errorf("Jumlah kembali (%d) tidak boleh melebihi jumlah ambil (%d)", *e.JumlahKembali, *e.JumlahAmbil)
		}
		if !jamValid(e.WaktuKembali) {
			return errors.New("Waktu kembali harus berformat HH:MM")
		}
	}
	return nil
}

func (s *Service) ListTransaksi(f model.FilterTransaksi) ([]model.Transaksi, error) {
	q := `SELECT ` + kolomTrx + ` FROM transaksi WHERE 1=1`
	var args []any
	if f.Dari != "" {
		q += ` AND tanggal>=?`
		args = append(args, f.Dari)
	}
	if f.Sampai != "" {
		q += ` AND tanggal<=?`
		args = append(args, f.Sampai)
	}
	if f.KelasID > 0 {
		q += ` AND kelas_id=?`
		args = append(args, f.KelasID)
	}
	if f.Status != "" {
		q += ` AND status=?`
		args = append(args, f.Status)
	}
	q += ` ORDER BY tanggal DESC, (SELECT urutan FROM kelas k WHERE k.id=transaksi.kelas_id), nama_kelas`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Transaksi{}
	for rows.Next() {
		t, err := scanTrx(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Service) UbahTransaksi(e model.EditTransaksi) error {
	if strings.TrimSpace(e.Alasan) == "" {
		return errors.New("Alasan perubahan wajib diisi")
	}
	if err := validasiEdit(e); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	lama, err := scanTrx(tx.QueryRow(`SELECT `+kolomTrx+` FROM transaksi WHERE id=?`, e.ID))
	if err != nil {
		return errors.New("Transaksi tidak ditemukan")
	}
	ja, jk, wa, wk := e.JumlahAmbil, e.JumlahKembali, e.WaktuAmbil, e.WaktuKembali
	switch e.Status {
	case "AMBIL", "BELUM_KEMBALI":
		jk, wk = nil, ""
	case "TIDAK_AMBIL":
		ja, jk, wa, wk = nil, nil, "", ""
	}
	catatan := strings.TrimSpace(e.Catatan)
	if _, err := tx.Exec(`UPDATE transaksi SET status=?,jumlah_ambil=?,jumlah_kembali=?,waktu_ambil=?,waktu_kembali=?,catatan=? WHERE id=?`,
		e.Status, nilInt(ja), nilInt(jk), nilStr(wa), nilStr(wk), nilStr(catatan), e.ID); err != nil {
		return err
	}
	baru := map[string]any{"status": e.Status, "jumlah_ambil": nilInt(ja), "jumlah_kembali": nilInt(jk),
		"waktu_ambil": wa, "waktu_kembali": wk, "catatan": catatan}
	if err := s.audit(tx, s.now(), "UBAH", "transaksi", itoa(e.ID), lama, baru, strings.TrimSpace(e.Alasan)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Service) HapusTransaksi(id int64, alasan string) error {
	alasan = strings.TrimSpace(alasan)
	if alasan == "" {
		return errors.New("Alasan penghapusan wajib diisi")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	lama, err := scanTrx(tx.QueryRow(`SELECT `+kolomTrx+` FROM transaksi WHERE id=?`, id))
	if err != nil {
		return errors.New("Transaksi tidak ditemukan")
	}
	if _, err := tx.Exec(`DELETE FROM transaksi WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "HAPUS", "transaksi", itoa(id), lama, nil, alasan); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	delete(s.lastSubmit, lama.KelasID)
	return nil
}

func (s *Service) ListAudit(f model.FilterAudit) ([]model.Audit, error) {
	q := `SELECT id,waktu,aksi,entitas,COALESCE(entitas_id,''),COALESCE(nilai_lama,''),COALESCE(nilai_baru,''),COALESCE(alasan,'') FROM audit_log WHERE 1=1`
	var args []any
	if f.Dari != "" {
		q += ` AND substr(waktu,1,10)>=?`
		args = append(args, f.Dari)
	}
	if f.Sampai != "" {
		q += ` AND substr(waktu,1,10)<=?`
		args = append(args, f.Sampai)
	}
	if f.Entitas != "" {
		q += ` AND entitas=?`
		args = append(args, f.Entitas)
	}
	lim := f.Limit
	if lim <= 0 || lim > 1000 {
		lim = 200
	}
	q += fmt.Sprintf(` ORDER BY id DESC LIMIT %d`, lim)
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Audit{}
	for rows.Next() {
		var a model.Audit
		if err := rows.Scan(&a.ID, &a.Waktu, &a.Aksi, &a.Entitas, &a.EntitasID, &a.NilaiLama, &a.NilaiBaru, &a.Alasan); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ for f in internal/service/ekspor.go internal/service/reader.go internal/export/export.go internal/export/excel.go internal/export/pdf.go frontend/js/tap.js frontend/js/dashboard.js frontend/js/riwayat.js; do echo "=== $f ==="; cat "$f"; done
=== internal/service/ekspor.go ===
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
=== internal/service/reader.go ===
package service

import (
	"sync"

	"mbg/internal/reader"
)

type readerState struct {
	mu sync.Mutex
	st reader.Status
}

func (s *Service) SetStatusReader(st reader.Status) {
	s.rdr.mu.Lock()
	s.rdr.st = st
	s.rdr.mu.Unlock()
}

// StatusReader: dipanggil frontend saat start (event pertama bisa terlewat).
func (s *Service) StatusReader() reader.Status {
	s.rdr.mu.Lock()
	defer s.rdr.mu.Unlock()
	return s.rdr.st
}
=== internal/export/export.go ===
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
=== internal/export/excel.go ===
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
=== internal/export/pdf.go ===
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
	lebar := []float64{8, 16, 38, 38, 38, 22, 20, 22, 16, 16, 43} // total 277
	heads := []string{"No.", "Kelas", "Nama Ketua Murid", "Nama Pengambil", "Nama Pengembali", "Tanggal",
		"Waktu Pengambilan", "Waktu Pengembalian", "Jumlah Ambil", "Jumlah Kembali", "Keterangan"}

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
		}{{b.Ketua, lebar[2]}, {b.Pengambil, lebar[3]}, {b.Pengembali, lebar[4]}, {b.Keterangan, lebar[10]}} {
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
		vals := []string{fmt.Sprint(i + 1), b.Kelas, b.Ketua, b.Pengambil, b.Pengembali, b.Tanggal, b.WaktuAmbil, b.WaktuKembali,
			b.JumlahAmbil, b.JumlahKembali, b.Keterangan}
		x, y := 10.0, pdf.GetY()
		for c, v := range vals {
			al := "C"
			if c == 2 || c == 3 || c == 4 || c == 10 {
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
=== frontend/js/tap.js ===
// Penangkap reader HID (keyboard emulator) berdasarkan pola ketikan cepat.
const Tap = {
    ambangMs: 50,
    handler: null,         // fungsi(uid) saat tap valid; di-override oleh tab Kartu
    buf: "", last: 0,
};

document.addEventListener("keydown", e => {
    const now = performance.now();
    if (e.key === "Enter") {
        if (Tap.buf.length >= 8 && Tap.buf.length <= 10) {
            e.preventDefault(); e.stopPropagation();
            const uid = Tap.buf; Tap.buf = "";
            if (Tap.handler) Tap.handler(uid);
            return;
        }
        Tap.buf = ""; return;
    }
    if (/^\d$/.test(e.key)) {
        if (now - Tap.last > Tap.ambangMs) Tap.buf = "";
        Tap.buf += e.key; Tap.last = now;
    } else if (e.key.length === 1) {
        Tap.buf = "";
    }
}, true);

// Simulasi HID: kirim event keydown cepat ke dokumen
function simulasiHID(uid) {
    for (const ch of uid) document.dispatchEvent(new KeyboardEvent("keydown", { key: ch, bubbles: true, cancelable: true }));
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }));
}=== frontend/js/dashboard.js ===
const Dash = {
    async muat() {
        const st = await API.StatusHariIni();
        document.getElementById("info-tanggal").textContent = st.tanggal_indo + (st.ditutup ? " (DITUTUP)" : "");
        const tb = document.querySelector("#tabel-hari tbody");
        tb.innerHTML = "";
        const label = { BELUM_AMBIL: "Belum ambil", AMBIL: "Sudah ambil", SELESAI: "Selesai", BELUM_KEMBALI: "Belum kembali", TIDAK_AMBIL: "Tidak mengambil" };
        let selesai = 0;
        st.baris.forEach((b, i) => {
            if (b.status === "SELESAI") selesai++;
            const tr = document.createElement("tr");
            if (b.merah) tr.className = "merah"; else if (b.status === "SELESAI") tr.className = "selesai";
            const v = x => (x === null || x === undefined) ? "-" : x;
            [i + 1, b.nama, b.nama_ketua, label[b.status] || b.status, v(b.jumlah_ambil), v(b.jumlah_kembali),
            b.waktu_ambil || "-", b.waktu_kembali || "-", b.keterangan].forEach(t => {
                const td = document.createElement("td"); td.textContent = t; tr.appendChild(td);
            });
            tb.appendChild(tr);
        });
        document.getElementById("ringkasan").textContent = `${selesai} dari ${st.baris.length} kelas selesai.`;
        Dash.terapkanHari(st);
    },

    // ---- popup ----
    tutupModal() { document.getElementById("modal").className = "hidden"; },
    bukaModal(html) { const m = document.getElementById("modal"); document.getElementById("modal-isi").innerHTML = html; m.className = ""; },

    popupJumlah(r, mode) {
        const ambil = mode === "AMBIL";
        this.bukaModal(`
      <h2>${ambil ? "Pengambilan" : "Pengembalian"} - ${esc(r.nama_kelas)}</h2>
      <p>Ketua: ${esc(r.nama_ketua)}${ambil ? "" : ` &middot; Diambil: <b>${r.jumlah_ambil}</b> ompreng`}</p>
      <p>${ambil ? "Pengambil" : "Pengembali"}: <b>${esc(r.nama_siswa || "(kartu tanpa nama)")}</b>${r.nisn ? " &middot; NISN " + esc(r.nisn) : ""}</p>
      <input id="in-jumlah" type="number" min="${ambil ? 1 : 0}" ${ambil ? "" : `max="${r.jumlah_ambil}"`} placeholder="Jumlah ${ambil ? "diambil" : "dikembalikan"}">
      <div class="galat" id="in-galat"></div>
      <div class="aksi"><button id="in-batal">Batal</button><button id="in-ok">Submit</button></div>`);
        const inp = document.getElementById("in-jumlah"); inp.focus();
        const kirim = async () => {
            const val = inp.value.trim();
            if (!/^\d+$/.test(val)) { document.getElementById("in-galat").textContent = "Isi dengan bilangan bulat"; return; }
            try {
                if (ambil) await API.SubmitAmbil(r.kelas_id, parseInt(val, 10));
                else await API.SubmitKembali(r.kelas_id, parseInt(val, 10));
                this.tutupModal();
                toast(`${r.nama_kelas}: ${ambil ? "pengambilan" : "pengembalian"} tercatat`, "ok");
                this.muat();
            } catch (err) {
                document.getElementById("in-galat").textContent = String(err.message || err);
                bunyi("err");
            }
        };
        document.getElementById("in-ok").onclick = kirim;
        inp.onkeydown = e => { if (e.key === "Enter") kirim(); };
        document.getElementById("in-batal").onclick = async () => { await API.BatalInput(); this.tutupModal(); };
    },

    popupKonfirmasi(r) {
        this.bukaModal(`
      <h2>Konfirmasi</h2>
      <p>Kelas <b>${esc(r.nama_kelas)}</b> baru mengambil ${r.menit_sejak} menit lalu. Pengembali: <b>${esc(r.nama_siswa || "-")}</b>. Tetap proses pengembalian?</p>
      <div class="aksi"><button id="kf-tidak">Tidak</button><button id="kf-ya">Ya</button></div>`);
        document.getElementById("kf-tidak").onclick = async () => { await API.BatalInput(); this.tutupModal(); };
        document.getElementById("kf-ya").onclick = () => this.popupJumlah(r, "KEMBALI");
    },
};

function esc(s) { return String(s).replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c])); }

async function prosesTap(uid) {
    try {
        const r = await API.ProsesTap(uid);
        if (r.aksi === "ABAIKAN") return;
        if (r.aksi === "PESAN") return toast(r.pesan, r.tipe);
        bunyi("ok");
        if (r.nama_siswa) toast(r.nama_siswa + " (" + r.nama_kelas + ")", "info");
        if (r.aksi === "AMBIL") Dash.popupJumlah(r, "AMBIL");
        else if (r.aksi === "KEMBALI") Dash.popupJumlah(r, "KEMBALI");
        else if (r.aksi === "KONFIRMASI_KEMBALI") Dash.popupKonfirmasi(r);
    } catch (err) { toast(String(err.message || err), "err"); }
}=== frontend/js/riwayat.js ===
const _r = id => document.getElementById(id);

// Tombol cetak/tutup/buka di Beranda mengikuti status hari
Dash.terapkanHari = function (st) {
    _r("btn-cetak").disabled = !st.boleh_cetak;
    _r("btn-tutup").classList.toggle("hidden", st.ditutup || st.semua_selesai);
    _r("btn-buka").classList.toggle("hidden", !st.ditutup);
};

const Riwayat = {
    pilih(s) {
        _r("r-trx").classList.toggle("hidden", s !== "trx");
        _r("r-audit").classList.toggle("hidden", s !== "audit");
        _r("r-btn-trx").classList.toggle("aktif", s === "trx");
        _r("r-btn-audit").classList.toggle("aktif", s === "audit");
        if (s === "trx") this.muatTrx(); else this.muatAudit();
    },

    async muat() {
        const st = await API.StatusHariIni();
        for (const id of ["rf-dari", "rf-sampai", "ra-dari", "ra-sampai"]) if (!_r(id).value) _r(id).value = st.tanggal;
        const ks = await API.ListKelas(true);
        const sel = _r("rf-kelas"), cur = sel.value;
        sel.innerHTML = '<option value="0">Semua kelas</option>' + ks.map(k => `<option value="${k.id}">${esc(k.nama)}</option>`).join("");
        sel.value = cur || "0";
        this.muatTrx();
    },

    async muatTrx() {
        try {
            const rows = await API.ListTransaksi({
                dari: _r("rf-dari").value, sampai: _r("rf-sampai").value,
                kelas_id: parseInt(_r("rf-kelas").value || "0", 10), status: _r("rf-status").value,
            });
            const tb = document.querySelector("#tabel-riwayat tbody");
            tb.innerHTML = "";
            const v = x => (x === null || x === undefined) ? "-" : x;
            rows.forEach(r => {
                const tr = document.createElement("tr");
                if (r.merah) tr.className = "merah";
                [r.tanggal, r.nama_kelas, r.nama_ketua, r.nama_ambil || "-", r.nama_kembali || "-", r.status, v(r.jumlah_ambil), v(r.jumlah_kembali),
                (r.waktu_ambil || "-") + " / " + (r.waktu_kembali || "-"), r.keterangan].forEach(t => {
                    const td = document.createElement("td"); td.textContent = t; tr.appendChild(td);
                });
                const td = document.createElement("td"); td.className = "aksi-sel";
                const wadah = document.createElement("div"); wadah.className = "aksi-baris";
                const be = document.createElement("button"); be.textContent = "Ubah"; be.onclick = () => this.dialogUbah(r);
                const bh = document.createElement("button"); bh.textContent = "Hapus"; bh.className = "bahaya"; bh.onclick = () => this.dialogHapus(r);
                wadah.append(be, bh); td.appendChild(wadah); tr.appendChild(td);
                tb.appendChild(tr);
            });
            if (!rows.length) tb.innerHTML = '<tr><td colspan="11">Tidak ada data.</td></tr>';
        } catch (e) { toast(String(e.message || e), "err"); }
    },

    async muatAudit() {
        try {
            const rows = await API.ListAudit({ dari: _r("ra-dari").value, sampai: _r("ra-sampai").value, entitas: _r("ra-entitas").value, limit: 300 });
            const tb = document.querySelector("#tabel-audit tbody");
            tb.innerHTML = "";
            rows.forEach(a => {
                const tr = document.createElement("tr");
                [a.waktu.replace("T", " ").slice(0, 19), a.aksi, a.entitas, a.entitas_id, a.nilai_lama, a.nilai_baru, a.alasan].forEach(t => {
                    const td = document.createElement("td"); td.textContent = t; td.title = t; tr.appendChild(td);
                });
                tb.appendChild(tr);
            });
            if (!rows.length) tb.innerHTML = '<tr><td colspan="7">Tidak ada data.</td></tr>';
        } catch (e) { toast(String(e.message || e), "err"); }
    },

    dialogUbah(r) {
        const n = x => (x === null || x === undefined) ? "" : x;
        const opsi = ["AMBIL", "SELESAI", "BELUM_KEMBALI", "TIDAK_AMBIL"].map(s => `<option ${s === r.status ? "selected" : ""}>${s}</option>`).join("");
        Dash.bukaModal(`
          <h2>Ubah transaksi - ${esc(r.nama_kelas)} (${esc(r.tanggal)})</h2>
          <label>Status <select id="e-status">${opsi}</select></label>
          <label>Jumlah ambil <input id="e-ja" type="number" min="1" value="${n(r.jumlah_ambil)}"></label>
          <label>Jumlah kembali <input id="e-jk" type="number" min="0" value="${n(r.jumlah_kembali)}"></label>
          <label>Waktu ambil (HH:MM) <input id="e-wa" maxlength="5" value="${esc(r.waktu_ambil)}"></label>
          <label>Waktu kembali (HH:MM) <input id="e-wk" maxlength="5" value="${esc(r.waktu_kembali)}"></label>
          <label>Catatan <input id="e-cat" value="${esc(r.catatan)}"></label>
          <label>Alasan perubahan (wajib) <input id="e-alasan"></label>
          <div class="galat" id="e-galat"></div>
          <div class="aksi"><button id="e-batal">Batal</button><button id="e-ok">Simpan</button></div>`);
        const num = id => { const t = _r(id).value.trim(); return t === "" ? null : parseInt(t, 10); };
        _r("e-batal").onclick = () => Dash.tutupModal();
        _r("e-ok").onclick = async () => {
            try {
                await API.UbahTransaksi({
                    id: r.id, status: _r("e-status").value, jumlah_ambil: num("e-ja"), jumlah_kembali: num("e-jk"),
                    waktu_ambil: _r("e-wa").value.trim(), waktu_kembali: _r("e-wk").value.trim(),
                    catatan: _r("e-cat").value, alasan: _r("e-alasan").value,
                });
                Dash.tutupModal(); toast("Perubahan tersimpan", "ok"); this.muatTrx();
            } catch (e) { _r("e-galat").textContent = String(e.message || e); bunyi("err"); }
        };
    },

    dialogHapus(r) {
        Dash.bukaModal(`
          <h2>Hapus transaksi?</h2>
          <p>${esc(r.nama_kelas)} - ${esc(r.tanggal)} (${esc(r.status)}). Tindakan ini tercatat di audit log.</p>
          <label>Alasan penghapusan (wajib) <input id="h-alasan"></label>
          <div class="galat" id="h-galat"></div>
          <div class="aksi"><button id="h-batal">Batal</button><button id="h-ok">Hapus</button></div>`);
        _r("h-batal").onclick = () => Dash.tutupModal();
        _r("h-ok").onclick = async () => {
            try {
                await API.HapusTransaksi(r.id, _r("h-alasan").value);
                Dash.tutupModal(); toast("Transaksi dihapus", "ok"); this.muatTrx();
            } catch (e) { _r("h-galat").textContent = String(e.message || e); bunyi("err"); }
        };
    },
};

document.addEventListener("DOMContentLoaded", () => {
    _r("r-btn-trx").onclick = () => Riwayat.pilih("trx");
    _r("r-btn-audit").onclick = () => Riwayat.pilih("audit");
    _r("rf-cari").onclick = () => Riwayat.muatTrx();
    _r("ra-cari").onclick = () => Riwayat.muatAudit();

    _r("btn-tutup").onclick = () => {
        Dash.bukaModal(`
          <h2>Tutup hari dengan catatan</h2>
          <p>Kelas yang sudah mengambil tapi belum mengembalikan menjadi <b>Belum kembali</b>. Kelas yang tidak mengambil dicatat <b>Tidak mengambil</b> (tidak masuk laporan).</p>
          <label>Catatan <input id="t-cat" placeholder="mis. kelas VII C tidak masuk"></label>
          <div class="galat" id="t-galat"></div>
          <div class="aksi"><button id="t-batal">Batal</button><button id="t-ok">Tutup hari</button></div>`);
        _r("t-batal").onclick = () => Dash.tutupModal();
        _r("t-ok").onclick = async () => {
            try { await API.TutupHari(_r("t-cat").value); Dash.tutupModal(); toast("Hari ditutup", "ok"); Dash.muat(); }
            catch (e) { _r("t-galat").textContent = String(e.message || e); bunyi("err"); }
        };
    };

    _r("btn-buka").onclick = () => {
        Dash.bukaModal(`
          <h2>Buka kembali hari ini?</h2>
          <p>Status otomatis dikembalikan: Belum kembali menjadi Sudah ambil, dan baris Tidak mengambil dihapus.</p>
          <label>Alasan (wajib) <input id="b-alasan"></label>
          <div class="galat" id="b-galat"></div>
          <div class="aksi"><button id="b-batal">Batal</button><button id="b-ok">Buka</button></div>`);
        _r("b-batal").onclick = () => Dash.tutupModal();
        _r("b-ok").onclick = async () => {
            try {
                const st = await API.StatusHariIni();
                await API.BukaHari(st.tanggal, _r("b-alasan").value);
                Dash.tutupModal(); toast("Hari dibuka kembali", "ok"); Dash.muat();
            } catch (e) { _r("b-galat").textContent = String(e.message || e); bunyi("err"); }
        };
    };
});
zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ for f in internal/service/service_test.go internal/service/tahap5_test.go; do echo "=== $f ==="; cat "$f"; done
wc -l internal/service/*.go internal/export/*.go frontend/js/*.js
=== internal/service/service_test.go ===
package service

import (
	"path/filepath"
	"testing"
	"time"

	"mbg/internal/db"
	"mbg/internal/timeutil"
)

func TestAlurTap(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	now := time.Date(2026, 1, 12, 11, 0, 0, 0, timeutil.WIB)
	s := NewWithClock(conn, t.TempDir(), func() time.Time { return now })

	id, err := s.TambahKelas("VII A", "Budi", 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.DaftarkanKartu("12345678", id, "", ""); err != nil {
		t.Fatal(err)
	}

	if r := s.ProsesTap("0012345678"); r.Aksi != "AMBIL" {
		t.Fatalf("tap 1: %+v", r)
	}
	// popup terbuka -> tap lain ditolak
	if r := s.ProsesTap("0099999999"); r.Aksi != "PESAN" {
		t.Fatalf("penguncian popup: %+v", r)
	}
	if err := s.SubmitAmbil(id, 0); err == nil {
		t.Fatal("jumlah 0 harus ditolak")
	}
	if err := s.SubmitAmbil(id, 31); err != nil {
		t.Fatal(err)
	}

	now = now.Add(5 * time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "KONFIRMASI_KEMBALI" {
		t.Fatalf("jeda minimum: %+v", r)
	}
	s.BatalInput()

	now = now.Add(15 * time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "KEMBALI" || r.JumlahAmbil != 31 {
		t.Fatalf("kembali: %+v", r)
	}
	if err := s.SubmitKembali(id, 32); err == nil {
		t.Fatal("kembali > ambil harus ditolak")
	}
	if err := s.SubmitKembali(id, 30); err != nil {
		t.Fatal(err)
	}

	now = now.Add(time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "PESAN" {
		t.Fatalf("sudah selesai: %+v", r)
	}

	st, _ := s.StatusHariIni()
	if !st.SemuaSelesai || st.Baris[0].Keterangan != "Kurang 1 Ompreng" || !st.Baris[0].Merah {
		t.Fatalf("status: %+v", st)
	}
}
=== internal/service/tahap5_test.go ===
package service

import (
	"path/filepath"
	"testing"
	"time"

	"mbg/internal/db"
	"mbg/internal/model"
	"mbg/internal/timeutil"
)

const tglUji = "2026-01-12"

func setupT5(t *testing.T) *Service {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	fixed := time.Date(2026, 1, 12, 10, 0, 0, 0, timeutil.WIB)
	return NewWithClock(conn, t.TempDir(), func() time.Time { return fixed })
}

func harusOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func statusPerKelas(t *testing.T, s *Service) (map[int64]string, model.StatusHari) {
	t.Helper()
	st, err := s.StatusHariIni()
	harusOK(t, err)
	m := map[int64]string{}
	for _, b := range st.Baris {
		m[b.KelasID] = b.Status
	}
	return m, st
}

func TestTutupDanBukaHari(t *testing.T) {
	s := setupT5(t)
	a, _ := s.TambahKelas("VII A", "Ani", 1)
	b, _ := s.TambahKelas("VII B", "Budi", 2)
	c, _ := s.TambahKelas("VII C", "Cici", 3)
	harusOK(t, s.DaftarkanKartu("1111111111", a, "", ""))
	harusOK(t, s.SubmitAmbil(a, 10))
	harusOK(t, s.SubmitAmbil(b, 8))
	harusOK(t, s.SubmitKembali(b, 8))

	if ok, _ := s.ApakahBolehCetak(tglUji); ok {
		t.Fatal("belum boleh cetak sebelum semua selesai/ditutup")
	}
	harusOK(t, s.TutupHari("uji"))
	m, st := statusPerKelas(t, s)
	if m[a] != "BELUM_KEMBALI" || m[b] != "SELESAI" || m[c] != "TIDAK_AMBIL" {
		t.Fatalf("status setelah tutup salah: %v", m)
	}
	if !st.Ditutup || !st.BolehCetak {
		t.Fatalf("Ditutup=%v BolehCetak=%v", st.Ditutup, st.BolehCetak)
	}
	if err := s.TutupHari("lagi"); err == nil {
		t.Fatal("tutup dua kali seharusnya ditolak")
	}
	if r := s.ProsesTap("1111111111"); r.Pesan != "Hari ini sudah ditutup" {
		t.Fatalf("tap saat ditutup: %+v", r)
	}

	if err := s.BukaHari(tglUji, ""); err == nil {
		t.Fatal("buka tanpa alasan seharusnya ditolak")
	}
	if err := s.BukaHari("2026-01-11", "x"); err == nil {
		t.Fatal("buka hari lampau seharusnya ditolak")
	}
	harusOK(t, s.BukaHari(tglUji, "salah tutup"))
	m, st = statusPerKelas(t, s)
	if m[a] != "AMBIL" || m[b] != "SELESAI" || m[c] != "BELUM_AMBIL" || st.Ditutup {
		t.Fatalf("status setelah buka salah: %v ditutup=%v", m, st.Ditutup)
	}
}

func TestUbahDanHapusTransaksi(t *testing.T) {
	s := setupT5(t)
	a, _ := s.TambahKelas("VII A", "Ani", 1)
	harusOK(t, s.SubmitAmbil(a, 10))
	harusOK(t, s.SubmitKembali(a, 10))
	list, err := s.ListTransaksi(model.FilterTransaksi{})
	harusOK(t, err)
	if len(list) != 1 {
		t.Fatalf("mau 1 transaksi, dapat %d", len(list))
	}
	id := list[0].ID
	ja, jk := 10, 12
	e := model.EditTransaksi{ID: id, JumlahAmbil: &ja, JumlahKembali: &jk, WaktuAmbil: "09:00", WaktuKembali: "10:00", Status: "SELESAI", Alasan: "salah ketik"}
	if err := s.UbahTransaksi(e); err == nil {
		t.Fatal("kembali > ambil seharusnya ditolak saat edit")
	}
	jk = 7
	e.Alasan = ""
	if err := s.UbahTransaksi(e); err == nil {
		t.Fatal("tanpa alasan seharusnya ditolak")
	}
	e.Alasan = "koreksi"
	harusOK(t, s.UbahTransaksi(e))
	list, _ = s.ListTransaksi(model.FilterTransaksi{})
	if list[0].Keterangan != "Kurang 3 Ompreng" || !list[0].Merah {
		t.Fatalf("keterangan setelah edit: %q merah=%v", list[0].Keterangan, list[0].Merah)
	}

	if err := s.HapusTransaksi(id, " "); err == nil {
		t.Fatal("hapus tanpa alasan seharusnya ditolak")
	}
	harusOK(t, s.HapusTransaksi(id, "dobel"))
	list, _ = s.ListTransaksi(model.FilterTransaksi{})
	if len(list) != 0 {
		t.Fatal("transaksi seharusnya sudah terhapus")
	}
	au, err := s.ListAudit(model.FilterAudit{Entitas: "transaksi"})
	harusOK(t, err)
	if len(au) < 3 || au[0].Aksi != "HAPUS" || au[0].Alasan != "dobel" {
		t.Fatalf("audit tidak sesuai: %+v", au)
	}
}
  252 internal/service/ekspor.go
   35 internal/service/ekspor_test.go
  120 internal/service/hari.go
  218 internal/service/master.go
   25 internal/service/reader.go
  224 internal/service/riwayat.go
   64 internal/service/service.go
   70 internal/service/service_test.go
  126 internal/service/tahap5_test.go
  270 internal/service/transaksi.go
  174 internal/export/excel.go
   30 internal/export/export.go
  152 internal/export/pdf.go
   28 frontend/js/api.js
   79 frontend/js/dashboard.js
   62 frontend/js/ekspor.js
   65 frontend/js/main.js
  105 frontend/js/master.js
  153 frontend/js/riwayat.js
   30 frontend/js/tap.js
 2282 total
zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ 

Terima kasih, kode sudah cukup untuk menyusun rencana Fitur 3. Sesuai SPEC, belum ada kode yang saya tulis; berikut rencananya untuk Anda setujui.

## 1. Keputusan desain: status baru `KURANG`

Saya pilih status baru `KURANG`, bukan `SELESAI` + flag, karena logika lama sudah memperlakukan `SELESAI` sebagai "final dan dihitung selesai". Dengan flag, saya harus mengubah semua tempat itu:

| Titik | Dengan status `KURANG` | Dengan `SELESAI` + flag |
|---|---|---|
| `SemuaSelesai` / `ApakahBolehCetak` | tidak berubah, `KURANG` otomatis tidak dihitung selesai | harus dipotong manual |
| `TutupHari` | cukup `status IN ('AMBIL','KURANG')` | harus mencari flag |
| `ProsesTap` | cabang baru yang jelas | cabang `SELESAI` menjadi bercabang lagi |
| `Keterangan()` | satu `case` baru | cabang flag di dalam `case SELESAI` |

Data lama dan edit manual tidak diubah. `SELESAI` dengan `kembali < ambil` tetap sah sebagai final, dan keterangannya tetap "Kurang N Ompreng", jadi `TestUbahDanHapusTransaksi` tetap lulus. `KURANG` hanya dihasilkan oleh alur tap baru.

## 2. Skema (migrasi idempoten)

Di `schema` pada `db.go`:

```sql
CREATE TABLE IF NOT EXISTS pengembalian (
  id INTEGER PRIMARY KEY,
  transaksi_id INTEGER NOT NULL REFERENCES transaksi(id) ON DELETE CASCADE,
  waktu TEXT NOT NULL, jumlah INTEGER NOT NULL, nisn TEXT, nama_siswa TEXT
);
CREATE INDEX IF NOT EXISTS idx_pengembalian_trx ON pengembalian(transaksi_id);
```

Di `Migrate()`, backfill satu baris per transaksi `SELESAI` lama yang belum punya rincian:

```sql
INSERT INTO pengembalian (transaksi_id,waktu,jumlah,nisn,nama_siswa)
SELECT id, waktu_kembali, jumlah_kembali, nisn_kembali, nama_kembali FROM transaksi t
WHERE status='SELECT' ... AND jumlah_kembali IS NOT NULL AND waktu_kembali IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM pengembalian p WHERE p.transaksi_id=t.id)
```

Pada versi final, `status='SELESAI'`. Karena memakai `NOT EXISTS`, aman dijalankan setiap aplikasi dibuka.

Kolom di `transaksi` dipertahankan sebagai ringkasan, jadi query riwayat, laporan, dan `kolomTrx` tidak perlu join:
- `jumlah_kembali` berisi total akumulasi.
- `waktu_kembali` berisi waktu pengembalian terakhir.
- `nama_kembali` berisi nama pengembali unik yang dipisah koma, diperbarui tiap submit.

## 3. Perubahan per berkas

**`model.go`**
- `TapResult` dapat field `Sisa`, dan `Aksi` dapat nilai baru `KEMBALI_SUSULAN`.
- Struct baru `Pengembalian{Waktu, Jumlah, NISN, NamaSiswa}`.
- `Transaksi` dapat `Pengembalian []Pengembalian` dan `Susulan bool`.

**`transaksi.go`**
- `ProsesTap` membaca `jumlah_kembali` juga, lalu bercabang menurut status:
  - `AMBIL`: seperti sekarang.
  - `KURANG`: aksi `KEMBALI_SUSULAN` dengan `Sisa`, tanpa pengecekan jeda minimum.
  - `SELESAI` dengan `kembali >= ambil`: tolak dengan pesan "Kelas X sudah mengembalikan ompreng dengan lengkap". Ini berlaku untuk kartu siapa pun di kelas itu, dan popup tidak dibuka.
  - `SELESAI` dengan `kembali < ambil` (data lama): pesan lama "sudah selesai hari ini".
- `SubmitKembali`:
  - Status `AMBIL`: jumlah boleh 0..ambil.
  - Status `KURANG`: jumlah harus 1..sisa.
  - Setiap submit menambah baris `pengembalian`, memperbarui ringkasan, dan menetapkan `SELESAI` bila total sama dengan ambil, selain itu `KURANG`.
  - Audit memakai `aksi=UBAH` dengan alasan "pengembalian susulan", nilai lama berupa status dan total lama, nilai baru berupa status, total, jumlah susulan, dan pengembali.
- `Keterangan()` menambah parameter `susulan bool`, dengan tiga pemanggil: `StatusHariIni`, `scanTrx`, `susunLaporan`:
  - `KURANG` menampilkan "Kurang N Ompreng" (merah).
  - `SELESAI` dari susulan menampilkan "Lengkap (susulan)".
  - `BELUM_KEMBALI` yang sudah punya sebagian pengembalian menampilkan "Belum lengkap, kurang N Ompreng" (merah).
- `susulan` dihitung lewat subquery `COUNT(*)` pada `pengembalian > 1`.

**`hari.go`**
- `TutupHari`: `status IN ('AMBIL','KURANG')` menjadi `BELUM_KEMBALI`.
- `BukaHari`: `BELUM_KEMBALI` dengan `jumlah_kembali IS NOT NULL` kembali ke `KURANG`, sedangkan yang `NULL` kembali ke `AMBIL` seperti sekarang.

**`riwayat.go`**
- `validasiEdit` menerima `KURANG`: jumlah ambil dan waktu ambil wajib, `jumlah_kembali` 0..ambil-1, waktu kembali wajib.
- `UbahTransaksi` menyinkronkan `pengembalian` agar total tetap konsisten:
  - Bila total baru berbeda dari jumlah baris rincian, rincian diganti satu baris "koreksi".
  - Bila status menjadi `AMBIL`, `BELUM_KEMBALI`, atau `TIDAK_AMBIL`, rincian dihapus.
  - Rincian lama ikut tercatat di audit.
- `ListTransaksi` mengisi `Pengembalian` dengan satu query yang dikelompokkan.
- `HapusTransaksi` otomatis menghapus rincian lewat `ON DELETE CASCADE`.

**`ekspor.go`**: tambah subquery `susulan`, panggil `Keterangan` yang baru. Kolom Pengembali sudah berasal dari `nama_kembali`, jadi nama ganda otomatis tampil dipisah koma. `export/*` tidak berubah.

**Frontend**
- `dashboard.js`: popup mode `SUSULAN` ("Pengembalian susulan", sisa, input 1..sisa) memakai `SubmitKembali` yang sama. Label status "Kurang".
- `riwayat.js`: opsi `KURANG` di dialog Ubah, dan rincian parsial (jam, jumlah, nama) di bawah keterangan.
- `index.html`: tambah opsi `KURANG` di filter status (perlu saya lihat dulu).

Tidak ada perubahan signature binding Wails. `models.ts` hanya berubah karena field baru, dan akan diregenerasi oleh `wails build`.

## 4. Cooldown dan debounce (keputusan terdokumentasi)

- Debounce 2 detik per UID tetap.
- Cooldown kelas 30 detik tetap setelah pengambilan dan pengembalian lengkap.
- Setelah pengembalian yang hasilnya `KURANG`, cooldown dipendekkan menjadi 5 detik. Alasannya, dengan 30 detik teman sekelas yang langsung menyusul membawa sisa ompreng akan ditolak "Baru saja tercatat", padahal itu alur yang sah. Lima detik tetap menahan tap ganda dari kartu yang masih menempel.
- Penolakan "sudah lengkap" dipindah sebelum cek cooldown agar pesannya selalu jelas.

## 5. Tes

- `susulan_test.go` baru, mencakup skenario SPEC: ambil 10, kembali 7, tap lagi (susulan, sisa 3), kembali 3 menjadi `SELESAI`, lalu tap ditolak.
- Susulan melebihi sisa ditolak, susulan 0 ditolak, dan kembali lengkap langsung lalu tap ditolak (juga dengan kartu siswa lain).
- Tutup hari saat `KURANG` menjadi `BELUM_KEMBALI` dengan keterangan sisa, buka hari kembali ke `KURANG`, dan tap saat ditutup tetap ditolak.
- Ekspor: keterangan "Kurang N" / "Lengkap (susulan)" dan pengembali ganda dipisah koma.
- Migrasi: DB lama tanpa tabel `pengembalian` ter-backfill, dan menjalankan migrasi dua kali tidak menggandakan baris.
- Edit transaksi menyinkronkan rincian, dan hapus transaksi ikut menghapus rincian.
- Cooldown 5 detik: tap susulan setelah 6 detik diterima.
- **`TestAlurTap` harus diubah.** Di sana `SubmitKembali(30)` dari ambil 31 dulu dianggap final, dan sekarang menjadi `KURANG` sehingga tap berikutnya membuka popup susulan.

## 6. Tiga hal yang perlu persetujuan Anda

1. **Tombol Cetak PDF.** `KURANG` tidak dihitung "selesai", jadi kelas yang kurang membuat Cetak PDF baru aktif setelah susulan lengkap atau hari ditutup. Ini konsisten dengan SPEC butir 8, tetapi mengubah README langkah 5. Setuju? Alternatifnya `KURANG` dihitung boleh cetak, tetapi itu mengubah `SemuaSelesai` dan `ApakahBolehCetak`.
2. **Pengembalian pertama 0 ompreng.** Saat ini diperbolehkan. Saya pertahankan: hasilnya `KURANG` dengan sisa penuh, dan susulan baru wajib 1..sisa. Setuju?
3. **Cooldown 5 detik** setelah hasil `KURANG`, seperti di bagian 4. Setuju?

Setelah Anda setuju, saya perlu melihat tiga berkas ini untuk menyesuaikan frontend dan tes yang sudah ada:

```bash
for f in frontend/index.html frontend/css/style.css internal/service/ekspor_test.go frontend/js/api.js; do echo "=== $f ==="; cat "$f"; done
```

Berkas kode akan saya berikan utuh per berkas, siap ditimpa. Saran versi untuk Fitur 3 adalah **0.2.0**, karena aturan bisnis dan skema berubah, dan akan saya konfirmasi lagi di ringkasan akhir.

zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ for f in frontend/index.html frontend/css/style.css internal/service/ekspor_test.go frontend/js/api.js; do echo "=== $f ==="; cat "$f"; done
=== frontend/index.html ===
<!DOCTYPE html>
<html lang="id">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>MBG - Pencatatan Ompreng</title>
    <link rel="stylesheet" href="css/style.css">
</head>

<body>
    <header>
        <h1>MBG <span>Pencatatan Ompreng</span></h1>
        <nav>
            <button data-tab="beranda" class="aktif">Beranda</button>
            <button data-tab="master">Master Data</button>
            <button data-tab="riwayat">Riwayat</button>
            <button data-tab="ekspor">Ekspor</button>
        </nav>
        <span id="status-reader" class="kecil">Reader serial: mencari...</span>
        <span id="info-tanggal"></span>
    </header>

    <main>
        <!-- BERANDA -->
        <section id="tab-beranda" class="tab aktif">
            <table id="tabel-hari">
                <thead>
                    <tr>
                        <th>No</th>
                        <th>Kelas</th>
                        <th>Ketua</th>
                        <th>Status</th>
                        <th>Ambil</th>
                        <th>Kembali</th>
                        <th>Waktu Ambil</th>
                        <th>Waktu Kembali</th>
                        <th>Keterangan</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
            <p id="ringkasan"></p>
            <div class="bar bar-beranda">
                <div class="grup">
                    <button id="btn-uid-manual">Input UID manual</button>
                </div>
                <div class="grup">
                    <button id="btn-tutup">Tutup hari</button>
                    <button id="btn-buka" class="hidden">Buka kembali hari ini</button>
                    <button id="btn-cetak" class="utama" disabled
                        title="Aktif jika semua kelas selesai atau hari ditutup">Cetak PDF</button>
                </div>
            </div>
            <details class="sim">
                <summary>Simulator tap (pengembangan)</summary>
                <input id="sim-uid" placeholder="UID, mis. 0012345678" maxlength="12">
                <button id="sim-kirim">Kirim tap</button>
                <button id="sim-hid">Simulasikan HID (ketik cepat)</button>
            </details>
        </section>

        <!-- MASTER -->
        <section id="tab-master" class="tab">
            <div class="subnav">
                <button data-sub="instansi" class="aktif">Instansi</button>
                <button data-sub="kelas">Kelas &amp; Ketua</button>
                <button data-sub="kartu">Kartu RFID</button>
            </div>

            <div id="sub-instansi" class="sub aktif">
                <label>Nama instansi <input id="p-nama"></label>
                <label>Alamat <input id="p-alamat"></label>
                <label>Penanggung jawab <input id="p-pj"></label>
                <label>NIP <input id="p-nip"></label>
                <label>Jabatan <input id="p-jabatan"></label>
                <label>Jeda minimum ambil &rarr; kembali (menit) <input id="p-jeda" type="number" min="0"></label>
                <label>Ambang kecepatan HID (ms) <input id="p-hid" type="number" min="10"></label>
                <div class="form-baris">
                    <span class="kecil" id="logo-status"></span>
                    <button id="logo-pilih">Pilih logo (PNG/JPEG)</button>
                    <button id="logo-hapus">Hapus logo</button>
                </div>
                <div class="form-baris"><button id="p-simpan" class="utama">Simpan</button></div>
                <p class="kecil" id="info-sistem"></p>
            </div>

            <div id="sub-kelas" class="sub">
                <div class="form-baris">
                    <input id="k-nama" placeholder="Nama kelas (VII A)">
                    <input id="k-ketua" placeholder="Nama ketua murid">
                    <input id="k-urutan" type="number" placeholder="Urutan" value="0" style="width:90px">
                    <button id="k-simpan" class="utama">Tambah</button>
                    <button id="k-batal" class="hidden">Batal ubah</button>
                </div>
                <label class="cek"><input type="checkbox" id="k-semua"> Tampilkan kelas nonaktif</label>
                <table id="tabel-kelas">
                    <thead>
                        <tr>
                            <th>Urutan</th>
                            <th>Kelas</th>
                            <th>Ketua</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody></tbody>
                </table>
            </div>

            <div id="sub-kartu" class="sub">
                <div class="form-baris">
                    <select id="c-kelas"></select>
                    <input id="c-uid" placeholder="Tap kartu atau ketik UID" maxlength="12">
                    <input id="c-nisn" placeholder="NISN (opsional)" maxlength="10">
                    <input id="c-nama" placeholder="Nama siswa (opsional)">
                    <button id="c-daftar" class="utama">Daftarkan</button>
                </div>
                <p class="kecil">Di tab ini, tap kartu akan mengisi kolom UID (tidak memproses ambil/kembali).</p>
                <table id="tabel-kartu">
                    <thead>
                        <tr>
                            <th>UID</th>
                            <th>NISN</th>
                            <th>Nama Siswa</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody></tbody>
                </table>
            </div>
        </section>

        <!-- RIWAYAT -->
        <section id="tab-riwayat" class="tab">
            <div class="bar">
                <button id="r-btn-trx" class="aktif">Transaksi</button>
                <button id="r-btn-audit">Audit Log</button>
            </div>
            <div id="r-trx">
                <div class="form-baris">
                    <label>Dari <input id="rf-dari" type="date"></label>
                    <label>Sampai <input id="rf-sampai" type="date"></label>
                    <select id="rf-kelas"></select>
                    <select id="rf-status">
                        <option value="">Semua status</option>
                        <option value="AMBIL">AMBIL</option>
                        <option value="SELESAI">SELESAI</option>
                        <option value="BELUM_KEMBALI">BELUM_KEMBALI</option>
                        <option value="TIDAK_AMBIL">TIDAK_AMBIL</option>
                    </select>
                    <button id="rf-cari" class="utama">Tampilkan</button>
                </div>
                <table id="tabel-riwayat">
                    <thead>
                        <tr>
                            <th>Tanggal</th>
                            <th>Kelas</th>
                            <th>Ketua</th>
                            <th>Pengambil</th>
                            <th>Pengembali</th>
                            <th>Status</th>
                            <th>Ambil</th>
                            <th>Kembali</th>
                            <th>Waktu</th>
                            <th>Keterangan</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody></tbody>
                </table>
            </div>
            <div id="r-audit" class="hidden">
                <div class="form-baris">
                    <label>Dari <input id="ra-dari" type="date"></label>
                    <label>Sampai <input id="ra-sampai" type="date"></label>
                    <select id="ra-entitas">
                        <option value="">Semua entitas</option>
                        <option>transaksi</option>
                        <option>kelas</option>
                        <option>kartu</option>
                        <option>pengaturan</option>
                        <option>hari</option>
                    </select>
                    <button id="ra-cari" class="utama">Tampilkan</button>
                </div>
                <table id="tabel-audit">
                    <thead>
                        <tr>
                            <th>Waktu</th>
                            <th>Aksi</th>
                            <th>Entitas</th>
                            <th>ID</th>
                            <th>Nilai lama</th>
                            <th>Nilai baru</th>
                            <th>Alasan</th>
                        </tr>
                    </thead>
                    <tbody></tbody>
                </table>
            </div>
        </section>

        <!-- EKSPOR -->
        <section id="tab-ekspor" class="tab">
            <div class="form-baris">
                <label>Dari <input id="ek-dari" type="date"></label>
                <label>Sampai <input id="ek-sampai" type="date"></label>
                <button id="ek-lihat">Pratinjau</button>
                <button id="ek-xlsx">Ekspor Excel</button>
                <button id="ek-pdf" class="utama">Ekspor PDF</button>
            </div>
            <p class="kecil" id="ek-info"></p>
            <table id="tabel-ekspor">
                <thead>
                    <tr>
                        <th>No</th>
                        <th>Kelas</th>
                        <th>Ketua</th>
                        <th>Pengambil</th>
                        <th>Pengembali</th>
                        <th>Tanggal</th>
                        <th>Ambil (waktu)</th>
                        <th>Kembali (waktu)</th>
                        <th>Jml Ambil</th>
                        <th>Jml Kembali</th>
                        <th>Keterangan</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
        </section>
    </main>

    <footer>
        built with <span class="hati">&#9829;</span> by
        <a href="https://zulfikriyahya.my.id" id="link-footer">Yahya Zulfikri</a>
    </footer>

    <div id="toast" class="hidden"></div>
    <div id="modal" class="hidden">
        <div class="kotak" id="modal-isi"></div>
    </div>

    <script src="js/api.js"></script>
    <script src="js/tap.js"></script>
    <script src="js/dashboard.js"></script>
    <script src="js/master.js"></script>
    <script src="js/riwayat.js"></script>
    <script src="js/ekspor.js"></script>
    <script src="js/main.js"></script>
    <script>
        document.getElementById("link-footer").addEventListener("click", function (e) {
            e.preventDefault();
            if (window.runtime && window.runtime.BrowserOpenURL) window.runtime.BrowserOpenURL(this.href);
            else window.open(this.href, "_blank");
        });
    </script>
</body>

</html>=== frontend/css/style.css ===
:root {
    --hijau: #15803d;
    --hijau-tua: #14532d;
    --hijau-muda: #dcfce7;
    --merah: #b91c1c;
    --merah-muda: #ffe4e6;
    --merah-teks: #9f1239;
    --teks: #0f172a;
    --redup: #64748b;
    --garis: #e2e8f0;
    --latar: #f1f5f9;
    --putih: #fff;
    --radius: 10px;
}

* {
    box-sizing: border-box
}

html,
body {
    height: 100%
}

body {
    margin: 0;
    display: flex;
    flex-direction: column;
    min-height: 100vh;
    font-family: "Inter", "Segoe UI", system-ui, -apple-system, "Noto Sans", "Helvetica Neue", Arial, sans-serif;
    font-size: 14px;
    line-height: 1.5;
    color: var(--teks);
    background: var(--latar);
    -webkit-font-smoothing: antialiased;
}

/* ---------- Header ---------- */
header {
    display: flex;
    align-items: center;
    gap: 20px;
    padding: 0 24px;
    min-height: 58px;
    background: var(--putih);
    border-bottom: 1px solid var(--garis);
    border-top: 4px solid var(--hijau);
}

header h1 {
    margin: 0;
    font-size: 17px;
    font-weight: 700;
    letter-spacing: -.01em;
    color: var(--hijau-tua);
}

header h1 span {
    font-weight: 400;
    color: var(--redup);
    margin-left: 4px
}

header nav {
    display: flex;
    gap: 2px;
    align-self: stretch
}

header nav button {
    background: transparent;
    color: var(--redup);
    border: 0;
    border-bottom: 3px solid transparent;
    border-radius: 0;
    padding: 0 14px;
    font-size: 14px;
    font-weight: 500;
}

header nav button:hover {
    color: var(--teks);
    background: transparent
}

header nav button.aktif {
    color: var(--hijau-tua);
    border-bottom-color: var(--hijau);
    font-weight: 600
}

#status-reader {
    font-weight: 500
}

#info-tanggal {
    margin-left: auto;
    font-weight: 600;
    color: var(--teks)
}

/* ---------- Layout ---------- */
main {
    flex: 1;
    width: 100%;
    max-width: 1280px;
    margin: 0 auto;
    padding: 24px;
}

.tab,
.sub {
    display: none
}

.tab.aktif,
.sub.aktif {
    display: block
}

/* ---------- Tabel ---------- */
table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
    background: var(--putih);
    border: 1px solid var(--garis);
    border-radius: var(--radius);
    overflow: hidden;
    box-shadow: 0 1px 2px rgba(15, 23, 42, .04);
    font-variant-numeric: tabular-nums;
}

th,
td {
    padding: 10px 12px;
    text-align: left;
    border-bottom: 1px solid var(--garis);
}

th {
    background: #f8fafc;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: .04em;
    color: var(--redup);
    white-space: nowrap;
}

tbody tr:last-child td {
    border-bottom: 0
}

tbody tr:hover td {
    background: #f8fafc
}

tr.selesai td {
    background: var(--hijau-muda)
}

tr.merah td {
    background: var(--merah-muda);
    color: var(--merah-teks)
}

tr.selesai:hover td {
    background: #cdf7da
}

tr.merah:hover td {
    background: #ffd5da
}

#ringkasan {
    margin: 12px 2px;
    color: var(--redup);
    font-weight: 500
}

/* ---------- Form ---------- */
.bar,
.form-baris {
    display: flex;
    align-items: flex-end;
    gap: 10px;
    margin: 16px 0;
    flex-wrap: wrap;
}

label {
    display: block;
    margin: 12px 0;
    font-weight: 500;
    color: var(--teks)
}

.form-baris label {
    margin: 0
}

label input {
    display: block;
    width: 440px;
    max-width: 100%;
    margin-top: 5px;
}

.form-baris label input {
    width: auto
}

label.cek {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 400;
    margin: 6px 0 14px
}

label.cek input {
    display: inline-block;
    width: auto;
    margin: 0
}

input,
select {
    padding: 8px 10px;
    font: inherit;
    color: var(--teks);
    background: var(--putih);
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    outline: none;
    transition: border-color .15s, box-shadow .15s;
}

input:focus,
select:focus {
    border-color: var(--hijau);
    box-shadow: 0 0 0 3px rgba(21, 128, 61, .18);
}

button {
    padding: 8px 16px;
    font: inherit;
    font-weight: 500;
    color: var(--teks);
    background: var(--putih);
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    cursor: pointer;
    transition: background .15s, border-color .15s;
}

button:hover:not(:disabled) {
    background: #f1f5f9;
    border-color: #94a3b8
}

button.utama,
#btn-cetak,
#in-ok,
#um-ok,
#kf-ya {
    color: #fff;
    background: var(--hijau);
    border-color: var(--hijau);
}

button.utama:hover:not(:disabled),
#btn-cetak:hover:not(:disabled),
#in-ok:hover:not(:disabled),
#um-ok:hover:not(:disabled),
#kf-ya:hover:not(:disabled) {
    background: var(--hijau-tua);
    border-color: var(--hijau-tua);
}

button:disabled {
    opacity: .45;
    cursor: not-allowed
}

.subnav {
    display: inline-flex;
    gap: 4px;
    padding: 4px;
    margin-bottom: 14px;
    background: #e2e8f0;
    border-radius: 10px
}

.subnav button {
    border: 0;
    background: transparent;
    padding: 6px 14px
}

.subnav button.aktif {
    background: var(--putih);
    font-weight: 600;
    box-shadow: 0 1px 2px rgba(15, 23, 42, .1)
}

#r-btn-trx.aktif,
#r-btn-audit.aktif {
    font-weight: 600;
    background: var(--hijau-muda);
    border-color: var(--hijau);
    color: var(--hijau-tua);
}

.kecil {
    font-size: 12px;
    color: var(--redup)
}

.hidden {
    display: none !important
}

.sim {
    margin-top: 24px;
    padding: 10px 14px;
    background: var(--putih);
    border: 1px dashed #94a3b8;
    border-radius: var(--radius);
}

.sim summary {
    cursor: pointer;
    color: var(--redup);
    font-weight: 500
}

.sim input,
.sim button {
    margin: 10px 6px 0 0
}

#tabel-audit td {
    max-width: 260px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap
}

/* ---------- Toast ---------- */
#toast {
    position: fixed;
    top: 76px;
    left: 50%;
    transform: translateX(-50%);
    padding: 12px 26px;
    font-size: 16px;
    font-weight: 600;
    color: #fff;
    border-radius: 999px;
    box-shadow: 0 8px 24px rgba(15, 23, 42, .25);
    z-index: 20;
}

#toast.ok {
    background: #15803d
}

#toast.err {
    background: #b91c1c
}

#toast.info {
    background: #1d4ed8
}

/* ---------- Modal ---------- */
#modal {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(15, 23, 42, .55);
    backdrop-filter: blur(2px);
    z-index: 10;
}

.kotak {
    min-width: 400px;
    max-width: 92vw;
    padding: 28px;
    background: var(--putih);
    border-radius: 14px;
    box-shadow: 0 20px 50px rgba(15, 23, 42, .35);
}

.kotak h2 {
    margin: 0 0 8px;
    font-size: 20px;
    letter-spacing: -.01em
}

.kotak p {
    margin: 6px 0 14px;
    color: var(--redup)
}

.kotak input#in-jumlah,
.kotak input#um {
    width: 100%;
    padding: 12px;
    font-size: 30px;
    font-weight: 600;
    text-align: center;
}

.kotak .galat {
    min-height: 22px;
    margin: 8px 0;
    color: var(--merah);
    font-weight: 500
}

.kotak .aksi {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
    margin-top: 8px
}

#modal label {
    display: block;
    margin: .5rem 0
}

#modal label input,
#modal label select {
    display: block;
    width: 100%;
    margin-top: 4px
}

/* ---------- Footer ---------- */
footer {
    padding: 14px 24px;
    text-align: center;
    font-size: 12px;
    color: var(--redup);
    border-top: 1px solid var(--garis);
    background: var(--putih);
}

footer a {
    color: var(--hijau-tua);
    font-weight: 600;
    text-decoration: none
}

footer a:hover {
    text-decoration: underline
}

footer .hati {
    color: #e11d48
}

/* ---------- Tombol aksi ---------- */
.bar-beranda {
    justify-content: space-between;
    align-items: center
}

.bar-beranda .grup {
    display: flex;
    gap: 10px;
    flex-wrap: wrap
}

td.aksi-sel {
    width: 1%;
    white-space: nowrap
}

.aksi-baris {
    display: flex;
    justify-content: flex-end;
    gap: 6px
}

.aksi-baris button {
    padding: 4px 12px;
    font-size: 13px
}

button.bahaya {
    color: var(--merah);
    border-color: #fca5a5
}

button.bahaya:hover:not(:disabled) {
    background: #fee2e2;
    border-color: var(--merah)
}

#h-ok,
#t-ok {
    color: #fff;
    background: var(--merah);
    border-color: var(--merah)
}

#h-ok:hover:not(:disabled),
#t-ok:hover:not(:disabled) {
    background: #991b1b;
    border-color: #991b1b
}

#e-ok,
#b-ok {
    color: #fff;
    background: var(--hijau);
    border-color: var(--hijau)
}

#e-ok:hover:not(:disabled),
#b-ok:hover:not(:disabled) {
    background: var(--hijau-tua);
    border-color: var(--hijau-tua)
}

/* Form Instansi: lebar rapi, tombol sejajar */
#sub-instansi {
    max-width: 640px;
    padding: 20px 24px;
    background: var(--putih);
    border: 1px solid var(--garis);
    border-radius: var(--radius)
}

#sub-instansi label input {
    width: 100%
}

#sub-instansi .form-baris {
    align-items: center;
    margin: 16px 0
}

#sub-instansi .form-baris .kecil {
    margin-right: auto
}

#p-simpan {
    min-width: 120px
}

/* Form kelas & kartu */
#sub-kelas .form-baris input,
#sub-kartu .form-baris input,
#sub-kartu .form-baris select {
    min-width: 200px
}

#sub-kelas .form-baris input[type=number] {
    min-width: 0
}

/* Modal: tombol seragam */
.kotak .aksi button {
    min-width: 90px
}

.kotak label {
    margin: 12px 0
}=== internal/service/ekspor_test.go ===
package service

import "testing"

func TestSusunLaporan(t *testing.T) {
	s := setupT5(t)
	a, _ := s.TambahKelas("VII A", "Ani", 1)
	b, _ := s.TambahKelas("VII B", "Budi", 2)
	_, _ = s.TambahKelas("VII C", "Cici", 3)
	harusOK(t, s.SubmitAmbil(a, 10))
	harusOK(t, s.SubmitKembali(a, 8))
	harusOK(t, s.SubmitAmbil(b, 5))
	harusOK(t, s.TutupHari(""))

	l, err := s.susunLaporan(tglUji, tglUji)
	harusOK(t, err)
	if len(l.Baris) != 2 {
		t.Fatalf("mau 2 baris (VII C TIDAK_AMBIL dikecualikan), dapat %d", len(l.Baris))
	}
	if x := l.Baris[0]; x.Keterangan != "Kurang 2 Ompreng" || !x.Merah || x.Tanggal != "12 Januari 2026" || x.WaktuAmbil != "10:00 WIB" {
		t.Fatalf("baris A salah: %+v", x)
	}
	if x := l.Baris[1]; x.Keterangan != "Belum kembali" || !x.Merah || x.JumlahKembali != "-" {
		t.Fatalf("baris B salah: %+v", x)
	}
	if l.Periode != "Hari/Tanggal: Senin, 12 Januari 2026" {
		t.Fatalf("periode: %q", l.Periode)
	}
	if _, err := s.susunLaporan("2026-02-01", "2026-02-02"); err == nil {
		t.Fatal("periode tanpa data seharusnya galat")
	}
	if _, err := s.susunLaporan("2026-02-02", "2026-02-01"); err == nil {
		t.Fatal("rentang terbalik seharusnya galat")
	}
}
=== frontend/js/api.js ===
// Binding Wails: window.go.<package>.<Type>.<Method>
const API = new Proxy({}, {
    get: (_, name) => (...args) => {
        const svc = window.go && window.go.service && window.go.service.Service;
        if (!svc || !svc[name]) return Promise.reject(new Error("Binding belum siap: " + name));
        return svc[name](...args);
    }
});

let _audio;
function bunyi(tipe) {
    try {
        _audio = _audio || new (window.AudioContext || window.webkitAudioContext)();
        const o = _audio.createOscillator(), g = _audio.createGain();
        o.frequency.value = tipe === "ok" ? 880 : tipe === "err" ? 220 : 520;
        g.gain.value = 0.15;
        o.connect(g); g.connect(_audio.destination);
        o.start(); o.stop(_audio.currentTime + (tipe === "err" ? 0.4 : 0.15));
    } catch (e) { }
}

let _toastT;
function toast(teks, tipe = "info") {
    const el = document.getElementById("toast");
    el.textContent = teks; el.className = tipe;
    bunyi(tipe);
    clearTimeout(_toastT);
    _toastT = setTimeout(() => el.className = "hidden", 2500);
}zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$

Tiga keputusan di rencana saya jalankan sesuai usulan, karena Anda mengirim berkas lanjutan tanpa keberatan: `KURANG` tidak mengaktifkan Cetak PDF, pengembalian pertama 0 diperbolehkan, dan cooldown setelah `KURANG` 5 detik. Satu penyesuaian dari rencana: `BELUM_KEMBALI` yang sudah punya pengembalian sebagian ditampilkan "Kurang N Ompreng", bukan "Belum lengkap, kurang N". Dengan begitu `ekspor_test.go` lulus tanpa diubah.

Saya belum menjalankan kodenya, jadi tolong jalankan `go test` di bawah dan kirim hasilnya bila ada yang gagal.

## 1. `internal/model/model.go` (timpa utuh)

```go
package model

type Kelas struct {
	ID        int64  `json:"id"`
	Nama      string `json:"nama"`
	NamaKetua string `json:"nama_ketua"`
	Urutan    int    `json:"urutan"`
	Aktif     bool   `json:"aktif"`
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
```

## 2. `internal/db/db.go` (timpa utuh)

```go
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite"
)

// DataDir: %APPDATA%\MBG (Windows), $XDG_DATA_HOME/mbg atau ~/.local/share/mbg (Linux).
func DataDir() string {
	if runtime.GOOS == "windows" {
		if a := os.Getenv("APPDATA"); a != "" {
			return filepath.Join(a, "MBG")
		}
	}
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "mbg")
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".local", "share", "mbg")
}

func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", filepath.ToSlash(path))
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1) // satu penulis, hindari SQLITE_BUSY
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	if err := Migrate(conn); err != nil {
		return nil, fmt.Errorf("migrasi: %w", err)
	}
	return conn, nil
}

func Migrate(conn *sql.DB) error {
	if _, err := conn.Exec(schema); err != nil {
		return err
	}
	for _, k := range []struct{ tabel, kolom string }{
		{"kartu", "nisn"}, {"kartu", "nama_siswa"},
		{"transaksi", "nisn_ambil"}, {"transaksi", "nama_ambil"},
		{"transaksi", "nisn_kembali"}, {"transaksi", "nama_kembali"},
	} {
		if err := tambahKolom(conn, k.tabel, k.kolom); err != nil {
			return err
		}
	}
	// Backfill: satu baris pengembalian per transaksi SELESAI lama yang belum punya rincian.
	// Idempoten berkat NOT EXISTS, aman dijalankan setiap aplikasi dibuka.
	if _, err := conn.Exec(`INSERT INTO pengembalian (transaksi_id,waktu,jumlah,nisn,nama_siswa)
		SELECT t.id,t.waktu_kembali,t.jumlah_kembali,t.nisn_kembali,t.nama_kembali FROM transaksi t
		WHERE t.status='SELESAI' AND t.jumlah_kembali IS NOT NULL AND t.waktu_kembali IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM pengembalian p WHERE p.transaksi_id=t.id)`); err != nil {
		return err
	}
	_, err := conn.Exec(`INSERT OR IGNORE INTO pengaturan (id, jabatan_penanggung_jawab) VALUES (1, 'Penanggung Jawab MBG')`)
	return err
}

func tambahKolom(conn *sql.DB, tabel, kolom string) error {
	rows, err := conn.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, tabel))
	if err != nil {
		return err
	}
	ada := false
	for rows.Next() {
		var cid, notnull, pk int
		var nama, tipe string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &nama, &tipe, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		if nama == kolom {
			ada = true
		}
	}
	rows.Close()
	if ada {
		return nil
	}
	_, err = conn.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s TEXT`, tabel, kolom))
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS pengaturan (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  nama_instansi TEXT, alamat_instansi TEXT,
  logo BLOB, logo_mime TEXT,
  penanggung_jawab TEXT, nip_penanggung_jawab TEXT,
  jabatan_penanggung_jawab TEXT,
  jeda_minimum_menit INTEGER NOT NULL DEFAULT 10,
  hid_ambang_ms INTEGER NOT NULL DEFAULT 50,
  port_serial TEXT
);
CREATE TABLE IF NOT EXISTS kelas (
  id INTEGER PRIMARY KEY,
  nama TEXT NOT NULL UNIQUE,
  nama_ketua TEXT NOT NULL,
  urutan INTEGER NOT NULL DEFAULT 0,
  aktif INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS kartu (
  uid TEXT PRIMARY KEY,
  kelas_id INTEGER NOT NULL REFERENCES kelas(id),
  aktif INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS transaksi (
  id INTEGER PRIMARY KEY,
  tanggal TEXT NOT NULL,
  kelas_id INTEGER NOT NULL REFERENCES kelas(id),
  nama_kelas TEXT NOT NULL,
  nama_ketua TEXT NOT NULL,
  waktu_ambil TEXT,
  waktu_kembali TEXT,
  jumlah_ambil INTEGER,
  jumlah_kembali INTEGER,
  status TEXT NOT NULL,
  catatan TEXT,
  UNIQUE (tanggal, kelas_id)
);
CREATE TABLE IF NOT EXISTS pengembalian (
  id INTEGER PRIMARY KEY,
  transaksi_id INTEGER NOT NULL REFERENCES transaksi(id) ON DELETE CASCADE,
  waktu TEXT NOT NULL,
  jumlah INTEGER NOT NULL,
  nisn TEXT,
  nama_siswa TEXT
);
CREATE INDEX IF NOT EXISTS idx_pengembalian_trx ON pengembalian(transaksi_id);
CREATE TABLE IF NOT EXISTS hari_ditutup (
  tanggal TEXT PRIMARY KEY,
  waktu_tutup TEXT NOT NULL,
  catatan TEXT
);
CREATE TABLE IF NOT EXISTS audit_log (
  id INTEGER PRIMARY KEY,
  waktu TEXT NOT NULL,
  aksi TEXT NOT NULL,
  entitas TEXT NOT NULL,
  entitas_id TEXT,
  nilai_lama TEXT,
  nilai_baru TEXT,
  alasan TEXT
);
`
```

## 3. `internal/service/transaksi.go` (timpa utuh)

```go
package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mbg/internal/model"
	"mbg/internal/reader"
	"mbg/internal/timeutil"
)

func itoa(i int64) string { return strconv.FormatInt(i, 10) }

const (
	debounceTap   = 2 * time.Second
	cooldownKelas = 30 * time.Second
	// cooldownSusulan: setelah pengembalian yang masih kurang (status KURANG), cooldown kelas
	// dipendekkan supaya teman sekelas yang menyusul membawa sisa ompreng tidak ditolak
	// "Baru saja tercatat" selama 30 detik. 5 detik tetap menahan tap ganda dari kartu yang menempel.
	cooldownSusulan = 5 * time.Second
)

func (s *Service) jedaMenit() int {
	var m int
	if err := s.db.QueryRow(`SELECT jeda_minimum_menit FROM pengaturan WHERE id=1`).Scan(&m); err != nil {
		return 10
	}
	return m
}

func pesan(tipe, teks string) model.TapResult {
	return model.TapResult{Aksi: "PESAN", Tipe: tipe, Pesan: teks}
}

// ProsesTap: satu pintu untuk semua sumber (serial, HID, manual, simulator).
func (s *Service) ProsesTap(rawUID string) model.TapResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	uid, ok := reader.NormalisasiUID(rawUID)
	if !ok {
		return pesan("err", "UID tidak valid")
	}
	now := s.now()

	if s.pending != nil {
		return pesan("info", "Selesaikan input kelas "+s.pending.Nama+" dulu")
	}
	if uid == s.lastUID && now.Sub(s.lastUIDAt) < debounceTap {
		return model.TapResult{Aksi: "ABAIKAN"}
	}
	s.lastUID, s.lastUIDAt = uid, now

	var k model.Kelas
	var kartuAktif, kelasAktif int
	var nisn, namaSiswa string
	err := s.db.QueryRow(`SELECT k.id,k.nama,k.nama_ketua,c.aktif,k.aktif,COALESCE(c.nisn,''),COALESCE(c.nama_siswa,'') FROM kartu c JOIN kelas k ON k.id=c.kelas_id WHERE c.uid=?`, uid).
		Scan(&k.ID, &k.Nama, &k.NamaKetua, &kartuAktif, &kelasAktif, &nisn, &namaSiswa)
	if err != nil || kartuAktif != 1 || kelasAktif != 1 {
		return pesan("err", "Kartu tidak terdaftar")
	}

	tgl := timeutil.Tanggal(now)
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n)
	if n > 0 {
		return pesan("err", "Hari ini sudah ditutup")
	}

	base := model.TapResult{KelasID: k.ID, NamaKelas: k.Nama, NamaKetua: k.NamaKetua, Tipe: "ok", NISN: nisn, NamaSiswa: namaSiswa}
	var status, waktuAmbil string
	var jAmbil, jKembali int
	adaTrx := true
	err = s.db.QueryRow(`SELECT status,COALESCE(waktu_ambil,''),COALESCE(jumlah_ambil,0),COALESCE(jumlah_kembali,0) FROM transaksi WHERE tanggal=? AND kelas_id=?`, tgl, k.ID).
		Scan(&status, &waktuAmbil, &jAmbil, &jKembali)
	if errors.Is(err, sql.ErrNoRows) {
		adaTrx = false
	} else if err != nil {
		return pesan("err", "Kesalahan database")
	}

	// Penolakan final ditempatkan sebelum cooldown agar pesannya selalu jelas,
	// untuk kartu siswa mana pun di kelas yang sama.
	if adaTrx && status == "SELESAI" {
		if jKembali >= jAmbil {
			return pesan("info", "Kelas "+k.Nama+" sudah mengembalikan ompreng dengan lengkap")
		}
		return pesan("info", "Kelas "+k.Nama+" sudah selesai hari ini") // data lama (SELESAI kurang) dianggap final
	}
	if adaTrx && status != "AMBIL" && status != "KURANG" {
		return pesan("info", "Kelas "+k.Nama+" sudah selesai hari ini")
	}

	if t, ok := s.lastSubmit[k.ID]; ok && now.Sub(t) < cooldownKelas {
		return pesan("info", "Baru saja tercatat")
	}

	if !adaTrx {
		s.pending = &pendingInput{k.ID, k.Nama}
		s.siswaNISN, s.siswaNama = nisn, namaSiswa
		base.Aksi = "AMBIL"
		return base
	}

	base.JumlahAmbil = jAmbil
	s.pending = &pendingInput{k.ID, k.Nama}
	s.siswaNISN, s.siswaNama = nisn, namaSiswa

	if status == "KURANG" {
		// Susulan: tanpa peringatan jeda minimum.
		base.Aksi = "KEMBALI_SUSULAN"
		base.Sisa = jAmbil - jKembali
		return base
	}

	if t, err := time.ParseInLocation("2006-01-02 15:04", tgl+" "+waktuAmbil, timeutil.WIB); err == nil {
		menit := int(now.Sub(t).Minutes())
		if menit < s.jedaMenit() {
			base.Aksi = "KONFIRMASI_KEMBALI"
			base.MenitSejak = menit
			return base
		}
	}
	base.Aksi = "KEMBALI"
	return base
}

// BatalInput: dipanggil UI saat popup/konfirmasi dibatalkan, melepas penguncian.
func (s *Service) BatalInput() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending = nil
	s.siswaNISN, s.siswaNama = "", ""
}

func (s *Service) SubmitAmbil(kelasID int64, jumlah int) error {
	if jumlah < 1 {
		return errors.New("Jumlah ambil harus bilangan bulat minimal 1")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var nama, ketua string
	if err := tx.QueryRow(`SELECT nama,nama_ketua FROM kelas WHERE id=?`, kelasID).Scan(&nama, &ketua); err != nil {
		return errors.New("Kelas tidak ditemukan")
	}
	res, err := tx.Exec(`INSERT INTO transaksi (tanggal,kelas_id,nama_kelas,nama_ketua,waktu_ambil,jumlah_ambil,status,nisn_ambil,nama_ambil) VALUES (?,?,?,?,?,?,'AMBIL',?,?)`,
		timeutil.Tanggal(now), kelasID, nama, ketua, timeutil.Jam(now), jumlah, s.siswaNISN, s.siswaNama)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("Kelas ini sudah tercatat mengambil hari ini")
		}
		return err
	}
	id, _ := res.LastInsertId()
	if err := s.audit(tx, now, "TAMBAH", "transaksi", itoa(id), nil,
		map[string]any{"kelas": nama, "jumlah_ambil": jumlah, "status": "AMBIL", "pengambil": s.siswaNama}, ""); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.lastSubmit[kelasID] = now
	s.pending = nil
	s.siswaNISN, s.siswaNama = "", ""
	return nil
}

// namaPengembali: nama unik semua pengembali (urut waktu), dipisah koma.
func namaPengembali(tx *sql.Tx, id int64) (string, error) {
	rows, err := tx.Query(`SELECT COALESCE(nama_siswa,'') FROM pengembalian WHERE transaksi_id=? ORDER BY id`, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var out []string
	ada := map[string]bool{}
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return "", err
		}
		n = strings.TrimSpace(n)
		if n == "" || ada[n] {
			continue
		}
		ada[n] = true
		out = append(out, n)
	}
	return strings.Join(out, ", "), rows.Err()
}

// SubmitKembali: pengembalian awal (status AMBIL, 0..jumlah_ambil) atau susulan (status KURANG, 1..sisa).
// Setiap pemanggilan menambah satu baris pengembalian; transaksi menyimpan total akumulasi.
func (s *Service) SubmitKembali(kelasID int64, jumlah int) error {
	if jumlah < 0 {
		return errors.New("Jumlah kembali tidak boleh negatif")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int64
	var status, nama string
	var jAmbil, jKembali int
	err = tx.QueryRow(`SELECT id,status,nama_kelas,COALESCE(jumlah_ambil,0),COALESCE(jumlah_kembali,0) FROM transaksi WHERE tanggal=? AND kelas_id=?`,
		timeutil.Tanggal(now), kelasID).Scan(&id, &status, &nama, &jAmbil, &jKembali)
	if err != nil {
		return errors.New("Belum ada pengambilan hari ini untuk kelas ini")
	}
	susulan := false
	switch status {
	case "AMBIL":
		jKembali = 0
		if jumlah > jAmbil {
			return fmt.Errorf("Jumlah kembali (%d) tidak boleh melebihi jumlah ambil (%d)", jumlah, jAmbil)
		}
	case "KURANG":
		susulan = true
		sisa := jAmbil - jKembali
		if jumlah < 1 {
			return fmt.Errorf("Jumlah susulan harus minimal 1 (sisa %d)", sisa)
		}
		if jumlah > sisa {
			return fmt.Errorf("Jumlah susulan (%d) tidak boleh melebihi sisa (%d)", jumlah, sisa)
		}
	default:
		return errors.New("Transaksi kelas ini sudah selesai")
	}

	total := jKembali + jumlah
	statusBaru := "KURANG"
	if total >= jAmbil {
		statusBaru = "SELESAI"
	}
	jam := timeutil.Jam(now)
	if _, err := tx.Exec(`INSERT INTO pengembalian (transaksi_id,waktu,jumlah,nisn,nama_siswa) VALUES (?,?,?,?,?)`,
		id, jam, jumlah, s.siswaNISN, s.siswaNama); err != nil {
		return err
	}
	namaGabung, err := namaPengembali(tx, id)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE transaksi SET waktu_kembali=?,jumlah_kembali=?,status=?,nisn_kembali=?,nama_kembali=? WHERE id=?`,
		jam, total, statusBaru, s.siswaNISN, namaGabung, id); err != nil {
		return err
	}

	lama := map[string]any{"status": status}
	baru := map[string]any{"status": statusBaru, "jumlah_kembali": total, "pengembali": s.siswaNama}
	alasan := ""
	if susulan {
		lama["jumlah_kembali"] = jKembali
		baru["jumlah_susulan"] = jumlah
		alasan = "pengembalian susulan"
	}
	if err := s.audit(tx, now, "UBAH", "transaksi", itoa(id), lama, baru, alasan); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if statusBaru == "KURANG" {
		// dimundurkan agar efektif hanya cooldownSusulan (lihat komentar konstanta)
		s.lastSubmit[kelasID] = now.Add(-(cooldownKelas - cooldownSusulan))
	} else {
		s.lastSubmit[kelasID] = now
	}
	s.pending = nil
	s.siswaNISN, s.siswaNama = "", ""
	return nil
}

// StatusHariIni: data dashboard (keterangan dihitung saat tampil).
func (s *Service) StatusHariIni() (model.StatusHari, error) {
	now := s.now()
	tgl := timeutil.Tanggal(now)
	out := model.StatusHari{Tanggal: tgl, TanggalIndo: timeutil.HariTanggalIndo(now), Baris: []model.BarisHari{}}

	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n)
	out.Ditutup = n > 0

	rows, err := s.db.Query(`SELECT k.id,k.nama,k.nama_ketua,t.status,t.jumlah_ambil,t.jumlah_kembali,
		COALESCE(t.waktu_ambil,''),COALESCE(t.waktu_kembali,''),
		(SELECT COUNT(*) FROM pengembalian p WHERE p.transaksi_id=t.id)
		FROM kelas k LEFT JOIN transaksi t ON t.kelas_id=k.id AND t.tanggal=?
		WHERE k.aktif=1 ORDER BY k.urutan,k.nama`, tgl)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	semua := true
	for rows.Next() {
		var b model.BarisHari
		var st sql.NullString
		var ja, jk sql.NullInt64
		var np int
		if err := rows.Scan(&b.KelasID, &b.Nama, &b.NamaKetua, &st, &ja, &jk, &b.WaktuAmbil, &b.WaktuKembali, &np); err != nil {
			return out, err
		}
		b.Status = "BELUM_AMBIL"
		if st.Valid {
			b.Status = st.String
		}
		if ja.Valid {
			v := int(ja.Int64)
			b.JumlahAmbil = &v
		}
		if jk.Valid {
			v := int(jk.Int64)
			b.JumlahKembali = &v
		}
		b.Keterangan, b.Merah = Keterangan(b.Status, b.JumlahAmbil, b.JumlahKembali, "", np > 1)
		if b.Status != "SELESAI" {
			semua = false
		}
		out.Baris = append(out.Baris, b)
	}
	out.SemuaSelesai = semua && len(out.Baris) > 0
	out.BolehCetak = out.Ditutup || out.SemuaSelesai
	return out, rows.Err()
}

// Keterangan: aturan 4.3. Dipakai dashboard, riwayat, dan ekspor.
// susulan = true bila transaksi punya lebih dari satu pengembalian.
func Keterangan(status string, ambil, kembali *int, catatan string, susulan bool) (string, bool) {
	var ket string
	merah := false
	kurang := ""
	if ambil != nil && kembali != nil && *kembali < *ambil {
		kurang = fmt.Sprintf("Kurang %d Ompreng", *ambil-*kembali)
	}
	switch status {
	case "SELESAI":
		switch {
		case kurang != "":
			ket, merah = kurang, true
		case susulan:
			ket = "Lengkap (susulan)"
		default:
			ket = "Lengkap"
		}
	case "KURANG":
		ket, merah = kurang, true
		if ket == "" {
			ket = "Kurang"
		}
	case "BELUM_KEMBALI":
		ket, merah = "Belum kembali", true
		if kurang != "" {
			ket = kurang // sudah ada pengembalian sebagian sebelum hari ditutup
		}
	case "TIDAK_AMBIL":
		ket, merah = "Tidak mengambil", true
	}
	if catatan != "" {
		if ket != "" {
			ket += " - "
		}
		ket += catatan
	}
	return ket, merah
}
```

## 4. `internal/service/hari.go` (timpa utuh)

```go
package service

import (
	"errors"
	"strings"

	"mbg/internal/timeutil"
)

// TutupHari: kelas AMBIL/KURANG -> BELUM_KEMBALI, kelas aktif tanpa transaksi -> TIDAK_AMBIL.
func (s *Service) TutupHari(catatan string) error {
	catatan = strings.TrimSpace(catatan)
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tgl := timeutil.Tanggal(now)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return errors.New("Hari ini sudah ditutup")
	}
	r1, err := tx.Exec(`UPDATE transaksi SET status='BELUM_KEMBALI' WHERE tanggal=? AND status IN ('AMBIL','KURANG')`, tgl)
	if err != nil {
		return err
	}
	belum, _ := r1.RowsAffected()
	r2, err := tx.Exec(`INSERT INTO transaksi (tanggal,kelas_id,nama_kelas,nama_ketua,status)
		SELECT ?,k.id,k.nama,k.nama_ketua,'TIDAK_AMBIL' FROM kelas k
		WHERE k.aktif=1 AND NOT EXISTS (SELECT 1 FROM transaksi t WHERE t.tanggal=? AND t.kelas_id=k.id)`, tgl, tgl)
	if err != nil {
		return err
	}
	tidak, _ := r2.RowsAffected()
	if _, err := tx.Exec(`INSERT INTO hari_ditutup (tanggal,waktu_tutup,catatan) VALUES (?,?,?)`,
		tgl, timeutil.Jam(now), catatan); err != nil {
		return err
	}
	if err := s.audit(tx, now, "TUTUP_HARI", "hari", tgl, nil,
		map[string]any{"belum_kembali": belum, "tidak_ambil": tidak}, catatan); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.pending = nil
	return nil
}

// BukaHari: hanya hari ini. Status otomatis dikembalikan supaya tap bisa dilanjutkan:
// BELUM_KEMBALI tanpa pengembalian -> AMBIL, BELUM_KEMBALI dengan pengembalian sebagian -> KURANG,
// baris TIDAK_AMBIL otomatis dihapus. Koreksi hari lampau dilakukan lewat Riwayat (edit/hapus).
func (s *Service) BukaHari(tanggal, alasan string) error {
	if strings.TrimSpace(alasan) == "" {
		return errors.New("Alasan wajib diisi")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if tanggal != timeutil.Tanggal(now) {
		return errors.New("Hanya hari ini yang bisa dibuka kembali; koreksi hari sebelumnya lewat menu Riwayat")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tanggal).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return errors.New("Hari ini tidak dalam keadaan ditutup")
	}
	r1, err := tx.Exec(`UPDATE transaksi SET status='AMBIL' WHERE tanggal=? AND status='BELUM_KEMBALI' AND jumlah_kembali IS NULL`, tanggal)
	if err != nil {
		return err
	}
	r3, err := tx.Exec(`UPDATE transaksi SET status='KURANG' WHERE tanggal=? AND status='BELUM_KEMBALI' AND jumlah_kembali IS NOT NULL AND jumlah_kembali < jumlah_ambil`, tanggal)
	if err != nil {
		return err
	}
	r2, err := tx.Exec(`DELETE FROM transaksi WHERE tanggal=? AND status='TIDAK_AMBIL' AND jumlah_ambil IS NULL`, tanggal)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM hari_ditutup WHERE tanggal=?`, tanggal); err != nil {
		return err
	}
	a, _ := r1.RowsAffected()
	k, _ := r3.RowsAffected()
	b, _ := r2.RowsAffected()
	if err := s.audit(tx, now, "BUKA_HARI", "hari", tanggal, nil,
		map[string]any{"dikembalikan_ke_ambil": a, "dikembalikan_ke_kurang": k, "baris_tidak_ambil_dihapus": b}, alasan); err != nil {
		return err
	}
	return tx.Commit()
}

// ApakahBolehCetak: hari ditutup, atau semua kelas aktif berstatus SELESAI pada tanggal itu.
// Status KURANG belum dihitung selesai.
func (s *Service) ApakahBolehCetak(tanggal string) (bool, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tanggal).Scan(&n); err != nil {
		return false, err
	}
	if n > 0 {
		return true, nil
	}
	var aktif, selesai int
	err := s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(CASE WHEN t.status='SELESAI' THEN 1 ELSE 0 END),0)
		FROM kelas k LEFT JOIN transaksi t ON t.kelas_id=k.id AND t.tanggal=? WHERE k.aktif=1`, tanggal).Scan(&aktif, &selesai)
	if err != nil {
		return false, err
	}
	return aktif > 0 && aktif == selesai, nil
}
```

## 5. `internal/service/riwayat.go` (timpa utuh)

```go
package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"mbg/internal/model"
)

const kolomTrx = `id,tanggal,kelas_id,nama_kelas,nama_ketua,COALESCE(waktu_ambil,''),COALESCE(waktu_kembali,''),jumlah_ambil,jumlah_kembali,status,COALESCE(catatan,''),COALESCE(nama_ambil,''),COALESCE(nama_kembali,''),(SELECT COUNT(*) FROM pengembalian p WHERE p.transaksi_id=transaksi.id)`

type pemindai interface{ Scan(dest ...any) error }

type penanya interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

func scanTrx(r pemindai) (model.Transaksi, error) {
	var t model.Transaksi
	var ja, jk sql.NullInt64
	var np int
	if err := r.Scan(&t.ID, &t.Tanggal, &t.KelasID, &t.NamaKelas, &t.NamaKetua, &t.WaktuAmbil, &t.WaktuKembali,
		&ja, &jk, &t.Status, &t.Catatan, &t.NamaAmbil, &t.NamaKembali, &np); err != nil {
		return t, err
	}
	if ja.Valid {
		v := int(ja.Int64)
		t.JumlahAmbil = &v
	}
	if jk.Valid {
		v := int(jk.Int64)
		t.JumlahKembali = &v
	}
	t.Susulan = np > 1
	t.Keterangan, t.Merah = Keterangan(t.Status, t.JumlahAmbil, t.JumlahKembali, t.Catatan, t.Susulan)
	return t, nil
}

// muatRincian: rincian pengembalian untuk sekumpulan transaksi, satu query.
func muatRincian(q penanya, ids []int64) (map[int64][]model.Pengembalian, error) {
	out := map[int64][]model.Pengembalian{}
	if len(ids) == 0 {
		return out, nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := q.Query(`SELECT transaksi_id,waktu,jumlah,COALESCE(nisn,''),COALESCE(nama_siswa,'') FROM pengembalian
		WHERE transaksi_id IN (`+ph+`) ORDER BY transaksi_id,id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tid int64
		var p model.Pengembalian
		if err := rows.Scan(&tid, &p.Waktu, &p.Jumlah, &p.NISN, &p.NamaSiswa); err != nil {
			return nil, err
		}
		out[tid] = append(out[tid], p)
	}
	return out, rows.Err()
}

func nilInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func nilStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func jamValid(s string) bool { _, err := time.Parse("15:04", s); return err == nil }

func validasiEdit(e model.EditTransaksi) error {
	switch e.Status {
	case "AMBIL", "KURANG", "SELESAI", "BELUM_KEMBALI", "TIDAK_AMBIL":
	default:
		return errors.New("Status tidak valid")
	}
	if e.Status == "TIDAK_AMBIL" {
		return nil
	}
	if e.JumlahAmbil == nil || *e.JumlahAmbil < 1 {
		return errors.New("Jumlah ambil harus bilangan bulat minimal 1")
	}
	if !jamValid(e.WaktuAmbil) {
		return errors.New("Waktu ambil harus berformat HH:MM")
	}
	if e.Status == "SELESAI" || e.Status == "KURANG" {
		label := "Selesai"
		if e.Status == "KURANG" {
			label = "Kurang"
		}
		if e.JumlahKembali == nil {
			return fmt.Errorf("Status %s wajib punya jumlah kembali", label)
		}
		if *e.JumlahKembali < 0 {
			return errors.New("Jumlah kembali tidak boleh negatif")
		}
		if *e.JumlahKembali > *e.JumlahAmbil {
			return fmt.Errorf("Jumlah kembali (%d) tidak boleh melebihi jumlah ambil (%d)", *e.JumlahKembali, *e.JumlahAmbil)
		}
		if e.Status == "KURANG" && *e.JumlahKembali >= *e.JumlahAmbil {
			return errors.New("Status Kurang mensyaratkan jumlah kembali lebih kecil dari jumlah ambil")
		}
		if !jamValid(e.WaktuKembali) {
			return errors.New("Waktu kembali harus berformat HH:MM")
		}
	}
	return nil
}

func (s *Service) ListTransaksi(f model.FilterTransaksi) ([]model.Transaksi, error) {
	q := `SELECT ` + kolomTrx + ` FROM transaksi WHERE 1=1`
	var args []any
	if f.Dari != "" {
		q += ` AND tanggal>=?`
		args = append(args, f.Dari)
	}
	if f.Sampai != "" {
		q += ` AND tanggal<=?`
		args = append(args, f.Sampai)
	}
	if f.KelasID > 0 {
		q += ` AND kelas_id=?`
		args = append(args, f.KelasID)
	}
	if f.Status != "" {
		q += ` AND status=?`
		args = append(args, f.Status)
	}
	q += ` ORDER BY tanggal DESC, (SELECT urutan FROM kelas k WHERE k.id=transaksi.kelas_id), nama_kelas`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Transaksi{}
	var ids []int64
	for rows.Next() {
		t, err := scanTrx(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
		ids = append(ids, t.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close() // koneksi tunggal: lepas sebelum query berikutnya
	rm, err := muatRincian(s.db, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Pengembalian = rm[out[i].ID]
	}
	return out, nil
}

func (s *Service) UbahTransaksi(e model.EditTransaksi) error {
	if strings.TrimSpace(e.Alasan) == "" {
		return errors.New("Alasan perubahan wajib diisi")
	}
	if err := validasiEdit(e); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	lama, err := scanTrx(tx.QueryRow(`SELECT `+kolomTrx+` FROM transaksi WHERE id=?`, e.ID))
	if err != nil {
		return errors.New("Transaksi tidak ditemukan")
	}
	rm, err := muatRincian(tx, []int64{e.ID})
	if err != nil {
		return err
	}
	lama.Pengembalian = rm[e.ID]

	ja, jk, wa, wk := e.JumlahAmbil, e.JumlahKembali, e.WaktuAmbil, e.WaktuKembali
	switch e.Status {
	case "AMBIL", "BELUM_KEMBALI":
		jk, wk = nil, ""
	case "TIDAK_AMBIL":
		ja, jk, wa, wk = nil, nil, "", ""
	}
	catatan := strings.TrimSpace(e.Catatan)
	if _, err := tx.Exec(`UPDATE transaksi SET status=?,jumlah_ambil=?,jumlah_kembali=?,waktu_ambil=?,waktu_kembali=?,catatan=? WHERE id=?`,
		e.Status, nilInt(ja), nilInt(jk), nilStr(wa), nilStr(wk), nilStr(catatan), e.ID); err != nil {
		return err
	}

	// Sinkronkan rincian pengembalian dengan total hasil edit.
	if (e.Status == "SELESAI" || e.Status == "KURANG") && jk != nil {
		var n, sum int
		if err := tx.QueryRow(`SELECT COUNT(*),COALESCE(SUM(jumlah),0) FROM pengembalian WHERE transaksi_id=?`, e.ID).Scan(&n, &sum); err != nil {
			return err
		}
		if n == 0 || sum != *jk {
			if _, err := tx.Exec(`DELETE FROM pengembalian WHERE transaksi_id=?`, e.ID); err != nil {
				return err
			}
			if _, err := tx.Exec(`INSERT INTO pengembalian (transaksi_id,waktu,jumlah,nisn,nama_siswa) VALUES (?,?,?,NULL,?)`,
				e.ID, wk, *jk, lama.NamaKembali); err != nil {
				return err
			}
		}
	} else {
		if _, err := tx.Exec(`DELETE FROM pengembalian WHERE transaksi_id=?`, e.ID); err != nil {
			return err
		}
	}

	baru := map[string]any{"status": e.Status, "jumlah_ambil": nilInt(ja), "jumlah_kembali": nilInt(jk),
		"waktu_ambil": wa, "waktu_kembali": wk, "catatan": catatan}
	if err := s.audit(tx, s.now(), "UBAH", "transaksi", itoa(e.ID), lama, baru, strings.TrimSpace(e.Alasan)); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) HapusTransaksi(id int64, alasan string) error {
	alasan = strings.TrimSpace(alasan)
	if alasan == "" {
		return errors.New("Alasan penghapusan wajib diisi")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	lama, err := scanTrx(tx.QueryRow(`SELECT `+kolomTrx+` FROM transaksi WHERE id=?`, id))
	if err != nil {
		return errors.New("Transaksi tidak ditemukan")
	}
	rm, err := muatRincian(tx, []int64{id})
	if err != nil {
		return err
	}
	lama.Pengembalian = rm[id]
	// rincian pengembalian ikut terhapus lewat ON DELETE CASCADE
	if _, err := tx.Exec(`DELETE FROM transaksi WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "HAPUS", "transaksi", itoa(id), lama, nil, alasan); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	delete(s.lastSubmit, lama.KelasID)
	return nil
}

func (s *Service) ListAudit(f model.FilterAudit) ([]model.Audit, error) {
	q := `SELECT id,waktu,aksi,entitas,COALESCE(entitas_id,''),COALESCE(nilai_lama,''),COALESCE(nilai_baru,''),COALESCE(alasan,'') FROM audit_log WHERE 1=1`
	var args []any
	if f.Dari != "" {
		q += ` AND substr(waktu,1,10)>=?`
		args = append(args, f.Dari)
	}
	if f.Sampai != "" {
		q += ` AND substr(waktu,1,10)<=?`
		args = append(args, f.Sampai)
	}
	if f.Entitas != "" {
		q += ` AND entitas=?`
		args = append(args, f.Entitas)
	}
	lim := f.Limit
	if lim <= 0 || lim > 1000 {
		lim = 200
	}
	q += fmt.Sprintf(` ORDER BY id DESC LIMIT %d`, lim)
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Audit{}
	for rows.Next() {
		var a model.Audit
		if err := rows.Scan(&a.ID, &a.Waktu, &a.Aksi, &a.Entitas, &a.EntitasID, &a.NilaiLama, &a.NilaiBaru, &a.Alasan); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
```

## 6. `internal/service/ekspor.go`: tiga penggantian kecil

Di fungsi `susunLaporan`:

**a.** Pada query laporan, ganti `COALESCE(nama_kembali,'') FROM transaksi` menjadi:

```go
COALESCE(nama_kembali,''),(SELECT COUNT(*) FROM pengembalian p WHERE p.transaksi_id=transaksi.id) FROM transaksi
```

**b.** Di dalam `for rows.Next()`, tambahkan `var np int` tepat di bawah `var ja, jk sql.NullInt64`, dan ubah Scan menjadi:

```go
if err := rows.Scan(&tgl, &kelas, &ketua, &wa, &wk, &ja, &jk, &status, &cat, &pengambil, &pengembali, &np); err != nil {
```

**c.** Ganti pemanggilan keterangan:

```go
ket, merah := Keterangan(status, pja, pjk, cat, np > 1)
```

Kolom Pengembali sudah memakai `nama_kembali`, yang kini berisi semua nama dipisah koma, dan waktu kembali memakai waktu pengembalian terakhir. Berkas `export/*` tidak berubah.

## 7. Frontend

**`frontend/js/dashboard.js`** (timpa utuh):

```js
const Dash = {
    async muat() {
        const st = await API.StatusHariIni();
        document.getElementById("info-tanggal").textContent = st.tanggal_indo + (st.ditutup ? " (DITUTUP)" : "");
        const tb = document.querySelector("#tabel-hari tbody");
        tb.innerHTML = "";
        const label = { BELUM_AMBIL: "Belum ambil", AMBIL: "Sudah ambil", KURANG: "Kurang", SELESAI: "Selesai", BELUM_KEMBALI: "Belum kembali", TIDAK_AMBIL: "Tidak mengambil" };
        let selesai = 0;
        st.baris.forEach((b, i) => {
            if (b.status === "SELESAI") selesai++;
            const tr = document.createElement("tr");
            if (b.merah) tr.className = "merah"; else if (b.status === "SELESAI") tr.className = "selesai";
            const v = x => (x === null || x === undefined) ? "-" : x;
            [i + 1, b.nama, b.nama_ketua, label[b.status] || b.status, v(b.jumlah_ambil), v(b.jumlah_kembali),
            b.waktu_ambil || "-", b.waktu_kembali || "-", b.keterangan].forEach(t => {
                const td = document.createElement("td"); td.textContent = t; tr.appendChild(td);
            });
            tb.appendChild(tr);
        });
        document.getElementById("ringkasan").textContent = `${selesai} dari ${st.baris.length} kelas selesai.`;
        Dash.terapkanHari(st);
    },

    // ---- popup ----
    tutupModal() { document.getElementById("modal").className = "hidden"; },
    bukaModal(html) { const m = document.getElementById("modal"); document.getElementById("modal-isi").innerHTML = html; m.className = ""; },

    // mode: AMBIL | KEMBALI | SUSULAN
    popupJumlah(r, mode) {
        const ambil = mode === "AMBIL";
        const susulan = mode === "SUSULAN";
        const sudah = (r.jumlah_ambil || 0) - (r.sisa || 0);
        const judul = ambil ? "Pengambilan" : susulan ? "Pengembalian susulan" : "Pengembalian";
        const info = ambil ? "" : susulan
            ? ` &middot; Diambil: <b>${r.jumlah_ambil}</b> &middot; Sudah kembali: <b>${sudah}</b> &middot; Sisa: <b>${r.sisa}</b>`
            : ` &middot; Diambil: <b>${r.jumlah_ambil}</b> ompreng`;
        const min = (ambil || susulan) ? 1 : 0;
        const max = susulan ? r.sisa : r.jumlah_ambil;
        const nama = ambil ? "diambil" : susulan ? "susulan" : "dikembalikan";
        this.bukaModal(`
      <h2>${judul} - ${esc(r.nama_kelas)}</h2>
      <p>Ketua: ${esc(r.nama_ketua)}${info}</p>
      <p>${ambil ? "Pengambil" : "Pengembali"}: <b>${esc(r.nama_siswa || "(kartu tanpa nama)")}</b>${r.nisn ? " &middot; NISN " + esc(r.nisn) : ""}</p>
      <input id="in-jumlah" type="number" min="${min}" ${ambil ? "" : `max="${max}"`} placeholder="Jumlah ${nama}">
      <div class="galat" id="in-galat"></div>
      <div class="aksi"><button id="in-batal">Batal</button><button id="in-ok">Submit</button></div>`);
        const inp = document.getElementById("in-jumlah"); inp.focus();
        const kirim = async () => {
            const val = inp.value.trim();
            if (!/^\d+$/.test(val)) { document.getElementById("in-galat").textContent = "Isi dengan bilangan bulat"; return; }
            const n = parseInt(val, 10);
            if (susulan && (n < 1 || n > r.sisa)) { document.getElementById("in-galat").textContent = `Jumlah susulan harus 1 sampai ${r.sisa}`; return; }
            try {
                if (ambil) await API.SubmitAmbil(r.kelas_id, n);
                else await API.SubmitKembali(r.kelas_id, n);
                this.tutupModal();
                toast(`${r.nama_kelas}: ${ambil ? "pengambilan" : susulan ? "pengembalian susulan" : "pengembalian"} tercatat`, "ok");
                this.muat();
            } catch (err) {
                document.getElementById("in-galat").textContent = String(err.message || err);
                bunyi("err");
            }
        };
        document.getElementById("in-ok").onclick = kirim;
        inp.onkeydown = e => { if (e.key === "Enter") kirim(); };
        document.getElementById("in-batal").onclick = async () => { await API.BatalInput(); this.tutupModal(); };
    },

    popupKonfirmasi(r) {
        this.bukaModal(`
      <h2>Konfirmasi</h2>
      <p>Kelas <b>${esc(r.nama_kelas)}</b> baru mengambil ${r.menit_sejak} menit lalu. Pengembali: <b>${esc(r.nama_siswa || "-")}</b>. Tetap proses pengembalian?</p>
      <div class="aksi"><button id="kf-tidak">Tidak</button><button id="kf-ya">Ya</button></div>`);
        document.getElementById("kf-tidak").onclick = async () => { await API.BatalInput(); this.tutupModal(); };
        document.getElementById("kf-ya").onclick = () => this.popupJumlah(r, "KEMBALI");
    },
};

function esc(s) { return String(s).replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c])); }

async function prosesTap(uid) {
    try {
        const r = await API.ProsesTap(uid);
        if (r.aksi === "ABAIKAN") return;
        if (r.aksi === "PESAN") return toast(r.pesan, r.tipe);
        bunyi("ok");
        if (r.nama_siswa) toast(r.nama_siswa + " (" + r.nama_kelas + ")", "info");
        if (r.aksi === "AMBIL") Dash.popupJumlah(r, "AMBIL");
        else if (r.aksi === "KEMBALI") Dash.popupJumlah(r, "KEMBALI");
        else if (r.aksi === "KEMBALI_SUSULAN") Dash.popupJumlah(r, "SUSULAN");
        else if (r.aksi === "KONFIRMASI_KEMBALI") Dash.popupKonfirmasi(r);
    } catch (err) { toast(String(err.message || err), "err"); }
}
```

**`frontend/js/riwayat.js`**, dua perubahan:

1. Di `dialogUbah`, tambahkan `KURANG` pada daftar opsi:

```js
const opsi = ["AMBIL", "KURANG", "SELESAI", "BELUM_KEMBALI", "TIDAK_AMBIL"].map(s => `<option ${s === r.status ? "selected" : ""}>${s}</option>`).join("");
```

2. Di `muatTrx`, tepat setelah blok `.forEach(t => { ... });` yang membuat sel-sel data dan sebelum `const td = document.createElement("td"); td.className = "aksi-sel";`, sisipkan:

```js
                if (r.pengembalian && r.pengembalian.length > 1) {
                    const d = document.createElement("div");
                    d.className = "kecil";
                    d.textContent = r.pengembalian.map(p => `${p.waktu} • ${p.jumlah} • ${p.nama_siswa || "-"}`).join("  |  ");
                    tr.lastElementChild.appendChild(d);
                }
```

**`frontend/index.html`**: di `#rf-status`, tambahkan setelah opsi `AMBIL`:

```html
                        <option value="KURANG">KURANG</option>
```

## 8. Tes

**`internal/service/service_test.go`** (timpa utuh): `TestAlurTap` disesuaikan karena `SubmitKembali(30)` dari ambil 31 kini berstatus `KURANG`.

```go
package service

import (
	"path/filepath"
	"testing"
	"time"

	"mbg/internal/db"
	"mbg/internal/timeutil"
)

func TestAlurTap(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	now := time.Date(2026, 1, 12, 11, 0, 0, 0, timeutil.WIB)
	s := NewWithClock(conn, t.TempDir(), func() time.Time { return now })

	id, err := s.TambahKelas("VII A", "Budi", 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.DaftarkanKartu("12345678", id, "", ""); err != nil {
		t.Fatal(err)
	}

	if r := s.ProsesTap("0012345678"); r.Aksi != "AMBIL" {
		t.Fatalf("tap 1: %+v", r)
	}
	// popup terbuka -> tap lain ditolak
	if r := s.ProsesTap("0099999999"); r.Aksi != "PESAN" {
		t.Fatalf("penguncian popup: %+v", r)
	}
	if err := s.SubmitAmbil(id, 0); err == nil {
		t.Fatal("jumlah 0 harus ditolak")
	}
	if err := s.SubmitAmbil(id, 31); err != nil {
		t.Fatal(err)
	}

	now = now.Add(5 * time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "KONFIRMASI_KEMBALI" {
		t.Fatalf("jeda minimum: %+v", r)
	}
	s.BatalInput()

	now = now.Add(15 * time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "KEMBALI" || r.JumlahAmbil != 31 {
		t.Fatalf("kembali: %+v", r)
	}
	if err := s.SubmitKembali(id, 32); err == nil {
		t.Fatal("kembali > ambil harus ditolak")
	}
	if err := s.SubmitKembali(id, 30); err != nil {
		t.Fatal(err)
	}

	// 30 < 31: status KURANG, transaksi masih terbuka untuk susulan
	st, _ := s.StatusHariIni()
	if st.SemuaSelesai || st.Baris[0].Status != "KURANG" || st.Baris[0].Keterangan != "Kurang 1 Ompreng" || !st.Baris[0].Merah {
		t.Fatalf("status kurang: %+v", st)
	}

	now = now.Add(time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "KEMBALI_SUSULAN" || r.Sisa != 1 {
		t.Fatalf("susulan: %+v", r)
	}
	if err := s.SubmitKembali(id, 1); err != nil {
		t.Fatal(err)
	}

	now = now.Add(time.Minute)
	if r := s.ProsesTap("0012345678"); r.Aksi != "PESAN" {
		t.Fatalf("sudah lengkap: %+v", r)
	}

	st, _ = s.StatusHariIni()
	if !st.SemuaSelesai || st.Baris[0].Keterangan != "Lengkap (susulan)" || st.Baris[0].Merah {
		t.Fatalf("status akhir: %+v", st)
	}
}
```

**`internal/service/susulan_test.go`** (berkas baru):

```go
package service

import (
	"path/filepath"
	"testing"
	"time"

	"mbg/internal/db"
	"mbg/internal/model"
	"mbg/internal/timeutil"
)

const (
	uidA = "1111111111" // Ani, VII A
	uidB = "2222222222" // Budi, VII A
)

// siapkanAmbil: VII A mengambil 10 pukul 10:00; jam di-maju ke 10:30.
func siapkanAmbil(t *testing.T) (*Service, *time.Time, int64) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	clk := time.Date(2026, 1, 12, 10, 0, 0, 0, timeutil.WIB)
	s := NewWithClock(conn, t.TempDir(), func() time.Time { return clk })
	id, err := s.TambahKelas("VII A", "Ketua A", 1)
	harusOK(t, err)
	harusOK(t, s.DaftarkanKartu(uidA, id, "1234567890", "Ani"))
	harusOK(t, s.DaftarkanKartu(uidB, id, "0987654321", "Budi"))
	if r := s.ProsesTap(uidA); r.Aksi != "AMBIL" {
		t.Fatalf("tap ambil: %+v", r)
	}
	harusOK(t, s.SubmitAmbil(id, 10))
	clk = clk.Add(30 * time.Minute)
	return s, &clk, id
}

// siapkanKurang: lanjutan siapkanAmbil; Ani mengembalikan 7 dari 10 (status KURANG) pukul 10:30.
func siapkanKurang(t *testing.T) (*Service, *time.Time, int64) {
	t.Helper()
	s, now, id := siapkanAmbil(t)
	if r := s.ProsesTap(uidA); r.Aksi != "KEMBALI" {
		t.Fatalf("tap kembali: %+v", r)
	}
	harusOK(t, s.SubmitKembali(id, 7))
	return s, now, id
}

func trxPertama(t *testing.T, s *Service) model.Transaksi {
	t.Helper()
	list, err := s.ListTransaksi(model.FilterTransaksi{})
	harusOK(t, err)
	if len(list) != 1 {
		t.Fatalf("mau 1 transaksi, dapat %d", len(list))
	}
	return list[0]
}

func jumlahRincian(t *testing.T, s *Service) (n, sum int) {
	t.Helper()
	harusOK(t, s.db.QueryRow(`SELECT COUNT(*),COALESCE(SUM(jumlah),0) FROM pengembalian`).Scan(&n, &sum))
	return
}

func TestSusulanAlurLengkap(t *testing.T) {
	s, now, id := siapkanKurang(t)

	trx := trxPertama(t, s)
	if trx.Status != "KURANG" || *trx.JumlahKembali != 7 || trx.Keterangan != "Kurang 3 Ompreng" || !trx.Merah {
		t.Fatalf("setelah kembali 7: %+v", trx)
	}

	*now = now.Add(6 * time.Second)
	r := s.ProsesTap(uidB)
	if r.Aksi != "KEMBALI_SUSULAN" || r.Sisa != 3 || r.JumlahAmbil != 10 || r.NamaSiswa != "Budi" {
		t.Fatalf("tap susulan: %+v", r)
	}
	if err := s.SubmitKembali(id, 0); err == nil {
		t.Fatal("susulan 0 harus ditolak")
	}
	if err := s.SubmitKembali(id, 4); err == nil {
		t.Fatal("susulan melebihi sisa harus ditolak")
	}
	harusOK(t, s.SubmitKembali(id, 3))

	trx = trxPertama(t, s)
	if trx.Status != "SELESAI" || *trx.JumlahKembali != 10 || trx.Keterangan != "Lengkap (susulan)" || trx.Merah {
		t.Fatalf("setelah susulan: %+v", trx)
	}
	if trx.NamaKembali != "Ani, Budi" || len(trx.Pengembalian) != 2 {
		t.Fatalf("pengembali ganda: %q rincian=%+v", trx.NamaKembali, trx.Pengembalian)
	}
	if err := s.SubmitKembali(id, 1); err == nil {
		t.Fatal("submit setelah lengkap harus ditolak")
	}

	want := "Kelas VII A sudah mengembalikan ompreng dengan lengkap"
	for _, uid := range []string{uidA, uidB} { // termasuk kartu siswa lain di kelas yang sama
		*now = now.Add(3 * time.Second)
		if r := s.ProsesTap(uid); r.Aksi != "PESAN" || r.Pesan != want {
			t.Fatalf("tap setelah lengkap (%s): %+v", uid, r)
		}
	}

	au, err := s.ListAudit(model.FilterAudit{Entitas: "transaksi"})
	harusOK(t, err)
	ada := false
	for _, a := range au {
		if a.Alasan == "pengembalian susulan" {
			ada = true
		}
	}
	if !ada {
		t.Fatalf("audit susulan tidak ada: %+v", au)
	}
}

func TestKembaliLengkapLangsungLaluTapDitolak(t *testing.T) {
	s, now, id := siapkanAmbil(t)
	if r := s.ProsesTap(uidA); r.Aksi != "KEMBALI" {
		t.Fatalf("tap kembali: %+v", r)
	}
	harusOK(t, s.SubmitKembali(id, 10))
	trx := trxPertama(t, s)
	if trx.Status != "SELESAI" || trx.Keterangan != "Lengkap" || trx.Merah {
		t.Fatalf("lengkap langsung: %+v", trx)
	}
	want := "Kelas VII A sudah mengembalikan ompreng dengan lengkap"
	for _, uid := range []string{uidB, uidA} {
		*now = now.Add(3 * time.Second)
		if r := s.ProsesTap(uid); r.Aksi != "PESAN" || r.Pesan != want {
			t.Fatalf("tap (%s): %+v", uid, r)
		}
	}
}

func TestCooldownSetelahKurang(t *testing.T) {
	s, now, _ := siapkanKurang(t)

	*now = now.Add(3 * time.Second)
	if r := s.ProsesTap(uidB); r.Aksi != "PESAN" || r.Pesan != "Baru saja tercatat" {
		t.Fatalf("3 detik setelah kurang harus tertahan: %+v", r)
	}
	*now = now.Add(3 * time.Second) // total 6 detik > 5 detik
	if r := s.ProsesTap(uidB); r.Aksi != "KEMBALI_SUSULAN" {
		t.Fatalf("6 detik setelah kurang harus lolos: %+v", r)
	}
}

func TestTutupDanBukaHariSaatKurang(t *testing.T) {
	s, now, id := siapkanKurang(t)

	harusOK(t, s.TutupHari("tutup"))
	st, err := s.StatusHariIni()
	harusOK(t, err)
	if b := st.Baris[0]; b.Status != "BELUM_KEMBALI" || b.Keterangan != "Kurang 3 Ompreng" || !b.Merah {
		t.Fatalf("setelah tutup: %+v", b)
	}
	*now = now.Add(6 * time.Second)
	if r := s.ProsesTap(uidB); r.Pesan != "Hari ini sudah ditutup" {
		t.Fatalf("tap saat ditutup: %+v", r)
	}
	if err := s.SubmitKembali(id, 1); err == nil {
		t.Fatal("susulan saat ditutup harus ditolak")
	}

	harusOK(t, s.BukaHari(tglUji, "salah tutup"))
	st, _ = s.StatusHariIni()
	if st.Baris[0].Status != "KURANG" || st.Ditutup {
		t.Fatalf("setelah buka: %+v", st)
	}
	*now = now.Add(10 * time.Second)
	if r := s.ProsesTap(uidB); r.Aksi != "KEMBALI_SUSULAN" || r.Sisa != 3 {
		t.Fatalf("susulan setelah buka: %+v", r)
	}
}

func TestEksporSusulan(t *testing.T) {
	s, now, id := siapkanKurang(t)

	l, err := s.susunLaporan(tglUji, tglUji)
	harusOK(t, err)
	if x := l.Baris[0]; x.Keterangan != "Kurang 3 Ompreng" || !x.Merah || x.Pengembali != "Ani" || x.WaktuKembali != "10:30 WIB" {
		t.Fatalf("sebelum susulan: %+v", x)
	}

	*now = now.Add(30 * time.Minute)
	if r := s.ProsesTap(uidB); r.Aksi != "KEMBALI_SUSULAN" {
		t.Fatalf("tap susulan: %+v", r)
	}
	harusOK(t, s.SubmitKembali(id, 3))

	l, err = s.susunLaporan(tglUji, tglUji)
	harusOK(t, err)
	if x := l.Baris[0]; x.Keterangan != "Lengkap (susulan)" || x.Merah || x.Pengembali != "Ani, Budi" ||
		x.WaktuKembali != "11:00 WIB" || x.JumlahKembali != "10" {
		t.Fatalf("setelah susulan: %+v", x)
	}
}

func TestMigrasiRincianPengembalian(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	harusOK(t, err)
	t.Cleanup(func() { conn.Close() })

	// simulasi DB lama: tidak ada tabel pengembalian, ada transaksi SELESAI dan AMBIL
	_, err = conn.Exec(`DROP TABLE pengembalian`)
	harusOK(t, err)
	_, err = conn.Exec(`INSERT INTO kelas (id,nama,nama_ketua) VALUES (1,'VII A','Ani'),(2,'VII B','Budi')`)
	harusOK(t, err)
	_, err = conn.Exec(`INSERT INTO transaksi (id,tanggal,kelas_id,nama_kelas,nama_ketua,waktu_ambil,waktu_kembali,jumlah_ambil,jumlah_kembali,status,nama_kembali)
		VALUES (1,'2026-01-10',1,'VII A','Ani','09:00','10:00',10,8,'SELESAI','Budi')`)
	harusOK(t, err)
	_, err = conn.Exec(`INSERT INTO transaksi (id,tanggal,kelas_id,nama_kelas,nama_ketua,waktu_ambil,jumlah_ambil,status)
		VALUES (2,'2026-01-10',2,'VII B','Budi','09:00',5,'AMBIL')`)
	harusOK(t, err)

	for i := 0; i < 2; i++ { // dua kali: harus idempoten
		harusOK(t, db.Migrate(conn))
		var n, jumlah int
		var nama string
		harusOK(t, conn.QueryRow(`SELECT COUNT(*),COALESCE(MAX(jumlah),0),COALESCE(MAX(nama_siswa),'') FROM pengembalian`).Scan(&n, &jumlah, &nama))
		if n != 1 || jumlah != 8 || nama != "Budi" {
			t.Fatalf("migrasi ke-%d: n=%d jumlah=%d nama=%q", i+1, n, jumlah, nama)
		}
	}
}

func TestEditTransaksiMenyinkronkanRincian(t *testing.T) {
	s, now, _ := siapkanKurang(t)
	*now = now.Add(6 * time.Second)
	if r := s.ProsesTap(uidB); r.Aksi != "KEMBALI_SUSULAN" {
		t.Fatalf("tap susulan: %+v", r)
	}
	harusOK(t, s.SubmitKembali(s.idKelasUji(t), 3))
	if n, sum := jumlahRincian(t, s); n != 2 || sum != 10 {
		t.Fatalf("awal: n=%d sum=%d", n, sum)
	}
	trx := trxPertama(t, s)
	ip := func(n int) *int { return &n }
	edit := func(status string, jk *int) model.EditTransaksi {
		return model.EditTransaksi{ID: trx.ID, Status: status, JumlahAmbil: ip(10), JumlahKembali: jk,
			WaktuAmbil: "10:00", WaktuKembali: "11:00", Alasan: "koreksi"}
	}

	harusOK(t, s.UbahTransaksi(edit("SELESAI", ip(10)))) // total sama: rincian dipertahankan
	if n, _ := jumlahRincian(t, s); n != 2 {
		t.Fatalf("total sama harus mempertahankan rincian, n=%d", n)
	}
	harusOK(t, s.UbahTransaksi(edit("SELESAI", ip(9)))) // total berbeda: diganti satu baris koreksi
	if n, sum := jumlahRincian(t, s); n != 1 || sum != 9 {
		t.Fatalf("setelah koreksi: n=%d sum=%d", n, sum)
	}
	if err := s.UbahTransaksi(edit("KURANG", ip(10))); err == nil {
		t.Fatal("KURANG dengan kembali == ambil harus ditolak")
	}
	harusOK(t, s.UbahTransaksi(edit("KURANG", ip(9))))
	if x := trxPertama(t, s); x.Status != "KURANG" || x.Keterangan != "Kurang 1 Ompreng" {
		t.Fatalf("edit ke KURANG: %+v", x)
	}
	harusOK(t, s.UbahTransaksi(model.EditTransaksi{ID: trx.ID, Status: "AMBIL", JumlahAmbil: ip(10), WaktuAmbil: "10:00", Alasan: "batal kembali"}))
	if n, _ := jumlahRincian(t, s); n != 0 {
		t.Fatalf("edit ke AMBIL harus menghapus rincian, n=%d", n)
	}
}

func TestHapusTransaksiMenghapusRincian(t *testing.T) {
	s, _, _ := siapkanKurang(t)
	trx := trxPertama(t, s)
	if n, _ := jumlahRincian(t, s); n != 1 {
		t.Fatalf("rincian awal n=%d", n)
	}
	harusOK(t, s.HapusTransaksi(trx.ID, "dobel"))
	if n, _ := jumlahRincian(t, s); n != 0 {
		t.Fatalf("rincian harus ikut terhapus, n=%d", n)
	}
}

// idKelasUji: ID kelas satu-satunya pada DB uji.
func (s *Service) idKelasUji(t *testing.T) int64 {
	t.Helper()
	var id int64
	harusOK(t, s.db.QueryRow(`SELECT id FROM kelas LIMIT 1`).Scan(&id))
	return id
}
```

Catatan tes: kartu didaftarkan dengan NISN 10 digit (`1234567890`, `0987654321`). Bila `DaftarkanKartu` punya aturan NISN lain, ubah nilainya di `siapkanAmbil`.

## 9. Jalankan dan commit

```bash
gofmt -l . ; go vet ./... && go test ./...
wails build        # regenerasi frontend/wailsjs/go/models.ts
git add -A
git commit -m "Pengembalian susulan: status KURANG, tabel pengembalian, popup susulan"
```

Jika `go vet` mengeluh ada pemanggil `Keterangan(...)` lain, misalnya di `master.go`, tambahkan argumen kelima `false`. Saya belum melihat berkas itu.

## 10. Ringkasan dan versi

- **Aturan baru:** pengembalian kurang membuat status `KURANG` dan tap berikutnya membuka popup susulan dengan sisa, tanpa jeda minimum. Pengembalian lengkap membuat `SELESAI` final, dan tap berikutnya ditolak dengan pesan jelas untuk kartu mana pun di kelas itu.
- **Skema:** tabel `pengembalian` dengan backfill idempoten dan `ON DELETE CASCADE`. Kolom di `transaksi` tetap sebagai ringkasan (total, waktu terakhir, nama dipisah koma), jadi laporan dan ekspor hampir tidak berubah.
- **Hari:** tutup hari mengubah `KURANG` menjadi `BELUM_KEMBALI` dengan keterangan sisa, dan buka hari mengembalikannya ke `KURANG`.
- **Cooldown:** 5 detik setelah `KURANG`, dan 30 detik tetap untuk kasus lain.
- **Riwayat:** edit menyinkronkan rincian, dan audit menyimpan rincian lama.

Saya sarankan versi **0.2.0**. Pengubahan nomor versi saya tunda ke Fitur 2, karena di sana versi dipindah ke `-ldflags`, jadi tidak perlu di-`sed` dua kali.

Kalau tes lulus, balas dengan hasilnya dan saya lanjut ke **Fitur 1 (Import/Export master)**, mulai dari rencananya.