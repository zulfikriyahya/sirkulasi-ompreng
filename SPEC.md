# Spesifikasi: Aplikasi Pencatatan Pengambilan & Pengembalian Ompreng MBG

Versi 1.1 (semua hal terbuka sudah diputuskan; siap scaffolding)

## 1. Keputusan yang sudah disepakati

| Topik | Keputusan |
|---|---|
| Backend | Go |
| Database | SQLite (`modernc.org/sqlite`, tanpa CGO untuk database) |
| UI | Wails v2 (webview), frontend vanilla HTML/CSS/JS |
| OS target | Windows 10 ke atas; Linux Debian dan Red Hat family |
| Reader RFID | Serial (ESP32-C3 Super Mini) dan HID keyboard-emulator, deteksi otomatis |
| Format UID | 10 digit desimal, sama untuk kedua jenis reader |
| Komputer | Satu komputer per instalasi; bisa dipasang di komputer lain lewat installer |
| Login | Tidak ada |
| Aturan kelas | Semua kelas wajib mengambil |
| Koreksi data | Edit dan hapus boleh, dengan konfirmasi dan audit log |
| Jumlah kembali > ambil | Ditolak dengan peringatan |
| Penutupan hari | Ada "Tutup hari dengan catatan" |
| Ekspor | Excel (.xlsx) dan PDF, A4 landscape, zona waktu WIB tetap |
| Input data awal | Manual (tanpa impor Excel) |

## 2. Arsitektur

```
+----------------------- Aplikasi (satu proses) ------------------------+
|  Frontend (webview)          Backend Go                              |
|  HTML/CSS/JS vanilla  <-->   Wails binding + event                   |
|                                |- service: transaksi, master, ekspor |
|                                |- reader manager (serial + HID)      |
|                                |- SQLite (file tunggal)              |
+----------------------------------------------------------------------+
```

- Frontend memanggil fungsi Go lewat binding Wails dan menerima event dari backend (misalnya `card:tap`).
- Reader serial dibaca oleh backend. Reader HID ditangkap oleh frontend (lihat bagian 5).
- Lokasi data: `%APPDATA%\MBG\mbg.db` (Windows), `~/.local/share/mbg/mbg.db` (Linux). Terpisah dari program, jadi update tidak menghapus data.

### Struktur proyek

```
mbg/
  main.go                 // entry Wails
  app.go                  // method yang diekspos ke frontend
  internal/
    db/                   // koneksi, migrasi
    model/                // struct domain
    service/              // transaksi, master, tutup hari, audit
    reader/               // serial scanner, normalisasi UID
    export/               // excel.go, pdf.go
    timeutil/             // WIB, nama bulan/hari Indonesia
  frontend/
    index.html
    css/
    js/                   // modul: dashboard, master, riwayat, ekspor, tap
  build/                  // ikon, installer (NSIS, nfpm)
  firmware/esp32c3_rc522/ // sketch reader ESP32-C3 + RC522 (lihat bagian 5.2 dan 12)
```

## 3. Skema database

```sql
CREATE TABLE pengaturan (          -- satu baris (id = 1)
  id INTEGER PRIMARY KEY CHECK (id = 1),
  nama_instansi TEXT, alamat_instansi TEXT,
  logo BLOB, logo_mime TEXT,
  penanggung_jawab TEXT, nip_penanggung_jawab TEXT,
  jabatan_penanggung_jawab TEXT
);

CREATE TABLE kelas (
  id INTEGER PRIMARY KEY,
  nama TEXT NOT NULL UNIQUE,       -- "VII A"
  nama_ketua TEXT NOT NULL,
  urutan INTEGER NOT NULL DEFAULT 0,
  aktif INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE kartu (
  uid TEXT PRIMARY KEY,            -- 10 digit, sudah dinormalisasi
  kelas_id INTEGER NOT NULL REFERENCES kelas(id),
  aktif INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE transaksi (
  id INTEGER PRIMARY KEY,
  tanggal TEXT NOT NULL,           -- 'YYYY-MM-DD' (WIB)
  kelas_id INTEGER NOT NULL REFERENCES kelas(id),
  nama_kelas TEXT NOT NULL,        -- snapshot saat transaksi
  nama_ketua TEXT NOT NULL,        -- snapshot saat transaksi
  waktu_ambil TEXT,                -- 'HH:MM' WIB
  waktu_kembali TEXT,
  jumlah_ambil INTEGER,
  jumlah_kembali INTEGER,
  status TEXT NOT NULL,            -- AMBIL | SELESAI | BELUM_KEMBALI | TIDAK_AMBIL
  catatan TEXT,                    -- catatan manual opsional
  UNIQUE (tanggal, kelas_id)
);

CREATE TABLE hari_ditutup (
  tanggal TEXT PRIMARY KEY,
  waktu_tutup TEXT NOT NULL,
  catatan TEXT
);

CREATE TABLE audit_log (
  id INTEGER PRIMARY KEY,
  waktu TEXT NOT NULL,             -- ISO 8601, WIB
  aksi TEXT NOT NULL,              -- TAMBAH | UBAH | HAPUS | TUTUP_HARI | BUKA_HARI | RESTORE
  entitas TEXT NOT NULL,           -- transaksi | kelas | kartu | pengaturan | hari
  entitas_id TEXT,
  nilai_lama TEXT,                 -- JSON
  nilai_baru TEXT,                 -- JSON
  alasan TEXT
);
```

