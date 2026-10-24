## Struktur Project

```bash
sirkulasi-ompreng on main ≡  ?2 ~3
➜ tree
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
│   ├── checksums.txt
│   ├── mbg-0.2.1-1.x86_64.rpm
│   ├── mbg_0.2.1_amd64.deb
│   ├── mbg-amd64-installer.exe
│   └── mbg-linux-amd64
├── firmware
│   └── esp32c3_rc522
│       └── esp32c3_rc522.ino
├── frontend
│   ├── css
│   │   └── style.css
│   ├── images
│   │   └── icon.ico
│   ├── index.html
│   ├── js
│   │   ├── api.js
│   │   ├── dashboard.js
│   │   ├── ekspor.js
│   │   ├── main.js
│   │   ├── master.js
│   │   ├── pembaruan.js
│   │   ├── riwayat.js
│   │   ├── sinkron.js
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
├── icon.png
├── internal
│   ├── db
│   │   └── db.go
│   ├── export
│   │   ├── excel.go
│   │   ├── export.go
│   │   └── pdf.go
│   ├── model
│   │   └── model.go
│   ├── pembaruan
│   │   ├── pembaruan.go
│   │   └── pembaruan_test.go
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
│   │   ├── pembaruan.go
│   │   ├── pembaruan_test.go
│   │   ├── reader.go
│   │   ├── riwayat.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── sinkron.go
│   │   ├── sinkron_test.go
│   │   ├── susulan_test.go
│   │   ├── tahap5_test.go
│   │   └── transaksi.go
│   ├── sinkron
│   │   └── sinkron.go
│   ├── timeutil
│   │   └── timeutil.go
│   └── versi
│       └── versi.go
├── Issue.md
├── LICENSE
├── main.go
├── nfpm.yaml
├── README.md
└── wails.json

28 directories, 75 files
sirkulasi-ompreng on main ≡  ?2 ~3
➜
```

