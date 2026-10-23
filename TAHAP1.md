ya lanjutkan.


Sebelum lanjut, satu catatan: saya belum menerima hasil `go test` untuk Fitur 3. Kalau ada yang gagal, kirim dulu, karena itu saya perbaiki sebelum menumpuk fitur baru. Kalau lulus, abaikan saja.

Sesuai SPEC, berikut **rencana Fitur 1** untuk Anda setujui. Belum ada kode. Saya belum melihat `master.go`, `master.js`, `main.js`, dan `main.go`, jadi bagian yang menyentuh berkas itu masih asumsi dan saya tandai.

## 1. Rancangan

**Format berkas**
- **Excel (.xlsx):** sheet `Pengaturan` (kunci/nilai), `Kelas` (nama, nama_ketua, urutan, aktif), `Kartu` (uid, kelas, nisn, nama_siswa, aktif). Kolom UID dan NISN ditulis sebagai teks dengan format sel `@`, jadi angka nol di depan tidak hilang.
- **CSV:** pengguna memilih jenis data (kelas atau kartu). UTF-8 dengan BOM, pemisah koma.
- **Template kosong:** xlsx berisi header, satu baris contoh, dan sheet `Petunjuk`. Ada juga template CSV.
- Baris `kartu.kelas` merujuk **nama kelas**, bukan ID, supaya portabel antar komputer.

**Alur impor (dua tahap, tanpa status tersimpan)**
1. `PratinjauImpor(opsi)` membuka dialog berkas, mem-parsing, lalu menghitung per baris: **baru / diubah / sama / dilewati / galat**. Tidak ada yang ditulis ke database.
2. `TerapkanImpor(path, opsi, alasan)` mem-parsing ulang berkas yang sama, memeriksa bahwa hasil hitungannya sama dengan pratinjau (hash ringkasan), lalu menulis dalam **satu transaksi**. Jika ada satu galat, seluruhnya dibatalkan.
3. Galat yang dideteksi: UID bukan angka 1-10 digit, UID ganda dalam berkas, nama kelas ganda, kelas kosong atau tidak dikenal, NISN bukan angka (maks 10 digit), ketua kosong, dan kolom wajib tidak ada.

**Mode**
- **Gabung:** tambah dan perbarui saja.
- **Ganti:** kartu yang tidak ada di berkas dihapus. Wajib mencentang konfirmasi dan mengisi alasan, dan pratinjau menampilkan jumlah yang akan dihapus.
- **Kelas tidak dikenal pada kartu:** opsi "Buat kelas otomatis" (ketua diisi `-`) atau tolak.
- **Audit:** satu entri ringkasan (`IMPOR_MASTER`) ditambah satu entri per kartu yang berubah, ditambah entri per kelas yang berubah.

## 2. Berkas yang disentuh

| Berkas | Perubahan |
|---|---|
| `internal/model/model.go` | tipe `OpsiImpor`, `PratinjauImpor`, `BarisPratinjau`, `HasilImpor` |
| `internal/sinkron/sinkron.go` (baru) | parse dan tulis xlsx/csv, validasi tanpa akses DB, mudah dites |
| `internal/service/sinkron.go` (baru) | `EksporMasterExcel`, `EksporMasterCSV(jenis)`, `UnduhTemplate`, `PratinjauImpor`, `TerapkanImpor` |
| `frontend/index.html` | sub-tab **Sinkronisasi** di Master Data, dengan peringatan data pribadi |
| `frontend/js/sinkron.js` (baru) dan `master.js` | tombol dan modal pratinjau |
| `internal/service/sinkron_test.go` (baru) | tes di bawah |

Tidak ada perubahan skema. Excelize sudah menjadi dependensi, jadi `go.mod` tidak bertambah.

## 3. Keputusan yang perlu persetujuan Anda

1. **Mode Ganti pada kelas.** Kelas yang punya transaksi tidak bisa dihapus karena ada foreign key. Usul saya: mode Ganti **menghapus kartu** yang tidak ada di berkas, sedangkan **kelas** yang tidak ada di berkas **dinonaktifkan** (`aktif=0`), bukan dihapus. Kelas yang sama sekali tanpa transaksi tetap bisa saya hapus bila Anda lebih suka. Setuju dinonaktifkan saja?
2. **Logo.** Usul saya logo **tidak diikutkan** pada ekspor dan impor, hanya teks pengaturan, karena logo bisa sampai 1 MB dan jarang berubah. Setuju?
3. **Pengaturan lokal.** `port_serial` dan `hid_ambang_ms` bergantung pada perangkat. Usul saya **hanya empat teks instansi** yang disinkronkan (nama, alamat, penanggung jawab, NIP, jabatan) ditambah jeda minimum. Port serial dan ambang HID tidak diikutkan. Setuju?
4. **Nol di depan yang sudah hilang.** Jika berkas Excel sudah kehilangan nol (UID `12345678` yang seharusnya `0012345678`), `NormalisasiUID` yang sudah ada menanganinya, dan NISN yang kurang dari 10 digit saya beri peringatan di pratinjau, bukan galat. Setuju?
5. **Lokasi berkas.** Dialog simpan dibuka di folder Dokumen atau Unduhan pengguna. Jika path pilihan berada di dalam folder yang punya `.git` di induknya, ekspor ditolak dengan pesan peringatan privasi. Setuju?

## 4. Tes yang akan ditulis

- **Round-trip:** ekspor, lalu impor ke DB kosong menghasilkan kelas, kartu, dan pengaturan identik, termasuk nol di depan UID dan NISN.
- UID ganda dalam berkas, UID tidak valid, dan kelas tidak dikenal (ditolak, lalu dibuat otomatis).
- Pratinjau tidak menulis apa pun ke DB.
- Mode Gabung tidak menghapus, dan mode Ganti menghapus kartu serta menonaktifkan kelas.
- Satu galat berarti seluruh impor batal (DB tetap utuh).
- Entri audit: satu ringkasan dan satu per perubahan.
- CSV memuat BOM dan terbaca ulang, dan template bisa diimpor tanpa galat.

Saran versi untuk fitur ini **0.3.0**, karena ada kemampuan baru tanpa perubahan skema. Saya konfirmasi lagi di ringkasan akhir.

## 5. Yang saya perlukan

Balas dengan keputusan 1 sampai 5 (cukup "setuju semua" atau sebutkan yang ingin diubah), lalu jalankan:

```bash
for f in internal/service/master.go frontend/js/master.js frontend/js/main.js frontend/js/ekspor.js main.go go.mod internal/reader/uid.go; do echo "=== $f ==="; cat "$f"; done
```

Berkas `master.go` penting karena saya harus memakai ulang validasi `DaftarkanKartu` dan pola audit kartu/kelas yang sudah ada, supaya impor memberi hasil yang sama dengan input manual.

zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ for f in internal/service/master.go frontend/js/master.js frontend/js/main.js frontend/js/ekspor.js main.go go.mod internal/reader/uid.go; do echo "=== $f ==="; cat "$f"; done
=== internal/service/master.go ===
package service

import (
	"errors"
	"strings"

	"mbg/internal/model"
	"mbg/internal/reader"
)

// ---------- Pengaturan ----------

func (s *Service) GetPengaturan() (model.Pengaturan, error) {
	var p model.Pengaturan
	err := s.db.QueryRow(`SELECT COALESCE(nama_instansi,''),COALESCE(alamat_instansi,''),
		COALESCE(penanggung_jawab,''),COALESCE(nip_penanggung_jawab,''),COALESCE(jabatan_penanggung_jawab,''),
		jeda_minimum_menit,hid_ambang_ms,COALESCE(port_serial,'') FROM pengaturan WHERE id=1`).
		Scan(&p.NamaInstansi, &p.AlamatInstansi, &p.PenanggungJawab, &p.NIPPenanggungJawab,
			&p.JabatanPenanggungJawab, &p.JedaMinimumMenit, &p.HIDAmbangMs, &p.PortSerial)
	return p, err
}

func (s *Service) SimpanPengaturan(p model.Pengaturan) error {
	if p.JedaMinimumMenit < 0 || p.JedaMinimumMenit > 600 {
		return errors.New("Jeda minimum harus 0-600 menit")
	}
	if p.HIDAmbangMs < 10 || p.HIDAmbangMs > 500 {
		return errors.New("Ambang HID harus 10-500 ms")
	}
	lama, _ := s.GetPengaturan()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`UPDATE pengaturan SET nama_instansi=?,alamat_instansi=?,penanggung_jawab=?,
		nip_penanggung_jawab=?,jabatan_penanggung_jawab=?,jeda_minimum_menit=?,hid_ambang_ms=?,port_serial=? WHERE id=1`,
		p.NamaInstansi, p.AlamatInstansi, p.PenanggungJawab, p.NIPPenanggungJawab,
		p.JabatanPenanggungJawab, p.JedaMinimumMenit, p.HIDAmbangMs, p.PortSerial)
	if err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "pengaturan", "1", lama, p, ""); err != nil {
		return err
	}
	return tx.Commit()
}

// ---------- Kelas ----------

