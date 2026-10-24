#!/usr/bin/env bash
# =====================================================================
# verifikasi.sh - pemeriksa kode SI'OM (frontend)
# Jalankan dari folder frontend:   bash verifikasi.sh
# Atau dari mana saja:             bash verifikasi.sh /jalur/ke/frontend
# Keluar dengan kode 0 bila semua lulus, 1 bila ada yang gagal.
# =====================================================================
cd "${1:-.}" || { echo "Folder tidak ditemukan"; exit 2; }

G=$'\e[32m'; R=$'\e[31m'; Y=$'\e[33m'; B=$'\e[1m'; N=$'\e[0m'
lulus_n=0; gagal_n=0; waspada_n=0

lulus()  { echo "${G}PASS${N}  $1"; lulus_n=$((lulus_n + 1)); }
gagal()  { echo "${R}FAIL${N}  $1"; gagal_n=$((gagal_n + 1)); }
waspada(){ echo "${Y}WARN${N}  $1"; waspada_n=$((waspada_n + 1)); }
judul()  { echo; echo "${B}== $1 ==${N}"; }

# ada "judul" berkas regex  -> harus ada minimal satu
ada() {
  [ -f "$2" ] || { gagal "$1 (berkas $2 tidak ada)"; return; }
  if grep -qE -- "$3" "$2"; then lulus "$1"; else gagal "$1  [$2]"; fi
}
# tak "judul" berkas regex  -> tidak boleh ada
tak() {
  [ -f "$2" ] || { gagal "$1 (berkas $2 tidak ada)"; return; }
  if grep -qE -- "$3" "$2"; then
    gagal "$1  [$2 baris $(grep -nE -- "$3" "$2" | head -3 | cut -d: -f1 | paste -sd, -)]"
  else lulus "$1"; fi
}
# opsional "judul" berkas regex -> bila ada, hanya peringatan
opsional() {
  [ -f "$2" ] || return
  if grep -qE -- "$3" "$2"; then waspada "$1  [$2]"; fi
}
# minimal "judul" berkas regex n
minimal() {
  [ -f "$2" ] || { gagal "$1 (berkas $2 tidak ada)"; return; }
  local n; n=$(grep -oE -- "$3" "$2" | wc -l)
  if [ "$n" -ge "$4" ]; then lulus "$1 ($n)"; else gagal "$1: ditemukan $n, seharusnya minimal $4  [$2]"; fi
}
# tepat "judul" berkas regex n
tepat() {
  [ -f "$2" ] || { gagal "$1 (berkas $2 tidak ada)"; return; }
  local n; n=$(grep -oE -- "$3" "$2" | wc -l)
  if [ "$n" -eq "$4" ]; then lulus "$1 ($n)"; else gagal "$1: ditemukan $n, seharusnya tepat $4  [$2]"; fi
}
# blok "judul" berkas awal jumlah_baris teks_literal
# -> teks literal harus muncul dalam N baris sejak baris yang memuat "awal"
blok() {
  [ -f "$2" ] || { gagal "$1 (berkas $2 tidak ada)"; return; }
  if awk -v s="$3" -v n="$4" -v lit="$5" '
      !f && index($0, s) { f = 1 }
      f { c++; if (index($0, lit)) ok = 1; if (c >= n) exit }
      END { exit !ok }' "$2"; then lulus "$1"; else gagal "$1  [$2]"; fi
}

[ -f index.html ] || { echo "${R}index.html tidak ada di $(pwd). Jalankan dari folder frontend.${N}"; exit 2; }

# ---------------------------------------------------------------------
judul "css/style.css"
C=css/style.css
ada  "token --di-atas-hijau"                 $C '^\s*--di-atas-hijau:'
ada  "strip tiga warna (body::before)"       $C 'body::before'
ada  ".aplikasi memakai calc(100vh - 6px)"   $C 'calc\(100vh - 6px\)'
ada  "keyframes getar"                       $C '@keyframes getar'
ada  ".kotak.getar"                          $C '\.kotak\.getar'
ada  "keyframes kedip"                       $C '@keyframes kedip'
ada  "tr.berubah td"                         $C 'tr\.berubah td'
ada  ".mode.bahaya"                          $C '\.mode\.bahaya'
ada  ".kotak.bahaya"                         $C '\.kotak\.bahaya'
ada  "tombol merah solid (bahaya-solid)"     $C 'button\.bahaya-solid:last-child'
ada  "ikon suara mati/nyala"                 $C '\[data-suara="mati"\]'
ada  ".lencana-versi"                        $C '\.lencana-versi'
ada  "chip lokasi data (button.jalur)"       $C 'button\.jalur'
ada  ".jalur-label"                          $C '\.jalur-label'
ada  ".info-sistem"                          $C '\.info-sistem'
ada  ".judul-topbar"                         $C '\.judul-topbar'
tak  "tidak ada backdrop-filter blur"        $C 'backdrop-filter:[[:space:]]*blur'
tak  "tidak ada bentuk pil 999px"            $C '999px'
tak  "token --segmen-* sudah dibuang"        $C '--segmen-'
tak  "keyframes denyut sudah dibuang"        $C '@keyframes denyut'
if [ -f $C ]; then
  o=$(tr -cd '{' < $C | wc -c); c=$(tr -cd '}' < $C | wc -c)
  if [ "$o" -eq "$c" ]; then lulus "kurung kurawal seimbang ($o pasang)"; else gagal "kurung kurawal tidak seimbang: { = $o, } = $c"; fi