Catatan desain:
- Nama kelas dan ketua disimpan sebagai snapshot di `transaksi`, supaya laporan lama tidak berubah saat ketua diganti.
- `UNIQUE (tanggal, kelas_id)` menjamin satu pengambilan per kelas per hari. Kalau nanti ada kelas yang mengambil dua kali sehari, tabel ini perlu dipecah (lihat bagian 11).
- Waktu disimpan sebagai teks WIB tetap. Jam sistem komputer diasumsikan benar, tetapi aplikasi selalu menghitung WIB sendiri dari UTC (`time.FixedZone("WIB", 7*3600)`), bukan dari zona waktu OS.
- Aktifkan `PRAGMA journal_mode=WAL` dan `foreign_keys=ON`.

## 4. Aturan bisnis

### 4.1 Logika tap kartu

```
tap(uid):
  uid <- normalisasi(uid)
  abaikan jika ada popup ambil/kembali yang sedang terbuka     (penguncian popup)
  abaikan jika uid sama dengan tap sebelumnya dalam < 2 detik  (debounce)
  kartu <- cari(uid); jika tidak ada/nonaktif -> pesan "Kartu tidak terdaftar"
  trx   <- transaksi hari ini untuk kartu.kelas_id
  jika hari ini sudah ditutup          -> tolak ("Hari ini sudah ditutup")
  jika kartu ini baru submit < 30 detik lalu -> abaikan, pesan "Baru saja tercatat"  (cooldown)
  jika trx tidak ada                   -> popup AMBIL (input jumlah_ambil)
  jika trx.status == AMBIL:
      jika selisih sekarang - waktu_ambil < jeda_minimum (default 10 menit)
          -> dialog konfirmasi "Kelas X baru mengambil N menit lalu. Tetap proses pengembalian?"
             (Ya -> popup KEMBALI, Tidak -> batal)
      selain itu                       -> popup KEMBALI (input jumlah_kembali)
  jika trx.status SELESAI              -> pesan "Kelas ini sudah selesai hari ini"
```

Aturan perlindungan tap ganda:
- **Penguncian popup**: selama popup ambil/kembali terbuka, semua tap (kelas mana pun) diabaikan dengan pesan "Selesaikan input kelas X dulu".
- **Cooldown 30 detik** setelah submit untuk kartu yang sama.
- **Jeda minimum ambil ke kembali** (default 10 menit, bisa diatur di Pengaturan): di bawah jeda itu muncul dialog konfirmasi, bukan penolakan mutlak.

### 4.2 Validasi
- `jumlah_ambil` bilangan bulat >= 1.
- `jumlah_kembali` bilangan bulat >= 0. **Jika lebih besar dari `jumlah_ambil` maka ditolak** dengan peringatan jelas, popup tetap terbuka.
- Waktu diambil otomatis dari jam komputer saat submit (bukan saat tap).

### 4.3 Keterangan otomatis (dihitung saat tampil dan ekspor)

| Kondisi | Keterangan | Baris merah |
|---|---|---|
| status SELESAI, kembali = ambil | `Lengkap` | tidak |
| status SELESAI, kembali < ambil | `Kurang N Ompreng` (N = selisih) | ya |
| status BELUM_KEMBALI | `Belum kembali` | ya |
| status TIDAK_AMBIL | (tidak ditampilkan; **baris tidak muncul di laporan** PDF/Excel) | - |
| catatan manual terisi | ditambahkan di belakang keterangan | mengikuti aturan di atas |