func (s *Service) ListKelas(termasukNonaktif bool) ([]model.Kelas, error) {
	q := `SELECT id,nama,nama_ketua,urutan,aktif FROM kelas`
	if !termasukNonaktif {
		q += ` WHERE aktif=1`
	}
	q += ` ORDER BY urutan,nama`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Kelas{}
	for rows.Next() {
		var k model.Kelas
		var a int
		if err := rows.Scan(&k.ID, &k.Nama, &k.NamaKetua, &k.Urutan, &a); err != nil {
			return nil, err
		}
		k.Aktif = a == 1
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *Service) TambahKelas(nama, ketua string, urutan int) (int64, error) {
	nama, ketua = strings.TrimSpace(nama), strings.TrimSpace(ketua)
	if nama == "" || ketua == "" {
		return 0, errors.New("Nama kelas dan nama ketua wajib diisi")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO kelas (nama,nama_ketua,urutan) VALUES (?,?,?)`, nama, ketua, urutan)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, errors.New("Nama kelas sudah ada")
		}
		return 0, err
	}
	id, _ := res.LastInsertId()
	if err := s.audit(tx, s.now(), "TAMBAH", "kelas", itoa(id), nil, map[string]any{"nama": nama, "ketua": ketua, "urutan": urutan}, ""); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (s *Service) UbahKelas(id int64, nama, ketua string, urutan int) error {
	nama, ketua = strings.TrimSpace(nama), strings.TrimSpace(ketua)
	if nama == "" || ketua == "" {
		return errors.New("Nama kelas dan nama ketua wajib diisi")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var lama model.Kelas
	if err := tx.QueryRow(`SELECT id,nama,nama_ketua,urutan FROM kelas WHERE id=?`, id).
		Scan(&lama.ID, &lama.Nama, &lama.NamaKetua, &lama.Urutan); err != nil {
		return errors.New("Kelas tidak ditemukan")
	}
	if _, err := tx.Exec(`UPDATE kelas SET nama=?,nama_ketua=?,urutan=? WHERE id=?`, nama, ketua, urutan, id); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("Nama kelas sudah ada")
		}
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "kelas", itoa(id), lama, map[string]any{"nama": nama, "ketua": ketua, "urutan": urutan}, ""); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) NonaktifkanKelas(id int64, alasan string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE kelas SET aktif=0 WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "kelas", itoa(id), map[string]any{"aktif": 1}, map[string]any{"aktif": 0}, alasan); err != nil {
		return err
	}
	return tx.Commit()
}

// ---------- Kartu ----------

func (s *Service) ListKartu(kelasID int64) ([]model.Kartu, error) {
	rows, err := s.db.Query(`SELECT uid,kelas_id,aktif,COALESCE(nisn,''),COALESCE(nama_siswa,'') FROM kartu WHERE kelas_id=? ORDER BY nama_siswa,uid`, kelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Kartu{}
	for rows.Next() {
		var k model.Kartu
		var a int
		if err := rows.Scan(&k.UID, &k.KelasID, &a, &k.NISN, &k.NamaSiswa); err != nil {
			return nil, err
		}
		k.Aktif = a == 1
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *Service) DaftarkanKartu(uid string, kelasID int64, nisn, nama string) error {
	u, ok := reader.NormalisasiUID(uid)
	if !ok {
		return errors.New("UID tidak valid (harus 1-10 digit angka)")
	}
	nisn, nama = strings.TrimSpace(nisn), strings.TrimSpace(nama)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var namaKelas string
	if err := tx.QueryRow(`SELECT k.nama FROM kartu c JOIN kelas k ON k.id=c.kelas_id WHERE c.uid=?`, u).Scan(&namaKelas); err == nil {
		return errors.New("Kartu " + u + " sudah terdaftar untuk kelas " + namaKelas)
	}
	if _, err := tx.Exec(`INSERT INTO kartu (uid,kelas_id,nisn,nama_siswa) VALUES (?,?,?,?)`, u, kelasID, nisn, nama); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "TAMBAH", "kartu", u, nil, map[string]any{"kelas_id": kelasID, "nisn": nisn, "nama": nama}, ""); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) HapusKartu(uid string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var kid int64
	if err := tx.QueryRow(`SELECT kelas_id FROM kartu WHERE uid=?`, uid).Scan(&kid); err != nil {
		return errors.New("Kartu tidak ditemukan")
	}
	if _, err := tx.Exec(`DELETE FROM kartu WHERE uid=?`, uid); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "HAPUS", "kartu", uid, map[string]any{"kelas_id": kid}, nil, ""); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) AktifkanKelas(id int64, alasan string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE kelas SET aktif=1 WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "kelas", itoa(id), map[string]any{"aktif": 0}, map[string]any{"aktif": 1}, alasan); err != nil {
		return err
	}
	return tx.Commit()
}
=== frontend/js/master.js ===
const Master = {
    editId: null,

    async muatInstansi() {
        const p = await API.GetPengaturan();
        document.getElementById("p-nama").value = p.nama_instansi;
        document.getElementById("p-alamat").value = p.alamat_instansi;
        document.getElementById("p-pj").value = p.penanggung_jawab;
        document.getElementById("p-nip").value = p.nip_penanggung_jawab;
        document.getElementById("p-jabatan").value = p.jabatan_penanggung_jawab;
        document.getElementById("p-jeda").value = p.jeda_minimum_menit;
        document.getElementById("p-hid").value = p.hid_ambang_ms;
        Tap.ambangMs = p.hid_ambang_ms;
        const i = await API.InfoSistem();
        document.getElementById("info-sistem").textContent = `Versi ${i.versi} - data: ${i.data_dir}`;
    },
    async simpanInstansi() {
        try {
            const lama = await API.GetPengaturan();
            lama.nama_instansi = val("p-nama"); lama.alamat_instansi = val("p-alamat");
            lama.penanggung_jawab = val("p-pj"); lama.nip_penanggung_jawab = val("p-nip");
            lama.jabatan_penanggung_jawab = val("p-jabatan");
            lama.jeda_minimum_menit = parseInt(val("p-jeda"), 10);
            lama.hid_ambang_ms = parseInt(val("p-hid"), 10);
            await API.SimpanPengaturan(lama);
            Tap.ambangMs = lama.hid_ambang_ms;
            toast("Pengaturan disimpan", "ok");
        } catch (e) { toast(String(e.message || e), "err"); }
    },

    async muatKelas() {
        const cb = document.getElementById("k-semua");
        const list = await API.ListKelas(true);
        const tb = document.querySelector("#tabel-kelas tbody"); tb.innerHTML = "";
        list.forEach(k => {
            const tr = document.createElement("tr");
            tr.innerHTML = `<td>${k.urutan}</td><td>${esc(k.nama)}</td><td>${esc(k.nama_ketua)}</td><td class="aksi-sel"><div class="aksi-baris"></div></td>`;
            const wadah = tr.querySelector(".aksi-baris");
            if (k.aktif) {
                const ub = document.createElement("button"); ub.textContent = "Ubah";
                ub.onclick = () => {
                    this.editId = k.id; setv("k-nama", k.nama); setv("k-ketua", k.nama_ketua); setv("k-urutan", k.urutan);
                    document.getElementById("k-simpan").textContent = "Simpan"; document.getElementById("k-batal").classList.remove("hidden");
                };
                const nb = document.createElement("button"); nb.textContent = "Nonaktifkan"; nb.className = "bahaya";
                nb.onclick = async () => {
                    if (!confirm(`Nonaktifkan kelas ${k.nama}?`)) return;
                    try { await API.NonaktifkanKelas(k.id, "dari layar master"); this.muatKelas(); Dash.muat(); } catch (e) { toast(String(e.message || e), "err"); }
                };
                wadah.append(ub, nb);
            } else {
                tr.style.opacity = "0.55";
                const ab = document.createElement("button"); ab.textContent = "Aktifkan";
                ab.onclick = async () => {
                    try { await API.AktifkanKelas(k.id, "dari layar master"); this.muatKelas(); Dash.muat(); toast("Kelas diaktifkan", "ok"); }
                    catch (e) { toast(String(e.message || e), "err"); }
                };
                wadah.append(ab);
            }
            tb.appendChild(tr);
        });
        // isi dropdown kelas di tab kartu
        const sel = document.getElementById("c-kelas"); const cur = sel.value; sel.innerHTML = "";
        list.filter(k => k.aktif).forEach(k => sel.add(new Option(k.nama + " - " + k.nama_ketua, k.id)));
        if (cur) sel.value = cur;
        this.muatKartu();
    },
    resetFormKelas() {
        this.editId = null; setv("k-nama", ""); setv("k-ketua", ""); setv("k-urutan", 0);
        document.getElementById("k-simpan").textContent = "Tambah"; document.getElementById("k-batal").classList.add("hidden");
    },
    async simpanKelas() {
        try {
            const n = val("k-nama"), k = val("k-ketua"), u = parseInt(val("k-urutan") || "0", 10);
            if (this.editId) await API.UbahKelas(this.editId, n, k, u); else await API.TambahKelas(n, k, u);
            this.resetFormKelas(); toast("Kelas tersimpan", "ok"); this.muatKelas(); Dash.muat();
        } catch (e) { toast(String(e.message || e), "err"); }
    },

    async muatKartu() {
        const kid = parseInt(val("c-kelas"), 10);
        const tb = document.querySelector("#tabel-kartu tbody"); tb.innerHTML = "";
        if (!kid) return;
        (await API.ListKartu(kid)).forEach(c => {
            const tr = document.createElement("tr");
            tr.innerHTML = `<td>${esc(c.uid)}</td><td>${esc(c.nisn || "-")}</td><td>${esc(c.nama_siswa || "-")}</td><td class="aksi-sel"><div class="aksi-baris"></div></td>`;
            const b = document.createElement("button"); b.textContent = "Hapus"; b.className = "bahaya";
            b.onclick = async () => {
                if (!confirm("Hapus kartu " + c.uid + "?")) return;
                try { await API.HapusKartu(c.uid); this.muatKartu(); } catch (e) { toast(String(e.message || e), "err"); }
            };
            tr.querySelector(".aksi-baris").appendChild(b); tb.appendChild(tr);
        });
    },
    async daftarKartu() {
        try {
            const kid = parseInt(val("c-kelas"), 10);
            if (!kid) throw new Error("Pilih kelas dulu");
            await API.DaftarkanKartu(val("c-uid"), kid, val("c-nisn").trim(), val("c-nama").trim());
            setv("c-uid", ""); setv("c-nisn", ""); setv("c-nama", "");
        } catch (e) { toast(String(e.message || e), "err"); }
    },
};

function val(id) { return document.getElementById(id).value; }
function setv(id, v) { document.getElementById(id).value = v; }=== frontend/js/main.js ===
let tabAktif = "beranda", subAktif = "instansi";

function pilihTab(t) {
    tabAktif = t;
    document.querySelectorAll("header nav button").forEach(b => b.classList.toggle("aktif", b.dataset.tab === t));
    document.querySelectorAll(".tab").forEach(s => s.classList.toggle("aktif", s.id === "tab-" + t));
    if (t === "master") Master.muatInstansi(), Master.muatKelas(); else if (t === "riwayat") Riwayat.muat(); else if (t === "ekspor") Ekspor.muat(); else Dash.muat();
    pasangHandler();
}
function pilihSub(s) {
    subAktif = s;
    document.querySelectorAll(".subnav button").forEach(b => b.classList.toggle("aktif", b.dataset.sub === s));
    document.querySelectorAll(".sub").forEach(x => x.classList.toggle("aktif", x.id === "sub-" + s));
    pasangHandler();
}
// Di tab Master > Kartu, tap mengisi kolom UID; di Riwayat/Ekspor tap ditolak; selain itu diproses sebagai tap transaksi.
function pasangHandler() {
    if (tabAktif === "master" && subAktif === "kartu") {
        Tap.handler = uid => { setv("c-uid", uid); toast("UID terbaca: " + uid, "info"); };
    } else if (tabAktif === "riwayat" || tabAktif === "ekspor") {
        Tap.handler = () => toast("Pindah ke tab Beranda untuk memproses tap", "info");
    } else {
        Tap.handler = prosesTap;
    }
}

document.addEventListener("DOMContentLoaded", async () => {
    document.querySelectorAll("header nav button").forEach(b => b.onclick = () => pilihTab(b.dataset.tab));
    document.querySelectorAll(".subnav button").forEach(b => b.onclick = () => pilihSub(b.dataset.sub));
    document.getElementById("p-simpan").onclick = () => Master.simpanInstansi();
    document.getElementById("k-simpan").onclick = () => Master.simpanKelas();
    document.getElementById("k-batal").onclick = () => Master.resetFormKelas();
    document.getElementById("c-daftar").onclick = () => Master.daftarKartu();
    document.getElementById("c-kelas").onchange = () => Master.muatKartu();
    document.getElementById("k-semua").onchange = () => Master.muatKelas();
    document.getElementById("btn-uid-manual").onclick = () => {
        Dash.bukaModal(`<h2>Input UID manual</h2><input id="um" maxlength="12" placeholder="UID"><div class="galat"></div>
      <div class="aksi"><button id="um-batal">Batal</button><button id="um-ok">Proses</button></div>`);
        const i = document.getElementById("um"); i.focus();
        const ok = () => { const u = i.value; Dash.tutupModal(); prosesTap(u); };
        document.getElementById("um-ok").onclick = ok; i.onkeydown = e => { if (e.key === "Enter") ok(); };
        document.getElementById("um-batal").onclick = () => Dash.tutupModal();
    };
    document.getElementById("sim-kirim").onclick = () => prosesTap(val("sim-uid"));
    document.getElementById("sim-hid").onclick = () => simulasiHID(val("sim-uid"));

    try { const p = await API.GetPengaturan(); Tap.ambangMs = p.hid_ambang_ms; } catch (e) { }
    pasangHandler();
    Dash.muat();
    setInterval(() => { if (tabAktif === "beranda" && document.getElementById("modal").className === "hidden") Dash.muat(); }, 30000);
});

// ---- Reader serial: event dari backend ----
function tampilStatusReader(st) {
    const el = document.getElementById("status-reader");
    if (!el) return;
    el.textContent = st && st.terhubung ? "Reader serial terhubung (" + st.port + ")" : "Reader serial terputus";
    el.style.color = st && st.terhubung ? "#15803d" : "#b91c1c";
}
document.addEventListener("DOMContentLoaded", () => {
    if (window.runtime && window.runtime.EventsOn) {
        window.runtime.EventsOn("card:tap", uid => { if (Tap.handler) Tap.handler(uid); });
        window.runtime.EventsOn("reader:status", tampilStatusReader);
    }
    API.StatusReader().then(tampilStatusReader).catch(() => { });
});=== frontend/js/ekspor.js ===
const Ekspor = {
    rentang() { return [_r("ek-dari").value, _r("ek-sampai").value]; },

    async muat() {
        const st = await API.StatusHariIni();
        for (const id of ["ek-dari", "ek-sampai"]) if (!_r(id).value) _r(id).value = st.tanggal;
        await this.pratinjau();
    },

    async pratinjau() {
        const tb = document.querySelector("#tabel-ekspor tbody");
        tb.innerHTML = "";
        _r("ek-info").textContent = "";
        try {
            const [d, s] = this.rentang();
            const rows = await API.PratinjauLaporan(d, s);
            rows.forEach((b, i) => {
                const tr = document.createElement("tr");
                if (b.merah) tr.className = "merah";
                [i + 1, b.kelas, b.ketua, b.pengambil, b.pengembali, b.tanggal, b.waktu_ambil, b.waktu_kembali, b.jumlah_ambil, b.jumlah_kembali, b.keterangan].forEach(t => {
                    const td = document.createElement("td"); td.textContent = t; tr.appendChild(td);
                });
                tb.appendChild(tr);
            });
            _r("ek-info").textContent = rows.length + " baris (kelas yang tidak mengambil tidak dimuat).";
        } catch (e) { _r("ek-info").textContent = String(e.message || e); }
    },

    async unduh(jenis, d, s) {
        try {
            const p = jenis === "xlsx" ? await API.EksporExcel(d, s) : await API.EksporPDF(d, s);
            if (p) toast("Tersimpan: " + p, "ok");
        } catch (e) { toast(String(e.message || e), "err"); }
    },

    async statusLogo() {
        try { _r("logo-status").textContent = (await API.AdaLogo()) ? "Logo: terpasang" : "Logo: belum dipilih"; } catch (e) { }
    },
};

// logo di tab Master > Instansi
const _muatInstansiAsli = Master.muatInstansi.bind(Master);
Master.muatInstansi = async function () { await _muatInstansiAsli(); await Ekspor.statusLogo(); };

document.addEventListener("DOMContentLoaded", () => {
    _r("ek-lihat").onclick = () => Ekspor.pratinjau();
    _r("ek-xlsx").onclick = () => { const [d, s] = Ekspor.rentang(); Ekspor.unduh("xlsx", d, s); };
    _r("ek-pdf").onclick = () => { const [d, s] = Ekspor.rentang(); Ekspor.unduh("pdf", d, s); };
    _r("btn-cetak").onclick = async () => {
        try {
            const st = await API.StatusHariIni();
            await Ekspor.unduh("pdf", st.tanggal, st.tanggal);
        } catch (e) { toast(String(e.message || e), "err"); }
    };
    _r("logo-pilih").onclick = async () => {
        try { if (await API.PilihLogo()) toast("Logo tersimpan", "ok"); Ekspor.statusLogo(); }
        catch (e) { toast(String(e.message || e), "err"); }
    };
    _r("logo-hapus").onclick = async () => {
        try { await API.HapusLogo(); toast("Logo dihapus", "ok"); Ekspor.statusLogo(); }
        catch (e) { toast(String(e.message || e), "err"); }
    };
});=== main.go ===
package main

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/options/linux"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"mbg/internal/db"
	"mbg/internal/reader"
	"mbg/internal/service"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	dir := db.DataDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("gagal membuat folder data %s: %v", dir, err)
	}

	conn, err := db.Open(filepath.Join(dir, "mbg.db"))
	if err != nil {
		log.Fatalf("gagal membuka database: %v", err)
	}
	defer conn.Close()

	svc := service.New(conn, dir)

	var appCtx context.Context
	mgr := reader.NewManager(
		func() string { p, _ := svc.GetPengaturan(); return p.PortSerial },
		func(uid string) { wruntime.EventsEmit(appCtx, "card:tap", uid) },
		func(st reader.Status) {
			svc.SetStatusReader(st)
			if appCtx != nil {
				wruntime.EventsEmit(appCtx, "reader:status", st)
			}
		},
	)

	err = wails.Run(&options.App{
		Title:       "MBG - Pencatatan Ompreng",
		Width:       1200,
		Height:      800,
		AssetServer: &assetserver.Options{Assets: assets},
		Linux:       &linux.Options{Icon: icon, ProgramName: "mbg"},
		OnStartup: func(ctx context.Context) {
			appCtx = ctx
			svc.SetContext(ctx)
			go mgr.Run(ctx)
		},
		Bind: []interface{}{svc},
	})
	if err != nil {
		log.Fatalf("wails: %v", err)
	}
}
=== go.mod ===
module mbg

go 1.26.3

require (
	github.com/go-pdf/fpdf v0.9.0
	github.com/wailsapp/wails/v2 v2.17.0
	github.com/xuri/excelize/v2 v2.11.0
	go.bug.st/serial v1.8.0
	modernc.org/sqlite v1.60.1
)

require (
	git.sr.ht/~jackmordaunt/go-toast/v2 v2.0.3 // indirect
	github.com/bep/debounce v1.2.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jchv/go-winloader v0.0.0-20210711035445-715c2860da7e // indirect
	github.com/labstack/echo/v4 v4.15.3 // indirect
	github.com/labstack/gommon v0.5.0 // indirect
	github.com/leaanthony/go-ansi-parser v1.6.1 // indirect
	github.com/leaanthony/gosod v1.0.4 // indirect
	github.com/leaanthony/slicer v1.6.0 // indirect
	github.com/leaanthony/u v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/richardlehane/mscfb v1.0.7 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/samber/lo v1.49.1 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/tkrajina/go-reflector v0.5.8 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/wailsapp/go-webview2 v1.0.22 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.39.0 // indirect
	modernc.org/libc v1.77.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
)
=== internal/reader/uid.go ===
package reader

import "strings"

// NormalisasiUID membuang semua non-digit lalu mengisi nol di kiri sampai 10 digit.
// Hasil dengan 0 atau >10 digit dianggap tidak valid.
func NormalisasiUID(s string) (string, bool) {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	if len(d) < 1 || len(d) > 10 {
		return "", false
	}
	return strings.Repeat("0", 10-len(d)) + d, true
}
zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ 

Anda mengirim berkas lanjutan tanpa keberatan, jadi kelima keputusan di rencana saya jalankan sesuai usulan. Mode Ganti menonaktifkan kelas (bukan menghapus), logo tidak diikutkan, hanya 5 teks instansi dan jeda minimum yang disinkronkan, NISN pendek diberi peringatan, dan ekspor ditolak bila folder tujuan berada di dalam repo git.

Saya belum menjalankan kodenya. Kirim hasil `go test ./...` bila ada yang gagal, termasuk dari Fitur 3 yang hasilnya belum saya terima.

Tiga tambahan dari rencana awal:
- Tombol Terapkan hanya aktif bila pratinjau tanpa galat. Sidik jari pratinjau (berkas + rencana) diperiksa ulang di dalam transaksi, jadi berkas atau DB yang berubah sejak pratinjau membuat impor ditolak.
- Delimiter CSV `;` (ekspor Excel lokal Indonesia) dikenali otomatis.
- Selama dialog impor terbuka, tap kartu tidak memproses transaksi.

## 1. `internal/sinkron/sinkron.go` (berkas baru)

```go
// Package sinkron membaca dan menulis berkas impor/ekspor data master (xlsx/csv)
// tanpa menyentuh database.
package sinkron

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	SheetPengaturan = "Pengaturan"
	SheetKelas      = "Kelas"
	SheetKartu      = "Kartu"
	SheetPetunjuk   = "Petunjuk"
	maksBerkas      = 5 << 20
)

// KunciPengaturan: pengaturan yang disinkronkan (port serial dan ambang HID khusus perangkat, tidak ikut).
var KunciPengaturan = []string{
	"nama_instansi", "alamat_instansi", "penanggung_jawab",
	"nip_penanggung_jawab", "jabatan_penanggung_jawab", "jeda_minimum_menit",
}

var (
	headKelas = []string{"nama", "nama_ketua", "urutan", "aktif"}
	headKartu = []string{"uid", "kelas", "nisn", "nama_siswa", "aktif"}
)

var PetunjukTemplate = []string{
	"PETUNJUK IMPOR DATA MASTER MBG",
	"",
	"Sheet Kelas: nama (wajib, kunci pencocokan), nama_ketua (wajib), urutan (angka, bawaan 0), aktif (1/0, bawaan 1).",
	"Sheet Kartu: uid (wajib, angka 1-10 digit, kunci pencocokan), kelas (wajib, harus sama dengan nama kelas), nisn (opsional, 10 digit), nama_siswa (opsional), aktif (1/0, bawaan 1).",
	"Kolom sudah berformat TEKS agar angka nol di depan UID dan NISN tidak hilang. Jangan ubah formatnya menjadi angka.",
	"Mode Gabung hanya menambah dan memperbarui. Mode Ganti juga menghapus kartu dan menonaktifkan kelas yang tidak ada di berkas.",
	"Berkas berisi data pribadi siswa (NISN dan nama): jangan dimasukkan ke repositori atau dibagikan sembarangan.",
}

type KelasBaris struct {
	Baris                 int
	Nama, Ketua, Urutan, Aktif string
}

type KartuBaris struct {
	Baris                            int
	UID, Kelas, NISN, Nama, Aktif string
}

type Data struct {
	SHA        string            // sha256 isi berkas (hanya hasil Baca)
	Pengaturan map[string]string // nil = berkas tidak memuat pengaturan
	AdaKelas   bool
	AdaKartu   bool
	Kelas      []KelasBaris
	Kartu      []KartuBaris
}

// ---------- Baca ----------

func Baca(path string) (Data, error) {
	var d Data
	st, err := os.Stat(path)
	if err != nil {
		return d, err
	}
	if st.Size() > maksBerkas {
		return d, errors.New("Berkas terlalu besar (maksimal 5 MB)")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return d, err
	}
	sum := sha256.Sum256(raw)
	d.SHA = hex.EncodeToString(sum[:])
	switch strings.ToLower(filepath.Ext(path)) {
	case ".xlsx":
		err = bacaXLSX(raw, &d)
	case ".csv":
		err = bacaCSV(raw, &d)
	default:
		err = errors.New("Format berkas harus .xlsx atau .csv")
	}
	return d, err
}

func kosong(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func petakan(rows [][]string, sheet string, wajib []string) (map[string]int, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("Sheet %q kosong: baris header tidak ada", sheet)
	}
	idx := map[string]int{}
	for i, h := range rows[0] {
		h = strings.ToLower(strings.TrimSpace(h))
		if _, ada := idx[h]; h != "" && !ada {
			idx[h] = i
		}
	}
	for _, w := range wajib {
		if _, ok := idx[w]; !ok {
			return nil, fmt.Errorf("Kolom %q tidak ada pada %q", w, sheet)
		}
	}
	return idx, nil
}

func sel(row []string, idx map[string]int, nama string) string {
	if j, ok := idx[nama]; ok && j < len(row) {
		return strings.TrimSpace(row[j])
	}
	return ""
}

func parseKelas(rows [][]string) ([]KelasBaris, error) {
	idx, err := petakan(rows, SheetKelas, []string{"nama", "nama_ketua"})
	if err != nil {
		return nil, err
	}
	var out []KelasBaris
	for i := 1; i < len(rows); i++ {
		if kosong(rows[i]) {
			continue
		}
		r := rows[i]
		out = append(out, KelasBaris{Baris: i + 1, Nama: sel(r, idx, "nama"), Ketua: sel(r, idx, "nama_ketua"),
			Urutan: sel(r, idx, "urutan"), Aktif: sel(r, idx, "aktif")})
	}
	return out, nil
}

func parseKartu(rows [][]string) ([]KartuBaris, error) {
	idx, err := petakan(rows, SheetKartu, []string{"uid", "kelas"})
	if err != nil {
		return nil, err
	}
	var out []KartuBaris
	for i := 1; i < len(rows); i++ {
		if kosong(rows[i]) {
			continue
		}
		r := rows[i]
		out = append(out, KartuBaris{Baris: i + 1, UID: sel(r, idx, "uid"), Kelas: sel(r, idx, "kelas"),
			NISN: sel(r, idx, "nisn"), Nama: sel(r, idx, "nama_siswa"), Aktif: sel(r, idx, "aktif")})
	}
	return out, nil
}

func parsePengaturan(rows [][]string) (map[string]string, error) {
	idx, err := petakan(rows, SheetPengaturan, []string{"kunci", "nilai"})
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	for i := 1; i < len(rows); i++ {
		if kosong(rows[i]) {
			continue
		}
		k := strings.ToLower(sel(rows[i], idx, "kunci"))
		if k != "" {
			m[k] = sel(rows[i], idx, "nilai")
		}
	}
	return m, nil
}

func bacaXLSX(raw []byte, d *Data) error {
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return errors.New("Berkas Excel tidak dapat dibaca: " + err.Error())
	}
	defer f.Close()
	for _, nama := range f.GetSheetList() {
		rows, err := f.GetRows(nama)
		if err != nil {
			return err
		}
		switch strings.ToLower(strings.TrimSpace(nama)) {
		case "pengaturan":
			m, err := parsePengaturan(rows)
			if err != nil {
				return err
			}
			d.Pengaturan = m
		case "kelas":
			k, err := parseKelas(rows)
			if err != nil {
				return err
			}
			d.AdaKelas, d.Kelas = true, k
		case "kartu":
			k, err := parseKartu(rows)
			if err != nil {
				return err
			}
			d.AdaKartu, d.Kartu = true, k
		}
	}
	if d.Pengaturan == nil && !d.AdaKelas && !d.AdaKartu {
		return errors.New("Berkas tidak memuat sheet Pengaturan, Kelas, maupun Kartu")
	}
	return nil
}

func bacaCSV(raw []byte, d *Data) error {
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	baris1 := raw
	if i := bytes.IndexByte(raw, '\n'); i >= 0 {
		baris1 = raw[:i]
	}
	koma := ','
	if bytes.Count(baris1, []byte(";")) > bytes.Count(baris1, []byte(",")) {
		koma = ';'
	}
	r := csv.NewReader(bytes.NewReader(raw))
	r.Comma = koma
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	rows, err := r.ReadAll()
	if err != nil {
		return errors.New("Berkas CSV tidak dapat dibaca: " + err.Error())
	}
	if len(rows) == 0 {
		return errors.New("Berkas CSV kosong")
	}
	head := map[string]bool{}
	for _, h := range rows[0] {
		head[strings.ToLower(strings.TrimSpace(h))] = true
	}
	switch {
	case head["uid"]:
		k, err := parseKartu(rows)
		if err != nil {
			return err
		}
		d.AdaKartu, d.Kartu = true, k
	case head["nama_ketua"]:
		k, err := parseKelas(rows)
		if err != nil {
			return err
		}
		d.AdaKelas, d.Kelas = true, k
	default:
		return errors.New("Header CSV tidak dikenali: gunakan kolom kelas (nama, nama_ketua, urutan, aktif) atau kartu (uid, kelas, nisn, nama_siswa, aktif)")
	}
	return nil
}

// ---------- Tulis ----------

func (d Data) barisKelas() [][]string {
	out := make([][]string, 0, len(d.Kelas))
	for _, k := range d.Kelas {
		out = append(out, []string{k.Nama, k.Ketua, k.Urutan, k.Aktif})
	}
	return out
}

func (d Data) barisKartu() [][]string {
	out := make([][]string, 0, len(d.Kartu))
	for _, k := range d.Kartu {
		out = append(out, []string{k.UID, k.Kelas, k.NISN, k.Nama, k.Aktif})
	}
	return out
}

// TulisXLSX: semua sel ditulis sebagai teks berformat "@" agar nol di depan UID/NISN tidak hilang.
// petunjuk (opsional) ditulis sebagai sheet pertama.
func TulisXLSX(path string, d Data, petunjuk []string) error {
	f := excelize.NewFile()
	defer f.Close()
	teks, err := f.NewStyle(&excelize.Style{NumFmt: 49})
	if err != nil {
		return err
	}
	tebal, err := f.NewStyle(&excelize.Style{NumFmt: 49, Font: &excelize.Font{Bold: true}})
	if err != nil {
		return err
	}
	pertama := true
	baru := func(nama string) error {
		if pertama {
			pertama = false
			return f.SetSheetName("Sheet1", nama)
		}
		_, err := f.NewSheet(nama)
		return err
	}
	if len(petunjuk) > 0 {
		if err := baru(SheetPetunjuk); err != nil {
			return err
		}
		_ = f.SetColWidth(SheetPetunjuk, "A", "A", 120)
		for i, ln := range petunjuk {
			cell, _ := excelize.CoordinatesToCellName(1, i+1)
			if err := f.SetCellStr(SheetPetunjuk, cell, ln); err != nil {
				return err
			}
		}
	}
	tabel := func(nama string, head []string, rows [][]string, lebar []float64) error {
		if err := baru(nama); err != nil {
			return err
		}
		akhir, _ := excelize.ColumnNumberToName(len(head))
		if err := f.SetColStyle(nama, "A:"+akhir, teks); err != nil {
			return err
		}
		for c, h := range head {
			cell, _ := excelize.CoordinatesToCellName(c+1, 1)
			if err := f.SetCellStr(nama, cell, h); err != nil {
				return err
			}
			if err := f.SetCellStyle(nama, cell, cell, tebal); err != nil {
				return err
			}
		}
		for r, row := range rows {
			for c, v := range row {
				cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
				if err := f.SetCellStr(nama, cell, v); err != nil {
					return err
				}
			}
		}
		for c, w := range lebar {
			col, _ := excelize.ColumnNumberToName(c + 1)
			_ = f.SetColWidth(nama, col, col, w)
		}
		return nil
	}
	if d.Pengaturan != nil {
		var rows [][]string
		for _, k := range KunciPengaturan {
			rows = append(rows, []string{k, d.Pengaturan[k]})
		}
		if err := tabel(SheetPengaturan, []string{"kunci", "nilai"}, rows, []float64{30, 50}); err != nil {
			return err
		}
	}
	if d.AdaKelas {
		if err := tabel(SheetKelas, headKelas, d.barisKelas(), []float64{18, 30, 10, 8}); err != nil {
			return err
		}
	}
	if d.AdaKartu {
		if err := tabel(SheetKartu, headKartu, d.barisKartu(), []float64{16, 18, 16, 32, 8}); err != nil {
			return err
		}
	}
	if pertama {
		return errors.New("Tidak ada data untuk ditulis")
	}
	if err := f.SaveAs(path); err != nil {
		return err
	}
	_ = os.Chmod(path, 0o600)
	return nil
}

// tulisCSV: UTF-8 dengan BOM agar terbuka benar di Excel.
func tulisCSV(path string, head []string, rows [][]string) error {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	if err := w.Write(head); err != nil {
		return err
	}
	if err := w.WriteAll(rows); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o600)
}