fi

# ---------------------------------------------------------------------
judul "index.html"
H=index.html
ada  "tombol suara #btn-suara"               $H 'id="btn-suara"'
ada  "ikon suara nyala"                      $H 'ikon-suara-nyala'
ada  "ikon suara mati"                       $H 'ikon-suara-mati'
ada  "lencana versi #info-versi"             $H 'id="info-versi"'
ada  "chip lokasi #info-data"                $H 'id="info-data"'
ada  "teks lokasi #info-data-teks"           $H 'id="info-data-teks"'
ada  "pembungkus .info-sistem"               $H 'class="info-sistem"'
tak  "<p id=info-sistem> lama sudah diganti" $H '<p class="kecil" id="info-sistem"'
ada  "modal #modal-isi"                      $H 'id="modal-isi"'
ada  "stylesheet css/style.css"              $H 'css/style.css'
opsional "belum ada judul topbar #judul-halaman (opsional)" $H 'NEVER_MATCH_PLACEHOLDER'
if ! grep -q 'id="judul-halaman"' $H; then waspada "#judul-halaman belum dipasang di topbar (opsional, hanya jika ingin judul di header)"; else lulus "#judul-halaman terpasang"; fi

# urutan pemuatan skrip
urut=(api tap dashboard master riwayat ekspor sinkron pembaruan wa logo main)
prev=0; benar=1
for n in "${urut[@]}"; do
  l=$(grep -n "src=\"js/$n.js\"" $H | head -1 | cut -d: -f1)
  if [ -z "$l" ]; then benar=0; echo "      skrip js/$n.js tidak dimuat"; continue; fi
  [ "$l" -gt "$prev" ] || { benar=0; echo "      js/$n.js salah urutan"; }
  prev=$l
done
if [ $benar -eq 1 ]; then lulus "urutan skrip: api > tap > dashboard > master > riwayat > ekspor > sinkron > pembaruan > wa > logo > main"; else gagal "urutan atau kelengkapan skrip salah"; fi

# ---------------------------------------------------------------------
judul "js/api.js"
A=js/api.js
ada  "katalog nada POLA"                     $A 'const POLA = \{'
ada  "objek Suara"                           $A 'const Suara = \{'
ada  "Suara.izinkan()"                       $A 'izinkan\(ms'
ada  "Suara.cabut()"                         $A 'cabut\(\)'
ada  "Suara.tahan()"                         $A 'tahan\(ms'
ada  "daftar TERTAHAN"                       $A 'const TERTAHAN'
ada  "daftar AKSI"                           $A 'const AKSI'
ada  "bunyi(tipe, paksa)"                    $A 'function bunyi\(tipe, paksa\)'
ada  "tab mencabut izin"                     $A 'tipe === "tab"\) Suara\.cabut'
ada  "getarKotak"                            $A 'function getarKotak'
ada  "toast dengan parameter suara"          $A 'function toast\(teks, tipe = "info", suara = tipe\)'
ada  "jaring unhandledrejection"             $A 'unhandledrejection'
ada  "izin dari event change"                $A 'addEventListener\("change"'
ada  "izin dari tombol Enter"                $A 'e\.key === "Enter"\) Suara\.izinkan'
ada  "tombol bisukan terhubung"              $A 'getElementById\("btn-suara"\)'
ada  "fungsi salin()"                        $A 'function salin\('
ada  "klik tab memakai nada tab"             $A 'bunyi\("tab"\)'
opsional "konfirmasi() lama masih ada, sudah tidak dipakai (boleh dihapus)" $A 'function konfirmasi\('

