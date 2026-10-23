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
Setiap fitur: rencana -> persetujuan -> implementasi -> tes -> ringkasan perubahan -> saran nomor versi (0.1.2, 0.2.0, ...).