func TulisCSVKelas(path string, d Data) error { return tulisCSV(path, headKelas, d.barisKelas()) }
func TulisCSVKartu(path string, d Data) error { return tulisCSV(path, headKartu, d.barisKartu()) }

// TemplateXLSX: petunjuk + sheet Kelas dan Kartu hanya berisi header (tanpa baris contoh,
// supaya template yang diimpor apa adanya tidak menambah data palsu).
func TemplateXLSX(path string) error {
	return TulisXLSX(path, Data{AdaKelas: true, AdaKartu: true}, PetunjukTemplate)
}

func TemplateCSV(path, jenis string) error {
	switch jenis {
	case "kelas":
		return tulisCSV(path, headKelas, nil)
	case "kartu":
		return tulisCSV(path, headKartu, nil)
	}
	return errors.New("Jenis template harus kelas atau kartu")
}
```

## 2. `internal/model/model.go`: tambahkan di akhir berkas

```go
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
```

## 3. `internal/service/sinkron.go` (berkas baru)

```go
package service

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"mbg/internal/model"
	"mbg/internal/reader"
	"mbg/internal/sinkron"
	"mbg/internal/timeutil"
)

// pembaca: dipenuhi *sql.DB dan *sql.Tx, supaya perencanaan impor bisa dijalankan
// di luar transaksi (pratinjau) maupun di dalamnya (terapkan).
type pembaca interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type kelasLama struct {
	id     int64
	ketua  string
	urutan int
	aktif  bool
}