# ---------------------------------------------------------------------
judul "js/dashboard.js"
D=js/dashboard.js
ada  "flag _aksi"                            $D '_aksi: false'
ada  "fanfare hanya untuk aksi (dariAksi)"   $D 'dariAksi && this\._semua === false'
ada  "submit menandai this._aksi = true"     $D 'this\._aksi = true'
ada  "Dash.konfirmasi(o)"                    $D 'konfirmasi\(o\)'
ada  "bukaModal mereset className"           $D 'k\.className = "kotak"'
ada  "bukaModal menerima senyap"             $D 'bukaModal\(html, senyap\)'
ada  "kedipan baris (berubah)"               $D 'classList\.add\("berubah"\)'
ada  "nada ambil/kembali di popupJumlah"     $D 'bunyi\(ambil \? "ambil" : "kembali"\)'
ada  "nada peringatan di popupKonfirmasi"    $D 'popupKonfirmasi\(r\) \{'
ada  "prosesTap memakai bunyi(tap)"          $D 'bunyi\("tap"\)'
ada  "toast nama siswa senyap"               $D '"info", false\)'
tak  "tidak ada bunyi(ok) liar di prosesTap" $D 'bunyi\("ok"\);'
ada  "Esc membatalkan modal konfirmasi"      $D 'e\.key === "Escape"'
ada  "fokus awal di Batal"                   $D 'kf-batal"\)\.focus'

# ---------------------------------------------------------------------
judul "js/master.js"
M=js/master.js
tepat "tiga modal Dash.konfirmasi (kelas, wali, kartu)" $M 'Dash\.konfirmasi\(' 3
tak  "tidak ada confirm() bawaan browser"    $M '(^|[^.A-Za-z_])confirm\('
tak  "tidak ada konfirmasi() lama"           $M '(^|[^.A-Za-z_])konfirmasi\('
minimal "toast nada hapus (kelas, wali, kartu)" $M '"ok", "hapus"' 3
ada  "toast wali diaktifkan"                 $M 'Wali kelas diaktifkan'
ada  "lencana versi diisi"                   $M 'getElementById\("info-versi"\)'
ada  "lokasi data diisi"                     $M 'info-data-teks'
ada  "dataset.salin"                         $M 'dataset\.salin'
tak  "baris textContent info-sistem lama sudah diganti" $M 'info-sistem"\)\.textContent'
ada  "alasan dari modal dikirim ke backend"  $M 'h\.alasan \|\| "dari layar master"'

# ---------------------------------------------------------------------
judul "js/main.js"
MN=js/main.js
ada  "izin saat tap serial (card:tap)"       $MN 'Suara\.izinkan\(\)'
blok "izinkan berada di dalam handler card:tap" $MN 'EventsOn("card:tap"' 4 'Suara.izinkan()'
ada  "variabel _readerOk"                    $MN 'let _readerOk'
ada  "reader terputus dipaksa berbunyi"      $MN 'bunyi\(ok \? "ok" : "err", true\)'
ada  "tap di tab salah: nada peringatan"     $MN 'Pindah ke tab Beranda.*"peringatan"'
ada  "segar berkala dibisukan"               $MN 'Suara\.tahan\(1500\)'
ada  "klik chip lokasi menyalin"             $MN 'getElementById\("info-data"\)'
ada  "memanggil salin()"                     $MN 'await salin\('

# ---------------------------------------------------------------------
judul "js/tap.js"
ada  "izin saat HID terbaca"                 js/tap.js 'Suara\.izinkan\(\)'

# ---------------------------------------------------------------------
judul "js/riwayat.js"
ada  "hapus transaksi: nada hapus"           js/riwayat.js 'Transaksi dihapus", "ok", "hapus"'
ada  "tutup hari: nada selesai"              js/riwayat.js 'Hari ditutup", "ok", "selesai"'

# ---------------------------------------------------------------------
judul "js/ekspor.js"
ada  "baris kosong seragam dengan Riwayat"   js/ekspor.js 'const kosong ='
ada  "colspan 11 pada baris kosong"          js/ekspor.js 'colspan="11"'
ada  "galat ditampilkan di baris kosong"     js/ekspor.js 'kosong\(String\(e\.message \|\| e\)\)'

# ---------------------------------------------------------------------
judul "js/sinkron.js dan js/pembaruan.js"
ada  "sinkron: tap ditolak bernada peringatan" js/sinkron.js 'Selesaikan impor dulu", "info", "peringatan"'
ada  "pembaruan: tap ditolak bernada peringatan" js/pembaruan.js 'Selesaikan pembaruan dulu", "info", "peringatan"'
tak  "pembaruan: tidak ada bunyi(info) saat mulai" js/pembaruan.js 'bunyi\("info"\)'
minimal "pembaruan: bunyi(selesai) di pasang & hasilLinux" js/pembaruan.js 'bunyi\("selesai"\)' 2
blok "hasilLinux: modal dibuka senyap"       js/pembaruan.js 'hasilLinux(h) {' 14 '      true'