## Issues
- gunakan icon pada frontend/images/icon.ico aplikasi ketimbang icon "M"
- gunakan sidebar ketimbang top navbar
- sidebar dapat dicollapse
- tambahkan data wali kelas
- tambahkan fitur notifikasi whatsapp yang dikirimkan ke wali kelas
- layanan whatsapp (https://whatsapp.mtsn1pandeglang.sch.id/api/v1/messages)

```txt
Panduan Integrasi API untuk Aplikasi Eksternal

Aplikasi eksternal milik instansi dapat mengirim notifikasi WhatsApp melalui endpoint POST /api/v1/messages menggunakan autentikasi berbasis HMAC-SHA256 - bukan Bearer token. Setiap request wajib menyertakan tiga header: X-API-Key (key_id API key Anda), X-Signature (HMAC-SHA256 dari body request mentah, memakai secret API key sebagai kunci), dan X-Timestamp (Unix timestamp saat request dibuat, dengan toleransi selisih waktu ±5 menit terhadap waktu server untuk mencegah serangan replay). Endpoint ini mendukung pengiriman pesan teks, dokumen, gambar, video, link, dan kontak vCard.

    Pastikan Anda sudah memiliki API key aktif (key_id + secret) - lihat Panduan API Key jika belum.
    Buat atau pilih template pesan yang sudah dikaitkan ke API key ini dari halaman Template, lengkap dengan daftar variabel wajib yang harus disertakan saat pengiriman.
    Normalisasi nomor tujuan (recipient) ke format internasional Indonesia (mis. 6281234567890, boleh dengan atau tanpa awalan 0/62/8 - sistem akan menormalisasi otomatis) sebelum dikirim.
    Susun body JSON berisi template_code, recipient, variables (wajib memuat seluruh variabel_wajib milik template), dan opsional media (untuk dokumen/gambar/video/link/kontak) serta reference_id (untuk idempotency).
    Hitung X-Signature sebagai HMAC-SHA256 dari body request mentah (byte persis yang dikirim) menggunakan secret API key Anda sebagai kunci, lalu encode hasilnya sebagai hex string.
    Sertakan header X-API-Key, X-Signature, dan X-Timestamp pada setiap request - request tanpa HMAC valid atau timestamp kedaluwarsa akan ditolak dengan status 401 Unauthorized.
    Gunakan reference_id yang unik per transaksi/aksi bisnis di sisi Anda untuk memanfaatkan idempotency (window 24 jam) - request dengan reference_id sama dan payload identik tidak akan membuat pengiriman ganda.
    Tangani response 429 (Too Many Requests) sebagai tanda guard rail API key sedang auto-pause akibat lonjakan volume - hubungi admin instansi untuk mengaktifkan kembali setelah penyebab diperiksa.
    Pantau status pengiriman secara real-time via GET /api/v1/messages/{job_id} (status: queued, processing, sent, delivered, read, atau failed), atau lihat riwayatnya di halaman Log Pesan/Antrian Job pada dashboard.

Panduan Pembuatan Template

Template pesan memungkinkan Anda menyusun format pesan WhatsApp sekali saja dan memakainya berulang dengan variabel berbeda-beda di setiap pengiriman - cocok untuk notifikasi berulang seperti presensi, OTP, tagihan, atau pengumuman. Setiap template memiliki kode unik per instansi, kategori bebas (mis. 'Presensi', 'OTP', 'Tagihan'), prioritas pengiriman (High/Medium/Low), 2 hingga 10 varian teks yang dipilih acak setiap kali dikirim (mengurangi pola pengiriman seragam yang berisiko terdeteksi sebagai spam), serta daftar variabel wajib yang harus disertakan aplikasi pengirim.

    Buka halaman Template pada dashboard instansi Anda (menu sidebar 'Template').
    Klik 'Buat Template Baru', isi kode unik (huruf kecil, tanpa spasi, mis. presensi_masuk), kategori bebas sesuai kebutuhan Anda, dan pilih prioritas: High untuk pesan mendesak seperti OTP/reset password, Medium untuk notifikasi transaksional rutin seperti presensi/tagihan, atau Low untuk informasi umum/pengumuman yang tidak mendesak.
    Tulis 2 hingga 10 varian teks dengan format placeholder {{nama_variabel}} - semakin banyak varian yang wajar dan berbeda gaya bahasanya, semakin baik untuk menyamarkan pola pengiriman otomatis.
    Perhatikan format yang dibatasi validator: hindari huruf kapital berlebihan dalam satu kalimat, tidak lebih dari satu tanda seru berurutan (!!), dan maksimal satu tautan (link) per varian - template yang melanggar akan ditolak sistem karena berisiko ditandai spam oleh WhatsApp.
    Isi daftar variabel wajib (dipisah koma, mis. nama_siswa,jam) - setiap variabel yang disebut di dalam teks tapi tidak terdaftar sebagai wajib tetap boleh dikirim kosong tanpa validasi ketat, jadi pastikan variabel penting selalu didaftarkan sebagai wajib.
    Kaitkan template ke satu API key tertentu dari daftar template jika ingin membatasi penggunaannya hanya untuk aplikasi tertentu (satu template hanya bisa terhubung ke satu API key dalam satu waktu; putuskan kaitan dari daftar yang sama jika ingin mengaitkan ulang ke API key lain).
    Simpan template, lalu gunakan kodenya sebagai nilai field template_code saat mengirim pesan melalui endpoint API eksternal.
    Uji coba pengiriman ke nomor WhatsApp Anda sendiri terlebih dahulu untuk memverifikasi seluruh variabel tampil benar sebelum dipakai secara luas ke penerima sesungguhnya.
    Anda dapat mengubah kategori, prioritas, varian, dan variabel wajib kapan saja tanpa mengubah kode template (sehingga integrasi API yang sudah berjalan tidak perlu diubah).

```
```bash
curl -X POST https://whatsapp.mtsn1pandeglang.id/api/v1/messages \
  -H "X-API-Key: key_xxx" \
  -H "X-Signature: <hmac_sha256_dari_body>" \
  -H "X-Timestamp: 1751875200" \
  -H "Content-Type: application/json" \
  -d '{
    "template_code": "presensi_masuk",
    "recipient": "6281234567890",
    "variables": { "nama": "Yahya Zulfikri", "jam": "07:15" },
    "media": null,
    "reference_id": "APP-INTERNAL-12345"
  }'

# Contoh cek status pengiriman:
curl -X GET https://whatsapp.mtsn1pandeglang.id/api/v1/messages/job_xxx \
  -H "X-API-Key: key_xxx" \
  -H "X-Signature: <hmac_sha256_dari_body_kosong>" \
  -H "X-Timestamp: 1751875200"

# Contoh isi salah satu varian template (kategori Presensi, prioritas Medium):
"Halo {{nama}}, kehadiran Anda pada {{tanggal}} pukul {{jam}} telah berhasil dicatat. Terima kasih."

Varian kedua untuk template yang sama (dipilih acak bergantian):
"Presensi tercatat: {{nama}} hadir pada {{tanggal}} pukul {{jam}}."

Dikirim aplikasi eksternal via field "variables":
{ "nama": "Yahya Zulfikri", "tanggal": "7 Juli 2026", "jam": "08:00" }

Contoh lain untuk kategori OTP (prioritas High, variabel wajib: kode_otp,durasi_menit):
"Kode OTP Anda: {{kode_otp}}. Berlaku {{durasi_menit}} menit, jangan bagikan ke siapa pun."
```