type kartuLama struct {
	kelas, nisn, nama string
	aktif             bool
}

type kelasRencana struct {
	baris  int
	nama   string
	ketua  string
	urutan int
	aktif  bool
	aksi   string // BARU | UBAH | SAMA | OTOMATIS | NONAKTIF
	lama   *kelasLama
}

type kartuRencana struct {
	baris                  int
	uid, kelas, nisn, nama string
	aktif                  bool
	aksi                   string // BARU | UBAH | SAMA | HAPUS
	lama                   *kartuLama
}

type rencanaImpor struct {
	kelas          []kelasRencana
	kartu          []kartuRencana
	pengaturan     map[string]string // hanya kunci yang berubah
	pengaturanLama model.Pengaturan
	baris          []model.BarisPratinjau
	r              model.Ring

Saya lanjutkan dari titik terpotong. Bagian 1 (`internal/sinkron/sinkron.go`) dan bagian 2 (tambahan `model.go`) di atas sudah lengkap. `internal/service/sinkron.go` terpotong di tengah, jadi saya kirim ulang utuh di bawah ini.

## 3. `internal/service/sinkron.go` (berkas baru, utuh)

```go
package service

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"mbg/internal/model"
	"mbg/internal/reader"
	"mbg/internal/sinkron"
	"mbg/internal/timeutil"
)

// pembaca: dipenuhi *sql.DB dan *sql.Tx, supaya perencanaan impor bisa dijalankan
// di luar transaksi (pratinjau) maupun di dalamnya (terapkan).
type pembaca interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type kelasLama struct {
	id     int64
	ketua  string
	urutan int
	aktif  bool
}

type kartuLama struct {
	kelas, nisn, nama string
	aktif             bool
}

type kelasRencana struct {
	baris  int
	nama   string
	ketua  string
	urutan int
	aktif  bool
	aksi   string // BARU | UBAH | SAMA | OTOMATIS | NONAKTIF
	lama   *kelasLama
}

type kartuRencana struct {
	baris                  int
	uid, kelas, nisn, nama string
	aktif                  bool
	aksi                   string // BARU | UBAH | SAMA | HAPUS
	lama                   *kartuLama
}

type rencanaImpor struct {
	kelas          []kelasRencana
	kartu          []kartuRencana
	pengaturan     map[string]string // hanya kunci yang berubah
	pengaturanLama model.Pengaturan
	baris          []model.BarisPratinjau
	r              model.RingkasanImpor
}

func (rc *rencanaImpor) perubahan() int {
	r := rc.r
	return r.KelasBaru + r.KelasUbah + r.KelasNonaktif + r.KelasOtomatis +
		r.KartuBaru + r.KartuUbah + r.KartuHapus + r.PengaturanUbah
}

// hitungHash: sidik jari berkas + opsi + rencana. Dicek ulang di dalam transaksi saat menerapkan,
// sehingga berkas atau database yang berubah sejak pratinjau membatalkan impor.
func (rc *rencanaImpor) hitungHash(sha string, o model.OpsiImpor) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%t|%+v\n", sha, o.Mode, o.BuatKelasOtomatis, rc.r)
	for _, b := range rc.baris {
		fmt.Fprintf(h, "%s|%d|%s|%s|%s\n", b.Entitas, b.Baris, b.Kunci, b.Aksi, b.Pesan)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (rc *rencanaImpor) hasil(path, hash string, o model.OpsiImpor) model.HasilPratinjau {
	baris := append([]model.BarisPratinjau{}, rc.baris...)
	rank := func(a string) int {
		switch a {
		case "GALAT":
			return 0
		case "PERINGATAN":
			return 1
		}
		return 2
	}
	sort.SliceStable(baris, func(i, j int) bool { return rank(baris[i].Aksi) < rank(baris[j].Aksi) })
	out := model.HasilPratinjau{Berkas: filepath.Base(path), Mode: o.Mode, Ringkasan: rc.r, Hash: hash}
	if len(baris) > 500 {
		baris = baris[:500]
		out.Terpotong = true
	}
	out.Baris = baris
	return out
}

// ---------- pembantu ----------

func semuaAngka(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func parseAktif(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "1", "ya", "y", "true", "aktif":
		return true, true
	case "0", "tidak", "t", "n", "false", "nonaktif":
		return false, true
	}
	return false, false
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func normalMode(o *model.OpsiImpor) error {
	switch strings.ToUpper(strings.TrimSpace(o.Mode)) {
	case "", "GABUNG":
		o.Mode = "GABUNG"
	case "GANTI":
		o.Mode = "GANTI"
	default:
		return errors.New("Mode impor tidak dikenal")
	}
	return nil
}

func bacaPengaturan(q pembaca) (model.Pengaturan, error) {
	var p model.Pengaturan
	err := q.QueryRow(`SELECT COALESCE(nama_instansi,''),COALESCE(alamat_instansi,''),
		COALESCE(penanggung_jawab,''),COALESCE(nip_penanggung_jawab,''),COALESCE(jabatan_penanggung_jawab,''),
		jeda_minimum_menit,hid_ambang_ms,COALESCE(port_serial,'') FROM pengaturan WHERE id=1`).
		Scan(&p.NamaInstansi, &p.AlamatInstansi, &p.PenanggungJawab, &p.NIPPenanggungJawab,
			&p.JabatanPenanggungJawab, &p.JedaMinimumMenit, &p.HIDAmbangMs, &p.PortSerial)
	return p, err
}

func muatKelasLama(q pembaca) (map[string]*kelasLama, error) {
	rows, err := q.Query(`SELECT id,nama,nama_ketua,urutan,aktif FROM kelas`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]*kelasLama{}
	for rows.Next() {
		var id int64
		var nama, ketua string
		var urutan, aktif int
		if err := rows.Scan(&id, &nama, &ketua, &urutan, &aktif); err != nil {
			return nil, err
		}
		out[nama] = &kelasLama{id: id, ketua: ketua, urutan: urutan, aktif: aktif == 1}
	}
	return out, rows.Err()
}

func muatKartuLama(q pembaca) (map[string]*kartuLama, error) {
	rows, err := q.Query(`SELECT c.uid,k.nama,COALESCE(c.nisn,''),COALESCE(c.nama_siswa,''),c.aktif
		FROM kartu c JOIN kelas k ON k.id=c.kelas_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]*kartuLama{}
	for rows.Next() {
		var uid, kelas, nisn, nama string
		var aktif int
		if err := rows.Scan(&uid, &kelas, &nisn, &nama, &aktif); err != nil {
			return nil, err
		}
		out[uid] = &kartuLama{kelas: kelas, nisn: nisn, nama: nama, aktif: aktif == 1}
	}
	return out, rows.Err()
}

func kunciPengaturanDikenal(k string) bool {
	for _, x := range sinkron.KunciPengaturan {
		if x == k {
			return true
		}
	}
	return false
}

// ---------- perencanaan ----------

