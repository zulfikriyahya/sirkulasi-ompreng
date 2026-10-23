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
function setv(id, v) { document.getElementById(id).value = v; }