### 4.4 Tombol cetak dan Tutup hari
- **Normal**: tombol "Cetak PDF" muncul ketika semua kelas aktif berstatus SELESAI pada hari itu.
- **Tutup hari dengan catatan**: tersedia kapan saja selama ada kelas yang belum selesai. Aksi ini meminta konfirmasi dan catatan, lalu:
  - kelas berstatus AMBIL menjadi BELUM_KEMBALI,
  - kelas tanpa transaksi dibuatkan baris TIDAK_AMBIL (tetap tampil di dashboard dan riwayat sebagai "Tidak mengambil", tetapi **dikecualikan dari ekspor laporan**),
  - hari dicatat di `hari_ditutup` dan tombol cetak aktif.
- Hari yang sudah ditutup bisa dibuka kembali (dengan konfirmasi dan tercatat di audit log) untuk koreksi.
- Kelas yang aktif didefinisikan per hari ini. Kelas yang dinonaktifkan tidak dihitung pada hari-hari berikutnya, tetapi transaksi lamanya tetap utuh.

### 4.5 Edit dan hapus transaksi
- Edit/hapus lewat layar Riwayat, selalu dengan dialog konfirmasi dan kolom alasan.
- Setiap perubahan menulis satu baris `audit_log` berisi nilai lama dan baru.
- Validasi jumlah (4.2) tetap berlaku saat edit.

## 5. Reader RFID

Semua sumber digabung menjadi satu alur: `uid 10 digit -> tap(uid)`.

### 5.1 Normalisasi UID
Buang semua karakter non-digit (spasi, `\r`, `\n`), lalu isi nol di kiri sampai 10 digit. Hasil yang bukan 1-10 digit dianggap input tidak valid dan diabaikan.

### 5.2 Reader serial (ESP32-C3)
- ESP32-C3 Super Mini memakai USB native (CDC), sehingga muncul sebagai port serial (Windows: `COMx`, Linux: `/dev/ttyACM*`). VID Espressif adalah `303A`.
- Backend memindai port tiap 2 detik (hotplug) dengan `go.bug.st/serial` dan `enumerator`. Port dipilih berdasarkan VID/PID, lalu dibuka dan dibaca per baris.
- Kontrak firmware: kirim satu baris per tap berupa 10 digit desimal diikuti `\n`, baud 115200. Contoh: `0012345678\n`.
- Rekomendasi tambahan untuk firmware: kirim juga `READY\n` saat boot sebagai tanda dikenali. Baris yang bukan 10 digit hanya dicatat di log debug.
- Jika port terlepas, backend menutupnya dan terus memindai. Status ("Reader serial terhubung/terputus") ditampilkan di pojok layar.

### 5.3 Reader HID (keyboard emulator)
- Reader HID tidak bisa dibedakan dari keyboard lewat deteksi perangkat. Deteksinya dilakukan di frontend lewat **pola ketikan**: listener `keydown` global mengumpulkan karakter; jika terkumpul 8-10 digit dengan jeda antar karakter < 50 ms lalu `Enter`, buffer dianggap tap kartu.
- Ketikan manusia (jeda lebih lama) dibuang dan tidak mengganggu input lain. Saat tap valid terdeteksi, event default pada field yang sedang fokus dibatalkan.
- Ambang jeda dibuat bisa diatur di Pengaturan agar bisa disesuaikan dengan reader yang dipakai.
- Kalau reader HID mengeluarkan format selain 10 digit desimal (misalnya hex), tambahkan konversi di normalisasi. Cek dulu dengan reader yang akan dibeli.

### 5.4 Cadangan manual
Tombol "Input UID manual" untuk kartu rusak atau hilang, dan di layar Master tersedia pendaftaran kartu baru lewat tap atau ketik.

## 6. Fungsi backend (binding Wails)

| Grup | Fungsi |
|---|---|
| Pengaturan | `GetPengaturan`, `SimpanPengaturan`, `UploadLogo` |
| Kelas | `ListKelas`, `TambahKelas`, `UbahKelas`, `NonaktifkanKelas` |
| Kartu | `DaftarkanKartu(uid, kelasID)`, `HapusKartu`, `ListKartu(kelasID)` |
| Tap | `ProsesTap(uid)` mengembalikan aksi yang dibutuhkan UI; `SubmitAmbil`, `SubmitKembali` |
| Dashboard | `StatusHariIni()` berisi per kelas: status, jumlah, waktu |
| Hari | `TutupHari(catatan)`, `BukaHari(tanggal, alasan)`, `ApakahBolehCetak(tanggal)` |
| Riwayat | `ListTransaksi(filter)`, `UbahTransaksi`, `HapusTransaksi`, `ListAudit(filter)` |
| Ekspor | `EksporExcel(tanggalDari, tanggalSampai)`, `EksporPDF(...)` (dialog simpan file) |
| Backup | `Backup(path)`, `Restore(path)` (backup memakai `VACUUM INTO`) |
| Reader | event `reader:status`, event `card:tap` |