// rencanakan: menghitung apa yang akan terjadi tanpa menulis apa pun.
func (s *Service) rencanakan(q pembaca, d sinkron.Data, o model.OpsiImpor) (*rencanaImpor, error) {
	rc := &rencanaImpor{}
	catat := func(ent string, baris int, kunci, aksi, pesan string) {
		rc.baris = append(rc.baris, model.BarisPratinjau{Entitas: ent, Baris: baris, Kunci: kunci, Aksi: aksi, Pesan: pesan})
		switch aksi {
		case "GALAT":
			rc.r.Galat++
		case "PERINGATAN":
			rc.r.Peringatan++
		}
	}

	kelasDB, err := muatKelasLama(q)
	if err != nil {
		return nil, err
	}
	kartuDB, err := muatKartuLama(q)
	if err != nil {
		return nil, err
	}

	// --- pengaturan ---
	if d.Pengaturan != nil {
		lama, err := bacaPengaturan(q)
		if err != nil {
			return nil, err
		}
		rc.pengaturanLama = lama
		cur := map[string]string{
			"nama_instansi": lama.NamaInstansi, "alamat_instansi": lama.AlamatInstansi,
			"penanggung_jawab": lama.PenanggungJawab, "nip_penanggung_jawab": lama.NIPPenanggungJawab,
			"jabatan_penanggung_jawab": lama.JabatanPenanggungJawab,
			"jeda_minimum_menit":       strconv.Itoa(lama.JedaMinimumMenit),
		}
		rc.pengaturan = map[string]string{}
		for _, k := range sinkron.KunciPengaturan {
			v, ada := d.Pengaturan[k]
			if !ada {
				continue
			}
			v = strings.TrimSpace(v)
			if k == "jeda_minimum_menit" {
				n, err := strconv.Atoi(v)
				if err != nil || n < 0 || n > 600 {
					catat("pengaturan", 0, k, "GALAT", "Jeda minimum harus bilangan 0-600")
					continue
				}
				v = strconv.Itoa(n)
			}
			if v != cur[k] {
				rc.pengaturan[k] = v
				rc.r.PengaturanUbah++
				catat("pengaturan", 0, k, "UBAH", "Nilai diperbarui")
			}
		}
		var asing []string
		for k := range d.Pengaturan {
			if !kunciPengaturanDikenal(k) {
				asing = append(asing, k)
			}
		}
		sort.Strings(asing)
		for _, k := range asing {
			catat("pengaturan", 0, k, "PERINGATAN", "Kunci tidak dikenal, diabaikan")
		}
	}

	// --- kelas ---
	dalamBerkas := map[string]bool{}
	otomatis := map[string]bool{}
	dipakai := map[string]bool{}
	if d.AdaKelas {
		if o.Mode == "GANTI" && len(d.Kelas) == 0 {
			catat("kelas", 0, "-", "PERINGATAN", "Berkas tanpa baris kelas: mode Ganti akan menonaktifkan semua kelas")
		}
		for _, b := range d.Kelas {
			nama := strings.TrimSpace(b.Nama)
			if nama == "" {
				catat("kelas", b.Baris, "-", "GALAT", "Nama kelas kosong")
				continue
			}
			if dalamBerkas[nama] {
				catat("kelas", b.Baris, nama, "GALAT", "Nama kelas ganda dalam berkas")
				continue
			}
			dalamBerkas[nama] = true
			ketua := strings.TrimSpace(b.Ketua)
			if ketua == "" {
				catat("kelas", b.Baris, nama, "GALAT", "Nama ketua kosong")
				continue
			}
			urutan := 0
			if u := strings.TrimSpace(b.Urutan); u != "" {
				n, err := strconv.Atoi(u)
				if err != nil {
					catat("kelas", b.Baris, nama, "GALAT", "Urutan harus bilangan bulat")
					continue
				}
				urutan = n
			}
			aktif, ok := parseAktif(b.Aktif)
			if !ok {
				catat("kelas", b.Baris, nama, "GALAT", "Kolom aktif harus 1 atau 0")
				continue
			}
			kr := kelasRencana{baris: b.Baris, nama: nama, ketua: ketua, urutan: urutan, aktif: aktif}
			if lama, ada := kelasDB[nama]; ada {
				kr.lama = lama
				if lama.ketua == ketua && lama.urutan == urutan && lama.aktif == aktif {
					kr.aksi = "SAMA"
					rc.r.KelasSama++
				} else {
					kr.aksi = "UBAH"
					rc.r.KelasUbah++
					catat("kelas", b.Baris, nama, "UBAH", "Data kelas diperbarui")
				}
			} else {
				kr.aksi = "BARU"
				rc.r.KelasBaru++
				catat("kelas", b.Baris, nama, "BARU", "Kelas baru")
			}
			rc.kelas = append(rc.kelas, kr)
		}
	}
	kelasKenal := func(n string) bool { return kelasDB[n] != nil || dalamBerkas[n] || otomatis[n] }

	// --- kartu ---
	uidBerkas := map[string]bool{}
	if d.AdaKartu {
		if o.Mode == "GANTI" && len(d.Kartu) == 0 {
			catat("kartu", 0, "-", "PERINGATAN", "Berkas tanpa baris kartu: mode Ganti akan menghapus semua kartu")
		}
		for _, b := range d.Kartu {
			raw := strings.TrimSpace(b.UID)
			if !semuaAngka(raw) || len(raw) > 10 {
				catat("kartu", b.Baris, raw, "GALAT", "UID bukan angka 1-10 digit")
				continue
			}
			uid, _ := reader.NormalisasiUID(raw)
			if uidBerkas[uid] {
				catat("kartu", b.Baris, uid, "GALAT", "UID ganda dalam berkas")
				continue
			}
			uidBerkas[uid] = true
			kelas := strings.TrimSpace(b.Kelas)
			if kelas == "" {
				catat("kartu", b.Baris, uid, "GALAT", "Kelas kosong")
				continue
			}
			nisn := strings.TrimSpace(b.NISN)
			if nisn != "" && (!semuaAngka(nisn) || len(nisn) > 10) {
				catat("kartu", b.Baris, uid, "GALAT", "NISN harus angka maksimal 10 digit")
				continue
			}
			aktif, ok := parseAktif(b.Aktif)
			if !ok {
				catat("kartu", b.Baris, uid, "GALAT", "Kolom aktif harus 1 atau 0")
				continue
			}
			if !kelasKenal(kelas) {
				if !o.BuatKelasOtomatis {
					catat("kartu", b.Baris, uid, "GALAT", fmt.Sprintf("Kelas %q tidak dikenal", kelas))
					continue
				}
				otomatis[kelas] = true
				rc.kelas = append(rc.kelas, kelasRencana{baris: b.Baris, nama: kelas, ketua: "-", aktif: true, aksi: "OTOMATIS"})
				rc.r.KelasOtomatis++
				catat("kelas", b.Baris, kelas, "OTOMATIS", `Dibuat otomatis (ketua "-")`)
			}
			if nisn != "" && len(nisn) < 10 {
				catat("kartu", b.Baris, uid, "PERINGATAN", "NISN kurang dari 10 digit (nol di depan mungkin hilang)")
			}
			dipakai[kelas] = true
			kr := kartuRencana{baris: b.Baris, uid: uid, kelas: kelas, nisn: nisn, nama: strings.TrimSpace(b.Nama), aktif: aktif}
			if lama, ada := kartuDB[uid]; ada {
				kr.lama = lama
				if lama.kelas == kelas && lama.nisn == kr.nisn && lama.nama == kr.nama && lama.aktif == aktif {
					kr.aksi = "SAMA"
					rc.r.KartuSama++
				} else {
					kr.aksi = "UBAH"
					rc.r.KartuUbah++
					pesan := "Data kartu diperbarui"
					if lama.kelas != kelas {
						pesan = fmt.Sprintf("Pindah kelas %s → %s", lama.kelas, kelas)
					}
					catat("kartu", b.Baris, uid, "UBAH", pesan)
				}
			} else {
				kr.aksi = "BARU"
				rc.r.KartuBaru++
				catat("kartu", b.Baris, uid, "BARU", "Kartu baru, kelas "+kelas)
			}
			rc.kartu = append(rc.kartu, kr)
		}
	}

	// --- mode Ganti ---
	if o.Mode == "GANTI" {
		if d.AdaKartu {
			var uids []string
			for uid := range kartuDB {
				if !uidBerkas[uid] {
					uids = append(uids, uid)
				}
			}
			sort.Strings(uids)
			for _, uid := range uids {
				rc.kartu = append(rc.kartu, kartuRencana{uid: uid, kelas: kartuDB[uid].kelas, aksi: "HAPUS", lama: kartuDB[uid]})
				rc.r.KartuHapus++
				catat("kartu", 0, uid, "HAPUS", "Tidak ada di berkas: kartu dihapus")
			}
		}
		if d.AdaKelas {
			var nama []string
			for n, k := range kelasDB {
				if k.aktif && !dalamBerkas[n] && !dipakai[n] {
					nama = append(nama, n)
				}
			}
			sort.Strings(nama)
			for _, n := range nama {
				rc.kelas = append(rc.kelas, kelasRencana{nama: n, aksi: "NONAKTIF", lama: kelasDB[n]})
				rc.r.KelasNonaktif++
				catat("kelas", 0, n, "NONAKTIF", "Tidak ada di berkas: kelas dinonaktifkan")
			}
		}
	}
	return rc, nil
}

// ---------- pratinjau dan terapkan ----------

// PilihBerkasImpor: dialog pilih berkas .xlsx/.csv. Mengembalikan "" bila dibatalkan.
func (s *Service) PilihBerkasImpor() (string, error) {
	if s.app.ctx == nil {
		return "", errors.New("Dialog file belum siap")
	}
	opts := wruntime.OpenDialogOptions{
		Title:   "Pilih berkas impor master",
		Filters: []wruntime.FileFilter{{DisplayName: "Excel/CSV (*.xlsx;*.csv)", Pattern: "*.xlsx;*.csv"}},
	}
	if d := folderAwal(); d != "" {
		opts.DefaultDirectory = d
	}
	p, err := wruntime.OpenFileDialog(s.app.ctx, opts)
	if err != nil {
		return "", err
	}
	return p, nil
}

// PratinjauImpor: tidak menulis apa pun ke database.
func (s *Service) PratinjauImpor(path string, o model.OpsiImpor) (model.HasilPratinjau, error) {
	var out model.HasilPratinjau
	if err := normalMode(&o); err != nil {
		return out, err
	}
	d, err := sinkron.Baca(path)
	if err != nil {
		return out, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rc, err := s.rencanakan(s.db, d, o)
	if err != nil {
		return out, err
	}
	return rc.hasil(path, rc.hitungHash(d.SHA, o), o), nil
}

// TerapkanImpor: seluruh impor dalam satu transaksi; satu galat berarti tidak ada yang berubah.
// o.Hash harus berasal dari PratinjauImpor dengan berkas dan opsi yang sama.
func (s *Service) TerapkanImpor(path string, o model.OpsiImpor) (model.HasilImpor, error) {
	var out model.HasilImpor
	if err := normalMode(&o); err != nil {
		return out, err
	}
	o.Alasan = strings.TrimSpace(o.Alasan)
	if o.Mode == "GANTI" {
		if !o.KonfirmasiGanti {
			return out, errors.New("Mode Ganti memerlukan konfirmasi eksplisit")
		}
		if o.Alasan == "" {
			return out, errors.New("Alasan wajib diisi untuk mode Ganti")
		}
	}
	d, err := sinkron.Baca(path)
	if err != nil {
		return out, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		return out, err
	}
	defer tx.Rollback()

	rc, err := s.rencanakan(tx, d, o)
	if err != nil {
		return out, err
	}
	if rc.r.Galat > 0 {
		return out, fmt.Errorf("Impor dibatalkan: %d galat pada berkas, tidak ada data yang diubah", rc.r.Galat)
	}
	if o.Hash == "" || o.Hash != rc.hitungHash(d.SHA, o) {
		return out, errors.New("Berkas atau data berubah sejak pratinjau; ulangi pratinjau")
	}
	out.Ringkasan = rc.r
	if rc.perubahan() == 0 {
		out.Pesan = "Tidak ada perubahan: data sudah sama dengan berkas"
		return out, nil
	}
	alasan := o.Alasan
	if alasan == "" {
		alasan = "impor master"
	}

	// pengaturan
	if len(rc.pengaturan) > 0 {
		keys := make([]string, 0, len(rc.pengaturan))
		for k := range rc.pengaturan {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys { // k berasal dari daftar putih KunciPengaturan
			var arg any = rc.pengaturan[k]
			if k == "jeda_minimum_menit" {
				n, _ := strconv.Atoi(rc.pengaturan[k])
				arg = n
			}
			if _, err := tx.Exec(`UPDATE pengaturan SET `+k+`=? WHERE id=1`, arg); err != nil {
				return out, err
			}
		}
		if err := s.audit(tx, now, "UBAH", "pengaturan", "1", rc.pengaturanLama, rc.pengaturan, alasan); err != nil {
			return out, err
		}
	}

	// kelas
	for _, k := range rc.kelas {
		switch k.aksi {
		case "BARU", "OTOMATIS":
			res, err := tx.Exec(`INSERT INTO kelas (nama,nama_ketua,urutan,aktif) VALUES (?,?,?,?)`,
				k.nama, k.ketua, k.urutan, boolInt(k.aktif))
			if err != nil {
				return out, err
			}
			id, _ := res.LastInsertId()
			if err := s.audit(tx, now, "TAMBAH", "kelas", itoa(id), nil,
				map[string]any{"nama": k.nama, "ketua": k.ketua, "urutan": k.urutan, "aktif": k.aktif}, alasan); err != nil {
				return out, err
			}
		case "UBAH":
			if _, err := tx.Exec(`UPDATE kelas SET nama_ketua=?,urutan=?,aktif=? WHERE id=?`,
				k.ketua, k.urutan, boolInt(k.aktif), k.lama.id); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "UBAH", "kelas", itoa(k.lama.id),
				map[string]any{"nama": k.nama, "ketua": k.lama.ketua, "urutan": k.lama.urutan, "aktif": k.lama.aktif},
				map[string]any{"nama": k.nama, "ketua": k.ketua, "urutan": k.urutan, "aktif": k.aktif}, alasan); err != nil {
				return out, err
			}
		case "NONAKTIF":
			if _, err := tx.Exec(`UPDATE kelas SET aktif=0 WHERE id=?`, k.lama.id); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "UBAH", "kelas", itoa(k.lama.id),
				map[string]any{"aktif": 1}, map[string]any{"aktif": 0}, alasan); err != nil {
				return out, err
			}
		}
	}

	// kartu
	idKelas := func(nama string) (int64, error) {
		var id int64
		if err := tx.QueryRow(`SELECT id FROM kelas WHERE nama=?`, nama).Scan(&id); err != nil {
			return 0, fmt.Errorf("Kelas %q tidak ditemukan", nama)
		}
		return id, nil
	}
	for _, c := range rc.kartu {
		switch c.aksi {
		case "BARU":
			kid, err := idKelas(c.kelas)
			if err != nil {
				return out, err
			}
			if _, err := tx.Exec(`INSERT INTO kartu (uid,kelas_id,aktif,nisn,nama_siswa) VALUES (?,?,?,?,?)`,
				c.uid, kid, boolInt(c.aktif), c.nisn, c.nama); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "TAMBAH", "kartu", c.uid, nil,
				map[string]any{"kelas": c.kelas, "nisn": c.nisn, "nama": c.nama, "aktif": c.aktif}, alasan); err != nil {
				return out, err
			}
		case "UBAH":
			kid, err := idKelas(c.kelas)
			if err != nil {
				return out, err
			}
			if _, err := tx.Exec(`UPDATE kartu SET kelas_id=?,aktif=?,nisn=?,nama_siswa=? WHERE uid=?`,
				kid, boolInt(c.aktif), c.nisn, c.nama, c.uid); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "UBAH", "kartu", c.uid,
				map[string]any{"kelas": c.lama.kelas, "nisn": c.lama.nisn, "nama": c.lama.nama, "aktif": c.lama.aktif},
				map[string]any{"kelas": c.kelas, "nisn": c.nisn, "nama": c.nama, "aktif": c.aktif}, alasan); err != nil {
				return out, err
			}
		case "HAPUS":
			if _, err := tx.Exec(`DELETE FROM kartu WHERE uid=?`, c.uid); err != nil {
				return out, err
			}
			if err := s.audit(tx, now, "HAPUS", "kartu", c.uid,
				map[string]any{"kelas": c.lama.kelas, "nisn": c.lama.nisn, "nama": c.lama.nama, "aktif": c.lama.aktif}, nil, alasan); err != nil {
				return out, err
			}
		}
	}

	if err := s.audit(tx, now, "IMPOR_MASTER", "master", "", nil,
		map[string]any{"mode": o.Mode, "berkas": filepath.Base(path), "ringkasan": rc.r}, alasan); err != nil {
		return out, err
	}
	if err := tx.Commit(); err != nil {
		return out, err
	}
	r := rc.r
	out.Pesan = fmt.Sprintf("Impor selesai: kelas %d baru, %d diubah, %d dinonaktifkan; kartu %d baru, %d diubah, %d dihapus; pengaturan %d diubah",
		r.KelasBaru+r.KelasOtomatis, r.KelasUbah, r.KelasNonaktif, r.KartuBaru, r.KartuUbah, r.KartuHapus, r.PengaturanUbah)
	return out, nil
}