# ---------------------------------------------------------------------
judul "js/wa.js"
ada  "pesan() punya parameter senyap"        js/wa.js 'const pesan = \(id, teks, galat, senyap\)'
ada  "pesan() membunyikan hasil"             js/wa.js 'bunyi\(galat \? "err" : "ok"\)'
ada  "Mengirim... senyap"                    js/wa.js 'Mengirim\.\.\.", false, true'
ada  "galat muat konfigurasi senyap"         js/wa.js 'String\(e\), true, true'
ada  "alert(e) didahului bunyi err"          js/wa.js 'bunyi\("err"\);[[:space:]]*alert\(e\)'
ada  "toast Diantrekan ulang"                js/wa.js 'Diantrekan ulang'

# ---------------------------------------------------------------------
judul "Kesesuaian id antara HTML dan JS"
tmp_html=$(mktemp); tmp_pakai=$(mktemp)
grep -ohE 'id="[^"]+"' index.html js/*.js 2>/dev/null | sed -E 's/id="([^"]+)"/\1/' | sort -u > "$tmp_html"
grep -ohE '(getElementById|_r|\$|val|setv)\("[^"]+"' js/*.js 2>/dev/null | sed -E 's/.*\("([^"]+)"/\1/' | sort -u > "$tmp_pakai"
hilang=$(comm -13 "$tmp_html" "$tmp_pakai")
if [ -z "$hilang" ]; then lulus "semua id yang dipakai JS ada di HTML atau template"; else
  while read -r i; do gagal "id \"$i\" dipakai JS tetapi tidak ada di HTML maupun template"; done <<< "$hilang"
fi
rm -f "$tmp_html" "$tmp_pakai"

# ---------------------------------------------------------------------
judul "Sintaks JavaScript"
if command -v node >/dev/null 2>&1; then
  for f in js/*.js; do
    if node --check "$f" 2>/tmp/siom_err; then lulus "sintaks $f"; else gagal "sintaks $f"; sed 's/^/      /' /tmp/siom_err | head -5; fi
  done
  # gabungan sesuai urutan muat: menangkap const/let ganda antar berkas
  gab=$(mktemp --suffix=.js)
  for n in "${urut[@]}"; do [ -f "js/$n.js" ] && { cat "js/$n.js"; echo; } >> "$gab"; done
  if node --check "$gab" 2>/tmp/siom_err; then lulus "gabungan semua berkas tidak bentrok (const/let ganda)"; else gagal "gabungan berkas bentrok"; sed 's/^/      /' /tmp/siom_err | head -5; fi
  rm -f "$gab" /tmp/siom_err
else
  waspada "node tidak terpasang, pemeriksaan sintaks dilewati (jalankan: node --check js/nama.js)"
fi

# ---------------------------------------------------------------------
echo
echo "${B}Ringkasan:${N} ${G}${lulus_n} lulus${N}, ${R}${gagal_n} gagal${N}, ${Y}${waspada_n} peringatan${N}"
if [ "$gagal_n" -eq 0 ]; then
  echo "${G}Semua kode statis sudah benar.${N} Lanjut uji suara secara manual:"
else
  echo "${R}Ada yang belum sesuai. Perbaiki baris FAIL di atas, lalu jalankan ulang.${N}"
fi
cat <<'EOF'

Uji manual (suara tidak bisa diperiksa dari kode):
  1. Pindah cepat antar Beranda, Master Data, Riwayat, Ekspor dan sub-tab WhatsApp:
     hanya terdengar tik pendek, tanpa buzzer, bip sukses, atau fanfare.
  2. Klik Simpan di Pengaturan: bunyi sukses keluar.
  3. Ubah dropdown wali kelas: bunyi "Wali kelas disimpan" keluar.
  4. Klik Nonaktifkan kelas/wali dan Hapus kartu: modal merah muncul, fokus di Batal, Esc membatalkan.
  5. Tap kartu: bip ganda, lalu nada naik (ambil) atau turun (kembali).
  6. Cabut reader: nada galat tetap berbunyi walau tidak menyentuh apa pun.
  7. Tab Ekspor tanpa data: baris TIDAK ADA DATA muncul di dalam tabel.
  8. Klik chip lokasi data di Pengaturan: tersalin dan label berubah jadi TERSALIN.
EOF
[ "$gagal_n" -eq 0 ]