## 7. Layar UI

1. **Beranda (Dashboard)**: daftar semua kelas dengan status (Belum ambil / Sudah ambil / Selesai), ringkasan jumlah, indikator reader, tombol "Tutup hari", dan tombol "Cetak PDF" saat syarat terpenuhi. Popup ambil/kembali muncul di atas layar ini saat tap.
2. **Master Data**: tab Instansi (nama, alamat, logo, penanggung jawab, NIP), tab Kelas & Ketua, tab Kartu RFID.
3. **Riwayat**: filter tanggal, kelas, status; edit/hapus dengan konfirmasi; tab Audit Log.
4. **Ekspor**: pilih rentang tanggal, pratinjau, tombol Excel dan PDF.
5. **Pengaturan**: ambang kecepatan HID, port serial manual (opsional), backup/restore, lokasi data.

Umpan balik tap: warna hijau untuk sukses, merah untuk ditolak, dengan bunyi pendek (`AudioContext`, tanpa file audio eksternal).

## 8. Ekspor

### Kolom (PDF dan Excel)
`No. | Kelas | Nama Ketua Murid | Tanggal | Waktu Pengambilan | Waktu Pengembalian | Jumlah Ambil | Jumlah Kembali | Keterangan | Paraf`

- **Kop surat**: logo, nama instansi, alamat, garis pemisah. Dokumen berbentuk **laporan formal**: di bawah kop ada judul tengah "LAPORAN PENGAMBILAN DAN PENGEMBALIAN OMPRENG MAKANAN BERGIZI GRATIS (MBG)" dan baris periode ("Hari/Tanggal: Senin, 12 Januari 2026" atau "Periode: 12 - 16 Januari 2026").
- Kelas berstatus `TIDAK_AMBIL` tidak dimasukkan ke tabel laporan.
- **Baris merah**: baris dengan kekurangan, belum kembali, atau tidak mengambil diberi latar merah muda dengan teks merah tua; baris lengkap polos. Berlaku sama di PDF dan Excel.
- **Paraf**: dibiarkan kosong dengan tinggi baris cukup untuk tanda tangan.
- **Blok tanda tangan** (kanan bawah, formal): kota dan tanggal cetak ("..., 12 Januari 2026"), jabatan ("Penanggung Jawab MBG"), ruang tanda tangan, nama (tebal, digarisbawahi), dan "NIP. ...". Jabatan bisa diubah di Master Data.
- **PDF**: A4 landscape, header tabel diulang di tiap halaman, nomor halaman, tanggal dalam format Indonesia ("12 Januari 2026"), waktu "11:46 WIB". Nama bulan dan hari ditulis oleh paket `timeutil`, bukan dari locale OS.
- **Excel**: logo dan kop di bagian atas, lebar kolom diatur, header dibekukan, pengaturan halaman A4 landscape agar bisa dicetak dari Excel.
- Pustaka: `excelize` (Excel) dan `gofpdf` atau fork yang dipelihara (PDF). Font tertanam (embed), tanpa dependensi font sistem.
- Rentang tanggal: baris diurutkan per tanggal lalu per urutan kelas; nomor urut berlanjut.

## 9. Build dan distribusi

| OS | Hasil | Catatan |
|---|---|---|
| Windows 10+ | Installer `.exe` (NSIS) | Wails menyertakan bootstrapper WebView2; ikon, shortcut, uninstaller |
| Debian/Ubuntu | `.deb` | Dependensi: `libwebkit2gtk-4.0-37` atau `-4.1-37` (sesuai versi distro) |
| Red Hat family | `.rpm` | Dependensi: `webkit2gtk3` (nama paket tergantung versi) |

- Karena Wails memakai CGO di Linux, build dilakukan di runner masing-masing OS lewat CI (GitHub Actions: `ubuntu`, `windows`), bukan cross-compile dari satu mesin.
- Paket `.deb`/`.rpm` dibuat dengan `nfpm`.
- Linux: tambahkan aturan udev atau instruksi agar user masuk grup `dialout` (akses `/dev/ttyACM*`).
- Windows: driver USB CDC bawaan Windows 10 sudah cukup untuk ESP32-C3.

## 10. Pengujian