// ---------- ekspor ----------

func (s *Service) dataMaster(denganPengaturan bool) (sinkron.Data, error) {
	d := sinkron.Data{AdaKelas: true, AdaKartu: true}
	if denganPengaturan {
		p, err := s.GetPengaturan()
		if err != nil {
			return d, err
		}
		d.Pengaturan = map[string]string{
			"nama_instansi": p.NamaInstansi, "alamat_instansi": p.AlamatInstansi,
			"penanggung_jawab": p.PenanggungJawab, "nip_penanggung_jawab": p.NIPPenanggungJawab,
			"jabatan_penanggung_jawab": p.JabatanPenanggungJawab,
			"jeda_minimum_menit":       strconv.Itoa(p.JedaMinimumMenit),
		}
	}
	kelas, err := s.db.Query(`SELECT nama,nama_ketua,urutan,aktif FROM kelas ORDER BY urutan,nama`)
	if err != nil {
		return d, err
	}
	for kelas.Next() {
		var nama, ketua string
		var urutan, aktif int
		if err := kelas.Scan(&nama, &ketua, &urutan, &aktif); err != nil {
			kelas.Close()
			return d, err
		}
		d.Kelas = append(d.Kelas, sinkron.KelasBaris{Nama: nama, Ketua: ketua, Urutan: strconv.Itoa(urutan), Aktif: strconv.Itoa(aktif)})
	}
	if err := kelas.Err(); err != nil {
		kelas.Close()
		return d, err
	}
	kelas.Close()

	kartu, err := s.db.Query(`SELECT c.uid,k.nama,COALESCE(c.nisn,''),COALESCE(c.nama_siswa,''),c.aktif
		FROM kartu c JOIN kelas k ON k.id=c.kelas_id ORDER BY k.urutan,k.nama,c.nama_siswa,c.uid`)
	if err != nil {
		return d, err
	}
	for kartu.Next() {
		var uid, nama, nisn, siswa string
		var aktif int
		if err := kartu.Scan(&uid, &nama, &nisn, &siswa, &aktif); err != nil {
			kartu.Close()
			return d, err
		}
		d.Kartu = append(d.Kartu, sinkron.KartuBaris{UID: uid, Kelas: nama, NISN: nisn, Nama: siswa, Aktif: strconv.Itoa(aktif)})
	}
	err = kartu.Err()
	kartu.Close()
	return d, err
}

func folderAwal() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, sub := range []string{"Documents", "Dokumen", "Downloads", "Unduhan"} {
		p := filepath.Join(h, sub)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return h
}

// tolakDalamRepo: berkas ekspor berisi NISN dan nama siswa, jangan sampai masuk repositori git.
func tolakDalamRepo(p string) error {
	dir, err := filepath.Abs(filepath.Dir(p))
	if err != nil {
		return nil
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return errors.New("Berkas berisi data pribadi siswa (NISN dan nama) dan tidak boleh disimpan di dalam repositori git. Pilih folder lain.")
		}
		induk := filepath.Dir(dir)
		if induk == dir {
			return nil
		}
		dir = induk
	}
}

func (s *Service) pilihSimpan(judul, nama, ext, label string, jagaRepo bool) (string, error) {
	if s.app.ctx == nil {
		return "", errors.New("Dialog file belum siap")
	}
	opts := wruntime.SaveDialogOptions{
		Title: judul, DefaultFilename: nama,
		Filters: []wruntime.FileFilter{{DisplayName: label, Pattern: "*." + ext}},
	}
	if d := folderAwal(); d != "" {
		opts.DefaultDirectory = d
	}
	p, err := wruntime.SaveFileDialog(s.app.ctx, opts)
	if err != nil || p == "" {
		return "", err
	}
	if !strings.HasSuffix(strings.ToLower(p), "."+ext) {
		p += "." + ext
	}
	if jagaRepo {
		if err := tolakDalamRepo(p); err != nil {
			return "", err
		}
	}
	return p, nil
}

// EksporMasterExcel: sheet Pengaturan, Kelas, Kartu. Mengembalikan path, atau "" bila dialog dibatalkan.
func (s *Service) EksporMasterExcel() (string, error) {
	d, err := s.dataMaster(true)
	if err != nil {
		return "", err
	}
	p, err := s.pilihSimpan("Simpan ekspor master", "master-mbg-"+timeutil.Tanggal(s.now())+".xlsx", "xlsx", "Excel (*.xlsx)", true)
	if err != nil || p == "" {
		return "", err
	}
	if err := sinkron.TulisXLSX(p, d, nil); err != nil {
		return "", err
	}
	return p, nil
}

// EksporMasterCSV: jenis "kelas" atau "kartu".
func (s *Service) EksporMasterCSV(jenis string) (string, error) {
	if jenis != "kelas" && jenis != "kartu" {
		return "", errors.New("Jenis data harus kelas atau kartu")
	}
	d, err := s.dataMaster(false)
	if err != nil {
		return "", err
	}
	p, err := s.pilihSimpan("Simpan ekspor "+jenis, "master-"+jenis+"-"+timeutil.Tanggal(s.now())+".csv", "csv", "CSV (*.csv)", true)
	if err != nil || p == "" {
		return "", err
	}
	if jenis == "kelas" {
		err = sinkron.TulisCSVKelas(p, d)
	} else {
		err = sinkron.TulisCSVKartu(p, d)
	}
	if err != nil {
		return "", err
	}
	return p, nil
}

func (s *Service) UnduhTemplateExcel() (string, error) {
	p, err := s.pilihSimpan("Simpan template impor", "template-impor-master.xlsx", "xlsx", "Excel (*.xlsx)", false)
	if err != nil || p == "" {
		return "", err
	}
	if err := sinkron.TemplateXLSX(p); err != nil {
		return "", err
	}
	return p, nil
}

func (s *Service) UnduhTemplateCSV(jenis string) (string, error) {
	if jenis != "kelas" && jenis != "kartu" {
		return "", errors.New("Jenis template harus kelas atau kartu")
	}
	p, err := s.pilihSimpan("Simpan template "+jenis, "template-"+jenis+".csv", "csv", "CSV (*.csv)", false)
	if err != nil || p == "" {
		return "", err
	}
	if err := sinkron.TemplateCSV(p, jenis); err != nil {
		return "", err
	}
	return p, nil
}
```

## 4. Frontend

**`frontend/index.html`**, tiga sisipan:

a. Di `.subnav` Master Data, setelah tombol `Kartu RFID`:

```html
                <button data-sub="sinkron">Sinkronisasi</button>
```

b. Tepat sebelum `</section>` penutup `tab-master` (setelah `</div>` milik `sub-kartu`):

```html
            <div id="sub-sinkron" class="sub">
                <div class="peringatan"><b>Data pribadi.</b> Berkas ekspor kartu berisi NISN dan nama siswa. Simpan di
                    lokasi aman, jangan di folder repositori atau rilis publik, dan hapus setelah dipakai.</div>
                <h3>Ekspor</h3>
                <div class="form-baris">
                    <button id="sk-xlsx" class="utama">Ekspor Master (Excel)</button>
                    <button id="sk-csv-kelas">Ekspor Kelas (CSV)</button>
                    <button id="sk-csv-kartu">Ekspor Kartu (CSV)</button>
                </div>
                <h3>Template impor kosong</h3>
                <div class="form-baris">
                    <button id="sk-tpl-xlsx">Template Excel</button>
                    <button id="sk-tpl-kelas">Template CSV Kelas</button>
                    <button id="sk-tpl-kartu">Template CSV Kartu</button>
                </div>
                <h3>Impor</h3>
                <div class="form-baris"><button id="sk-impor" class="utama">Impor Master...</button></div>
                <p class="kecil">Impor selalu menampilkan pratinjau dulu dan tidak menulis apa pun sebelum Anda menekan
                    Terapkan. CSV yang dibuka lalu disimpan ulang di Excel bisa kehilangan nol di depan; UID tetap
                    dinormalkan saat impor, tetapi NISN pendek akan diberi peringatan. Logo, port serial, dan ambang HID
                    tidak ikut disinkronkan.</p>
            </div>
```

c. Di filter audit `#ra-entitas`, tambahkan setelah `<option>hari</option>`:

```html
                        <option>master</option>
```

d. Di daftar skrip, setelah `js/ekspor.js` dan sebelum `js/main.js`:

```html
    <script src="js/sinkron.js"></script>
```

**`frontend/js/sinkron.js`** (berkas baru):