- **Unit test**: normalisasi UID, logika tap, validasi jumlah, perhitungan keterangan, syarat cetak, tutup hari.
- **Simulator reader**: mode pengembangan yang menyuntikkan tap palsu (HID dan serial virtual), supaya alur bisa diuji tanpa perangkat.
- **Uji manual di lapangan**: reader ESP32-C3 asli dan satu reader HID pasaran, di Windows 10/11 dan satu distro Debian dan Red Hat.
- **Uji ekspor**: cetak PDF sungguhan dan buka Excel di Microsoft Excel dan LibreOffice.

## 11. Keputusan akhir hal terbuka

1. Kelas yang tidak mengambil saat "Tutup hari": **tidak muncul di laporan** (status internal `TIDAK_AMBIL` tetap disimpan untuk dashboard dan audit).
2. **Hanya satu kali** pengambilan dan pengembalian per kelas per hari; `UNIQUE (tanggal, kelas_id)` dipertahankan.
3. Reader HID: **10 digit desimal** (default pasaran). Tetap diverifikasi dengan reader asli saat uji lapangan.
4. Hari libur: **cukup tidak membuka aplikasi**; tidak ada fitur tandai libur. Laporan rentang tanggal hanya memuat hari yang punya data.
5. Dokumen berbentuk **laporan formal** (judul, periode, kop, blok tanda tangan; lihat bagian 8).
6. Firmware ESP32-C3 + **RC522** sudah ditulis: `firmware/esp32c3_rc522/esp32c3_rc522.ino` (bagian 12).

## 12. Firmware ESP32-C3 + RC522

### Wiring (ESP32-C3 Super Mini ke RC522)

| RC522 | ESP32-C3 | Catatan |
|---|---|---|
| SDA (SS) | GPIO 7 | |
| SCK | GPIO 4 | |
| MOSI | GPIO 6 | |
| MISO | GPIO 5 | |
| RST | GPIO 3 | |
| 3.3V | 3V3 | **jangan 5V**, RC522 bekerja di 3.3V |
| GND | GND | |
| IRQ | tidak dipakai | |

### Cara upload
1. Arduino IDE: pasang board "esp32 by Espressif", pilih **ESP32C3 Dev Module**, aktifkan **USB CDC On Boot: Enabled**.
2. Pasang library **MFRC522** lewat Library Manager.
3. Buka `esp32c3_rc522.ino` dan upload. Jika board tidak terdeteksi: tahan tombol BOOT, tekan RESET, lepas BOOT.

### Perilaku
- Tiap tap mengirim `0012345678\n` (10 digit desimal) dengan baud 115200; kartu sama dalam 1,5 detik diabaikan (debounce).
- Mengirim `READY` saat boot, dan `ERR:RC522` jika modul tidak terbaca (otomatis mencoba pulih tiap 5 detik).
- Angka UID = 4 byte UID kartu sebagai bilangan 32-bit (urutan little-endian secara default, seperti kebanyakan reader HID 13.56 MHz). **Kalibrasi wajib**: tap satu kartu di reader HID dan di alat ini; jika angkanya berbeda, ubah `UID_LITTLE_ENDIAN` ke `false`. Hal ini penting jika kartu didaftarkan lewat satu reader tetapi dipakai di reader lain.
- Kartu MIFARE dengan UID 7 atau 10 byte memakai 4 byte terakhir; jika kartu sekolah memakai UID panjang, hal ini perlu diuji.

## 13. Tahapan pengerjaan

1. **Fondasi**: inisialisasi Wails, skema dan migrasi SQLite, paket `timeutil`, `audit`.
2. **Master data**: layar Instansi, Kelas, Kartu.
3. **Inti transaksi**: logika tap, popup ambil/kembali, dashboard, simulator reader.
4. **Reader nyata**: serial otomatis, deteksi HID, status reader.
5. **Riwayat dan koreksi**: edit/hapus, audit log, tutup/buka hari.
6. **Ekspor**: Excel, lalu PDF, lalu uji cetak.
7. **Backup/restore dan Pengaturan**.
8. **Packaging**: CI, installer Windows, `.deb`, `.rpm`, uji instalasi di mesin bersih.

---

## Struktur Project
```bash
.
├── draft.md
├── firmware
│   └── esp32c3_rc522
│       └── esp32c3_rc522.ino
├── frontend
│   ├── css
│   │   └── style.css
│   ├── index.html
│   └── js
│       ├── api.js
│       ├── dashboard.js
│       ├── ekspor.js
│       ├── main.js
│       ├── master.js
│       ├── riwayat.js
│       └── tap.js
├── go.mod
├── go.sum
├── internal
│   ├── db
│   │   └── db.go
│   ├── export
│   │   ├── excel.go
│   │   ├── export.go
│   │   ├── export_test.go
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
├── main.go
├── SPEC.md
└── wails.json
```