```js
const Sinkron = {
    path: "", hasil: null,

    async unduh(fn) {
        try { const p = await fn(); if (p) toast("Tersimpan: " + p, "ok"); }
        catch (e) { toast(String(e.message || e), "err"); }
    },

    async mulaiImpor() {
        try {
            const p = await API.PilihBerkasImpor();
            if (!p) return;
            this.path = p; this.hasil = null;
            this.tampilModal();
        } catch (e) { toast(String(e.message || e), "err"); }
    },

    tutup() { Dash.tutupModal(); pasangHandler(); },

    opsi() {
        return {
            mode: _r("im-mode").value,
            buat_kelas_otomatis: _r("im-otomatis").checked,
            konfirmasi_ganti: _r("im-konfirm").checked,
            alasan: _r("im-alasan").value,
            hash: this.hasil ? this.hasil.hash : "",
        };
    },

    // pilihan berubah -> pratinjau lama tidak berlaku lagi
    reset() {
        this.hasil = null;
        _r("im-hasil").innerHTML = ""; _r("im-galat").textContent = "";
        _r("im-ok").disabled = true;
        _r("im-ganti").classList.toggle("hidden", _r("im-mode").value !== "GANTI");
    },

    tampilModal() {
        Tap.handler = () => toast("Selesaikan impor dulu", "info"); // tap tidak boleh memproses transaksi
        Dash.bukaModal(`
          <h2>Impor data master</h2>
          <p class="kecil">Berkas: <b>${esc(this.path.split(/[\\/]/).pop())}</b></p>
          <label>Mode
            <select id="im-mode">
              <option value="GABUNG">Gabung (tambah dan perbarui, tidak menghapus)</option>
              <option value="GANTI">Ganti (hapus kartu / nonaktifkan kelas yang tidak ada di berkas)</option>
            </select>
          </label>
          <label class="cek"><input type="checkbox" id="im-otomatis"> Buat kelas otomatis bila kelas pada kartu belum ada</label>
          <div id="im-ganti" class="hidden">
            <div class="peringatan">Mode Ganti menghapus kartu dan menonaktifkan kelas yang tidak ada di berkas.</div>
            <label class="cek"><input type="checkbox" id="im-konfirm"> Saya paham dan ingin melanjutkan</label>
            <label>Alasan (wajib) <input id="im-alasan"></label>
          </div>
          <div id="im-hasil"></div>
          <div class="galat" id="im-galat"></div>
          <div class="aksi"><button id="im-batal">Batal</button><button id="im-lihat">Pratinjau</button><button id="im-ok" disabled>Terapkan</button></div>`);
        _r("im-mode").onchange = () => this.reset();
        _r("im-otomatis").onchange = () => this.reset();
        _r("im-batal").onclick = () => this.tutup();
        _r("im-lihat").onclick = () => this.lihat();
        _r("im-ok").onclick = () => this.terapkan();
    },

    async lihat() {
        _r("im-galat").textContent = "";
        const o = this.opsi(); o.hash = "";
        try {
            this.hasil = await API.PratinjauImpor(this.path, o);
            this.render();
        } catch (e) { this.hasil = null; _r("im-galat").textContent = String(e.message || e); }
    },

    render() {
        const h = this.hasil, r = h.ringkasan;
        const ubah = r.kelas_baru + r.kelas_ubah + r.kelas_nonaktif + r.kelas_otomatis +
            r.kartu_baru + r.kartu_ubah + r.kartu_hapus + r.pengaturan_ubah;
        const ringkas = [
            `Kelas: ${r.kelas_baru} baru, ${r.kelas_ubah} diubah, ${r.kelas_otomatis} dibuat otomatis, ${r.kelas_nonaktif} dinonaktifkan, ${r.kelas_sama} sama`,
            `Kartu: ${r.kartu_baru} baru, ${r.kartu_ubah} diubah, ${r.kartu_hapus} dihapus, ${r.kartu_sama} sama`,
            `Pengaturan: ${r.pengaturan_ubah} diubah`,
            `Galat: ${r.galat} · Peringatan: ${r.peringatan}`,
        ];
        let html = `<p>${ringkas.map(esc).join("<br>")}</p>`;
        if (r.galat === 0 && ubah === 0) html += `<p class="kecil">Tidak ada perubahan: data sudah sama dengan berkas.</p>`;
        if (h.baris.length) {
            html += `<div class="pratinjau-wadah"><table><thead><tr><th>Entitas</th><th>Baris</th><th>Kunci</th><th>Aksi</th><th>Keterangan</th></tr></thead><tbody>` +
                h.baris.map(b => `<tr class="${b.aksi === "GALAT" ? "merah" : ""}"><td>${esc(b.entitas)}</td><td>${b.baris || "-"}</td><td>${esc(b.kunci)}</td><td>${esc(b.aksi)}</td><td>${esc(b.pesan)}</td></tr>`).join("") +
                `</tbody></table></div>`;
            if (h.terpotong) html += `<p class="kecil">Hanya 500 baris pertama yang ditampilkan.</p>`;
        }
        _r("im-hasil").innerHTML = html;
        _r("im-ok").disabled = r.galat > 0 || ubah === 0;
    },

    async terapkan() {
        _r("im-galat").textContent = "";
        try {
            const h = await API.TerapkanImpor(this.path, this.opsi());
            this.tutup();
            toast(h.pesan, "ok");
            Master.muatKelas(); Dash.muat();
        } catch (e) { _r("im-galat").textContent = String(e.message || e); bunyi("err"); }
    },
};

document.addEventListener("DOMContentLoaded", () => {
    _r("sk-xlsx").onclick = () => Sinkron.unduh(() => API.EksporMasterExcel());
    _r("sk-csv-kelas").onclick = () => Sinkron.unduh(() => API.EksporMasterCSV("kelas"));
    _r("sk-csv-kartu").onclick = () => Sinkron.unduh(() => API.EksporMasterCSV("kartu"));
    _r("sk-tpl-xlsx").onclick = () => Sinkron.unduh(() => API.UnduhTemplateExcel());
    _r("sk-tpl-kelas").onclick = () => Sinkron.unduh(() => API.UnduhTemplateCSV("kelas"));
    _r("sk-tpl-kartu").onclick = () => Sinkron.unduh(() => API.UnduhTemplateCSV("kartu"));
    _r("sk-impor").onclick = () => Sinkron.mulaiImpor();
});
```

**`frontend/css/style.css`**: tambahkan di akhir berkas:

```css
/* Sinkronisasi */
.peringatan {
    padding: 10px 14px;
    margin: 12px 0;
    background: #fef3c7;
    border: 1px solid #f59e0b;
    border-radius: var(--radius);
    color: #78350f;
}

#sub-sinkron {
    max-width: 760px
}

#sub-sinkron h3 {
    margin: 22px 0 0;
    font-size: 15px
}

#modal label.cek {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 400
}

#modal label.cek input {
    display: inline-block;
    width: auto;
    margin: 0
}

.pratinjau-wadah {
    max-height: 260px;
    overflow: auto;
    margin: 10px 0
}

.pratinjau-wadah th,
.pratinjau-wadah td {
    padding: 6px 10px;
    font-size: 13px
}

#im-ok {
    color: #fff;
    background: var(--hijau);
    border-color: var(--hijau)
}

#im-ok:hover:not(:disabled) {
    background: var(--hijau-tua);
    border-color: var(--hijau-tua)
}
```

## 5. Tes: `internal/service/sinkron_test.go` (berkas baru)

```go
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mbg/internal/model"
	"mbg/internal/sinkron"
)

func isiMaster(t *testing.T, s *Service) {
	t.Helper()
	p, err := s.GetPengaturan()
	harusOK(t, err)
	p.NamaInstansi, p.AlamatInstansi = "SMPN 1 Contoh", "Jl. Mawar 1"
	p.PenanggungJawab, p.NIPPenanggungJawab, p.JabatanPenanggungJawab = "Dra. Ani", "0123456789", "Kepala Sekolah"
	p.JedaMinimumMenit = 15
	harusOK(t, s.SimpanPengaturan(p))
	a, err := s.TambahKelas("VII A", "Ani", 1)
	harusOK(t, err)
	b, err := s.TambahKelas("VII B", "Budi", 2)
	harusOK(t, err)
	harusOK(t, s.DaftarkanKartu("0012345678", a, "0123456789", "Siti"))
	harusOK(t, s.DaftarkanKartu("0000000042", a, "", ""))
	harusOK(t, s.DaftarkanKartu("9876543210", b, "1234567890", "Doni"))
	harusOK(t, s.NonaktifkanKelas(b, "uji"))
}

func dumpMaster(t *testing.T, s *Service) string {
	t.Helper()
	var sb strings.Builder
	p, err := s.GetPengaturan()
	harusOK(t, err)
	fmt.Fprintf(&sb, "P|%s|%s|%s|%s|%s|%d\n", p.NamaInstansi, p.AlamatInstansi, p.PenanggungJawab,
		p.NIPPenanggungJawab, p.JabatanPenanggungJawab, p.JedaMinimumMenit)
	rows, err := s.db.Query(`SELECT nama,nama_ketua,urutan,aktif FROM kelas ORDER BY nama`)
	harusOK(t, err)
	for rows.Next() {
		var n, k string
		var u, a int
		harusOK(t, rows.Scan(&n, &k, &u, &a))
		fmt.Fprintf(&sb, "K|%s|%s|%d|%d\n", n, k, u, a)
	}
	rows.Close()
	rows, err = s.db.Query(`SELECT c.uid,k.nama,COALESCE(c.nisn,''),COALESCE(c.nama_siswa,''),c.aktif
		FROM kartu c JOIN kelas k ON k.id=c.kelas_id ORDER BY c.uid`)
	harusOK(t, err)
	for rows.Next() {
		var uid, kelas, nisn, nama string
		var a int
		harusOK(t, rows.Scan(&uid, &kelas, &nisn, &nama, &a))
		fmt.Fprintf(&sb, "C|%s|%s|%s|%s|%d\n", uid, kelas, nisn, nama, a)
	}
	rows.Close()
	return sb.String()
}

func hitung(t *testing.T, s *Service, query string) int {
	t.Helper()
	var n int
	harusOK(t, s.db.QueryRow(query).Scan(&n))
	return n
}

func tulisCSVUji(t *testing.T, isi string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "uji.csv")
	harusOK(t, os.WriteFile(p, []byte("\xEF\xBB\xBF"+isi), 0o600))
	return p
}

// impor: pratinjau lalu terapkan dengan hash dari pratinjau (alur UI).
func impor(t *testing.T, s *Service, path string, o model.OpsiImpor) (model.HasilImpor, error) {
	t.Helper()
	pr, err := s.PratinjauImpor(path, o)
	harusOK(t, err)
	o.Hash = pr.Hash
	return s.TerapkanImpor(path, o)
}

func TestSinkronRoundTripExcel(t *testing.T) {
	asal := setupT5(t)
	isiMaster(t, asal)
	berkas := filepath.Join(t.TempDir(), "master.xlsx")
	d, err := asal.dataMaster(true)
	harusOK(t, err)
	harusOK(t, sinkron.TulisXLSX(berkas, d, nil))

	tujuan := setupT5(t)
	pr, err := tujuan.PratinjauImpor(berkas, model.OpsiImpor{})
	harusOK(t, err)
	r := pr.Ringkasan
	if r.Galat != 0 || r.Peringatan != 0 || r.KelasBaru != 2 || r.KartuBaru != 3 || r.PengaturanUbah != 6 {
		t.Fatalf("pratinjau: %+v", r)
	}
	h, err := impor(t, tujuan, berkas, model.OpsiImpor{})
	harusOK(t, err)
	if h.Ringkasan.KartuBaru != 3 {
		t.Fatalf("hasil: %+v", h)
	}
	got, want := dumpMaster(t, tujuan), dumpMaster(t, asal)
	if got != want {
		t.Fatalf("round-trip tidak identik:\n%s\n---\n%s", got, want)
	}
	if !strings.Contains(got, "C|0012345678|VII A|0123456789|Siti|1") || !strings.Contains(got, "C|0000000042") {
		t.Fatalf("nol di depan UID/NISN hilang:\n%s", got)
	}

	// impor ulang: semuanya sama, tidak ada perubahan
	h, err = impor(t, tujuan, berkas, model.OpsiImpor{})
	harusOK(t, err)
	if !strings.HasPrefix(h.Pesan, "Tidak ada perubahan") {
		t.Fatalf("impor ulang: %+v", h)
	}
}

func TestSinkronGalatBaris(t *testing.T) {
	s := setupT5(t)
	_, err := s.TambahKelas("VII A", "Ani", 1)
	harusOK(t, err)
	p := tulisCSVUji(t, "uid,kelas,nisn,nama_siswa,aktif\n"+
		"1234567890,VII A,1234567890,Ani,1\n"+
		"1234567890,VII A,,Dobel,1\n"+
		"12ab,VII A,,X,1\n"+
		"2222222222,VII Z,,Y,1\n"+
		"3333333333,VII A,12x,Z,1\n")

	audit0, kartu0, kelas0 := hitung(t, s, `SELECT COUNT(*) FROM audit_log`), hitung(t, s, `SELECT COUNT(*) FROM kartu`), hitung(t, s, `SELECT COUNT(*) FROM kelas`)
	pr, err := s.PratinjauImpor(p, model.OpsiImpor{})
	harusOK(t, err)
	if pr.Ringkasan.Galat != 4 || pr.Ringkasan.KartuBaru != 1 {
		t.Fatalf("ringkasan: %+v", pr.Ringkasan)
	}
	var pesan []string
	for _, b := range pr.Baris {
		if b.Aksi == "GALAT" {
			pesan = append(pesan, b.Pesan)
		}
	}
	semua := strings.Join(pesan, "|")
	for _, w := range []string{"ganda", "1-10 digit", "tidak dikenal", "NISN"} {
		if !strings.Contains(semua, w) {
			t.Fatalf("galat %q tidak ada: %s", w, semua)
		}
	}
	// pratinjau tidak menulis apa pun
	if hitung(t, s, `SELECT COUNT(*) FROM audit_log`) != audit0 || hitung(t, s, `SELECT COUNT(*) FROM kartu`) != kartu0 || hitung(t, s, `SELECT COUNT(*) FROM kelas`) != kelas0 {
		t.Fatal("pratinjau menulis ke database")
	}
	// satu galat = seluruh impor batal, termasuk baris yang valid
	if _, err := impor(t, s, p, model.OpsiImpor{}); err == nil {
		t.Fatal("impor dengan galat harus ditolak")
	}
	if hitung(t, s, `SELECT COUNT(*) FROM kartu`) != 0 {
		t.Fatal("impor gagal tidak boleh menulis sebagian")
	}

	// opsi buat kelas otomatis: kelas tidak dikenal tidak lagi galat
	pr, err = s.PratinjauImpor(p, model.OpsiImpor{BuatKelasOtomatis: true})
	harusOK(t, err)
	if pr.Ringkasan.Galat != 3 || pr.Ringkasan.KelasOtomatis != 1 {
		t.Fatalf("otomatis: %+v", pr.Ringkasan)
	}

	p2 := tulisCSVUji(t, "uid,kelas\n2222222222,VII Z\n")
	if _, err := impor(t, s, p2, model.OpsiImpor{}); err == nil {
		t.Fatal("kelas tidak dikenal harus ditolak tanpa opsi otomatis")
	}
	_, err = impor(t, s, p2, model.OpsiImpor{BuatKelasOtomatis: true})
	harusOK(t, err)
	var ketua string
	harusOK(t, s.db.QueryRow(`SELECT nama_ketua FROM kelas WHERE nama='VII Z'`).Scan(&ketua))
	if ketua != "-" || hitung(t, s, `SELECT COUNT(*) FROM kartu WHERE uid='2222222222'`) != 1 {
		t.Fatalf("kelas otomatis: ketua=%q", ketua)
	}
}

func TestSinkronGabungDanGanti(t *testing.T) {
	s := setupT5(t)
	a, _ := s.TambahKelas("VII A", "Ani", 1)
	b, _ := s.TambahKelas("VII B", "Budi", 2)
	harusOK(t, s.DaftarkanKartu("1111111111", a, "", ""))
	harusOK(t, s.DaftarkanKartu("2222222222", b, "", ""))

	d := sinkron.Data{
		AdaKelas: true, AdaKartu: true,
		Kelas: []sinkron.KelasBaris{{Baris: 2, Nama: "VII A", Ketua: "Ani Baru", Urutan: "1", Aktif: "1"}},
		Kartu: []sinkron.KartuBaris{
			{Baris: 2, UID: "1111111111", Kelas: "VII A", Aktif: "1"},
			{Baris: 3, UID: "12345678", Kelas: "VII A", Nama: "Eka", Aktif: "1"}, // nol di depan dinormalkan
		},
	}
	berkas := filepath.Join(t.TempDir(), "uji.xlsx")
	harusOK(t, sinkron.TulisXLSX(berkas, d, nil))

	// Gabung: tidak menghapus apa pun
	_, err := impor(t, s, berkas, model.OpsiImpor{Mode: "GABUNG"})
	harusOK(t, err)
	var ketua string
	harusOK(t, s.db.QueryRow(`SELECT nama_ketua FROM kelas WHERE nama='VII A'`).Scan(&ketua))
	if ketua != "Ani Baru" || hitung(t, s, `SELECT COUNT(*) FROM kartu`) != 3 ||
		hitung(t, s, `SELECT COUNT(*) FROM kartu WHERE uid='0012345678'`) != 1 ||
		hitung(t, s, `SELECT aktif FROM kelas WHERE nama='VII B'`) != 1 {
		t.Fatalf("setelah Gabung: ketua=%q", ketua)
	}

	// Ganti tanpa konfirmasi atau alasan ditolak
	if _, err := impor(t, s, berkas, model.OpsiImpor{Mode: "GANTI"}); err == nil {
		t.Fatal("Ganti tanpa konfirmasi harus ditolak")
	}
	if _, err := impor(t, s, berkas, model.OpsiImpor{Mode: "GANTI", KonfirmasiGanti: true}); err == nil {
		t.Fatal("Ganti tanpa alasan harus ditolak")
	}
	_, err = impor(t, s, berkas, model.OpsiImpor{Mode: "GANTI", KonfirmasiGanti: true, Alasan: "sinkron ulang"})
	harusOK(t, err)
	if hitung(t, s, `SELECT COUNT(*) FROM kartu`) != 2 || hitung(t, s, `SELECT COUNT(*) FROM kartu WHERE uid='2222222222'`) != 0 {
		t.Fatal("Ganti harus menghapus kartu yang tidak ada di berkas")
	}
	if hitung(t, s, `SELECT aktif FROM kelas WHERE nama='VII B'`) != 0 || hitung(t, s, `SELECT aktif FROM kelas WHERE nama='VII A'`) != 1 {
		t.Fatal("Ganti harus menonaktifkan (bukan menghapus) kelas yang tidak ada di berkas")
	}

	// audit: ringkasan per impor dan entri per kartu
	au, err := s.ListAudit(model.FilterAudit{Entitas: "master"})
	harusOK(t, err)
	if len(au) != 2 || au[0].Aksi != "IMPOR_MASTER" || au[0].Alasan != "sinkron ulang" {
		t.Fatalf("audit master: %+v", au)
	}
	au, err = s.ListAudit(model.FilterAudit{Entitas: "kartu"})
	harusOK(t, err)
	hapus := false
	for _, x := range au {
		if x.Aksi == "HAPUS" && x.EntitasID == "2222222222" {
			hapus = true
		}
	}
	if !hapus {
		t.Fatalf("audit hapus kartu tidak ada: %+v", au)
	}
}

func TestSinkronHashBerubah(t *testing.T) {
	s := setupT5(t)
	_, err := s.TambahKelas("VII A", "Ani", 1)
	harusOK(t, err)
	p := tulisCSVUji(t, "uid,kelas\n1111111111,VII A\n")
	pr, err := s.PratinjauImpor(p, model.OpsiImpor{})
	harusOK(t, err)
	// berkas diubah setelah pratinjau
	harusOK(t, os.WriteFile(p, []byte("uid,kelas\n1111111111,VII A\n2222222222,VII A\n"), 0o600))
	if _, err := s.TerapkanImpor(p, model.OpsiImpor{Hash: pr.Hash}); err == nil || !strings.Contains(err.Error(), "berubah") {
		t.Fatalf("hash berbeda harus ditolak: %v", err)
	}
	if _, err := s.TerapkanImpor(p, model.OpsiImpor{}); err == nil {
		t.Fatal("tanpa hash harus ditolak")
	}
	if hitung(t, s, `SELECT COUNT(*) FROM kartu`) != 0 {
		t.Fatal("impor yang ditolak tidak boleh menulis")
	}
}

func TestSinkronCSVTemplateDanSemikolon(t *testing.T) {
	s := setupT5(t)
	isiMaster(t, s)
	d, err := s.dataMaster(false)
	harusOK(t, err)

	p := filepath.Join(t.TempDir(), "kartu.csv")
	harusOK(t, sinkron.TulisCSVKartu(p, d))
	raw, err := os.ReadFile(p)
	harusOK(t, err)
	if !strings.HasPrefix(string(raw), "\xEF\xBB\xBF") {
		t.Fatal("CSV harus berawalan BOM UTF-8")
	}
	got, err := sinkron.Baca(p)
	harusOK(t, err)
	if len(got.Kartu) != 3 || got.AdaKelas || got.Kartu[0].UID != "0012345678" && got.Kartu[1].UID != "0012345678" {
		t.Fatalf("baca ulang CSV: %+v", got)
	}

	// template Excel bisa diimpor tanpa galat dan tanpa perubahan
	tp := filepath.Join(t.TempDir(), "template.xlsx")
	harusOK(t, sinkron.TemplateXLSX(tp))
	tujuan := setupT5(t)
	pr, err := tujuan.PratinjauImpor(tp, model.OpsiImpor{})
	harusOK(t, err)
	if pr.Ringkasan != (model.RingkasanImpor{}) {
		t.Fatalf("template: %+v", pr.Ringkasan)
	}

	// template CSV terbaca
	tc := filepath.Join(t.TempDir(), "tk.csv")
	harusOK(t, sinkron.TemplateCSV(tc, "kartu"))
	g3, err := sinkron.Baca(tc)
	harusOK(t, err)
	if !g3.AdaKartu || len(g3.Kartu) != 0 {
		t.Fatalf("template CSV: %+v", g3)
	}

	// pemisah titik koma (Excel lokal Indonesia)
	g2, err := sinkron.Baca(tulisCSVUji(t, "uid;kelas\n12345678;VII A\n"))
	harusOK(t, err)
	if len(g2.Kartu) != 1 || g2.Kartu[0].UID != "12345678" || g2.Kartu[0].Kelas != "VII A" {
		t.Fatalf("CSV titik koma: %+v", g2)
	}
}
```

## 6. Jalankan dan commit

```bash
mkdir -p internal/sinkron      # lalu simpan berkas bagian 1; tambahkan bagian 2 ke model.go; simpan bagian 3 dan 5
gofmt -l . ; go vet ./... && go test ./...
wails build                    # regenerasi frontend/wailsjs/go/models.ts dan Service.js
git add -A
git commit -m "Sinkronisasi master data: ekspor/impor Excel dan CSV dengan pratinjau"
```

Hal yang perlu dicek bila ada yang gagal:
- Fungsi excelize `SetColStyle`, `SetCellStr`, dan `NewSheet` saya tulis dari ingatan API v2.x, dan di v2.11 ada kemungkinan tanda tangannya sedikit berbeda. Kirim pesan galat kompilasinya dan saya sesuaikan.
- Di `TestSinkronCSVTemplateDanSemikolon`, kondisi `||` pada pengecekan `got.Kartu` sengaja longgar soal urutan baris. Bila `go vet` mengeluh soal prioritas operator, tambahkan kurung.
- `wails build` perlu dijalankan sebelum UI dicoba, karena binding baru (`PratinjauImpor`, dan seterusnya) belum ada di `Service.js` lama.

## 7. Ringkasan dan versi

- **Ekspor:** satu Excel dengan sheet Pengaturan, Kelas, dan Kartu, atau CSV per jenis (UTF-8 BOM). Semua sel berformat teks sehingga nol di depan UID dan NISN terjaga. Ekspor ditolak bila folder tujuan berada di dalam repo git.
- **Impor:** selalu lewat pratinjau (baru, diubah, sama, dilewati, galat per baris). Impor berjalan dalam satu transaksi dengan hash pratinjau yang diperiksa ulang. Mode Gabung hanya menambah dan memperbarui. Mode Ganti butuh konfirmasi dan alasan, menghapus kartu, dan menonaktifkan kelas.
- **Audit:** satu entri `IMPOR_MASTER`, ditambah satu entri per kelas, kartu, atau pengaturan yang berubah.
- **Skema:** tidak berubah, dan `go.mod` tidak bertambah.

Saran versi **0.3.0** (fitur baru, skema tetap). Nomor versinya belum saya ubah, karena di Fitur 2 versi dipindah ke `-ldflags`.

Setelah tes lulus, kirim hasilnya dan jalankan perintah ini supaya saya bisa menyusun rencana Fitur 2 (pembaruan via GitHub Release):

```bash
for f in wails.json nfpm.yaml .gitignore build/linux/postinst.sh build/linux/99-mbg-reader.rules; do echo "=== $f ==="; cat "$f"; done
git remote -v; git tag | tail -5
```

zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ for f in wails.json nfpm.yaml .gitignore build/linux/postinst.sh build/linux/99-mbg-reader.rules; do echo "=== $f ==="; cat "$f"; done
git remote -v; git tag | tail -5
=== wails.json ===
{
    "$schema": "https://wails.io/schemas/config.v2.json",
    "name": "mbg",
    "outputfilename": "mbg",
    "frontend:dir": "frontend",
    "assetdir": "frontend",
    "build:tags": "webkit2_41"
}
=== nfpm.yaml ===
name: mbg
arch: amd64
platform: linux
version: 0.1.1
section: misc
maintainer: Yahya Zulfikri <https://zulfikriyahya.my.id>
description: Aplikasi pencatatan pengambilan dan pengembalian ompreng MBG
homepage: https://zulfikriyahya.my.id
contents:
  - src: dist/mbg-linux-amd64
    dst: /usr/bin/mbg
    file_info:
      mode: 0755
  - src: build/linux/mbg.desktop
    dst: /usr/share/applications/mbg.desktop
    file_info:
      mode: 0644
  - src: build/linux/mbg.png
    dst: /usr/share/icons/hicolor/512x512/apps/mbg.png
    file_info:
      mode: 0644
  - src: build/linux/99-mbg-reader.rules
    dst: /etc/udev/rules.d/99-mbg-reader.rules
    file_info:
      mode: 0644
scripts:
  postinstall: build/linux/postinst.sh
overrides:
  deb:
    depends:
      - libgtk-3-0
      - libwebkit2gtk-4.1-0
  rpm:
    depends:
      - gtk3
      - webkit2gtk4.1
=== .gitignore ===
# Hasil build
dist/
build/bin/
build/windows/installer/tmp/
frontend/wailsjs/

# Log dan data lokal
*.log
*.db
*.db-wal
*.db-shm

# Catatan sementara
draft.md
=== build/linux/postinst.sh ===
#!/bin/sh
udevadm control --reload-rules 2>/dev/null || true
udevadm trigger 2>/dev/null || true
=== build/linux/99-mbg-reader.rules ===
# Reader RFID ESP32-C3 (Espressif) bisa diakses semua pengguna
SUBSYSTEM=="tty", ATTRS{idVendor}=="303a", MODE="0666"
origin	https://github.com/zulfikriyahya/sirkulasi-ompreng.git (fetch)
origin	https://github.com/zulfikriyahya/sirkulasi-ompreng.git (push)
v0.1.0
v0.1.1
zulfikriyahya@zedlabs:~/ZEDLABS/transaksi-mbg/mbg$ 
