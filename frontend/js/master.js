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
    const bv = document.getElementById("info-versi");
    bv.textContent = "Versi " + i.versi;
    bv.classList.toggle("dev", String(i.versi).toLowerCase() === "dev");
    const bd = document.getElementById("info-data");
    document.getElementById("info-data-teks").textContent = i.data_dir;
    bd.dataset.salin = i.data_dir;
    bd.title = "Klik untuk menyalin: " + i.data_dir;
    bd.disabled = false;
  },
  async simpanInstansi() {
    try {
      const lama = await API.GetPengaturan();
      lama.nama_instansi = val("p-nama");
      lama.alamat_instansi = val("p-alamat");
      lama.penanggung_jawab = val("p-pj");
      lama.nip_penanggung_jawab = val("p-nip");
      lama.jabatan_penanggung_jawab = val("p-jabatan");
      lama.jeda_minimum_menit = parseInt(val("p-jeda"), 10);
      lama.hid_ambang_ms = parseInt(val("p-hid"), 10);
      await API.SimpanPengaturan(lama);
      Tap.ambangMs = lama.hid_ambang_ms;
      toast("Pengaturan disimpan", "ok");
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },

  async muatKelas() {
    const cb = document.getElementById("k-semua");
    const list = await API.ListKelas(true);
    const wali = await API.ListWaliKelas(false);
    const tb = document.querySelector("#tabel-kelas tbody");
    tb.innerHTML = "";
    list.forEach((k) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${k.urutan}</td><td>${esc(k.nama)}</td><td>${esc(k.nama_ketua)}</td><td class="sel-wali"></td><td class="aksi-sel"><div class="aksi-baris"></div></td>`;
      const sw = document.createElement("select");
      sw.add(new Option("- belum ada -", 0));
      wali.forEach((w) => sw.add(new Option(w.nama, w.id)));
      sw.value = k.wali_kelas_id || 0;
      sw.disabled = !k.aktif;
      sw.onchange = async () => {
        try {
          await API.SetWaliKelas(k.id, parseInt(sw.value, 10));
          k.wali_kelas_id = parseInt(sw.value, 10);
          toast("Wali kelas disimpan", "ok");
          this.muatWali();
        } catch (e) {
          sw.value = k.wali_kelas_id || 0;
          toast(String(e.message || e), "err");
        }
      };
      tr.querySelector(".sel-wali").appendChild(sw);
      const wadah = tr.querySelector(".aksi-baris");
      if (k.aktif) {
        const ub = document.createElement("button");
        ub.textContent = "Ubah";
        ub.onclick = () => {
          this.editId = k.id;
          setv("k-nama", k.nama);
          setv("k-ketua", k.nama_ketua);
          setv("k-urutan", k.urutan);
          document.getElementById("k-simpan").textContent = "Simpan";
          document.getElementById("k-batal").classList.remove("hidden");
        };
        const nb = document.createElement("button");
        nb.textContent = "Nonaktifkan";
        nb.className = "bahaya";
        nb.onclick = async () => {
          const h = await Dash.konfirmasi({
            mode: "Nonaktifkan",
            judul: "Nonaktifkan kelas?",
            label: "Kelas",
            nama: k.nama,
            detail: "Ketua: " + k.nama_ketua,
            pesan: "Kelas bisa diaktifkan kembali lewat opsi Tampilkan kelas nonaktif.",
            aksi: "Nonaktifkan",
            alasan: true,
          });
          if (!h) return;
          try {
            await API.NonaktifkanKelas(k.id, h.alasan || "dari layar master");
            this.muatKelas();
            Dash.muat();
            toast("Kelas dinonaktifkan", "ok", "hapus");
          } catch (e) {
            toast(String(e.message || e), "err");
          }
        };
        wadah.append(ub, nb);
      } else {
        tr.style.opacity = "0.55";
        const ab = document.createElement("button");
        ab.textContent = "Aktifkan";
        ab.onclick = async () => {
          try {
            await API.AktifkanKelas(k.id, "dari layar master");
            this.muatKelas();
            Dash.muat();
            toast("Kelas diaktifkan", "ok");
          } catch (e) {
            toast(String(e.message || e), "err");
          }
        };
        wadah.append(ab);
      }
      tb.appendChild(tr);
    });
    // isi dropdown kelas di tab kartu
    const sel = document.getElementById("c-kelas");
    const cur = sel.value;
    sel.innerHTML = "";
    list
      .filter((k) => k.aktif)
      .forEach((k) => sel.add(new Option(k.nama + " - " + k.nama_ketua, k.id)));
    if (cur) sel.value = cur;
    this.muatKartu();
  },
  resetFormKelas() {
    this.editId = null;
    setv("k-nama", "");
    setv("k-ketua", "");
    setv("k-urutan", 0);
    document.getElementById("k-simpan").textContent = "Tambah";
    document.getElementById("k-batal").classList.add("hidden");
  },
  async simpanKelas() {
    try {
      const n = val("k-nama"),
        k = val("k-ketua"),
        u = parseInt(val("k-urutan") || "0", 10);
      if (this.editId) await API.UbahKelas(this.editId, n, k, u);
      else await API.TambahKelas(n, k, u);
      this.resetFormKelas();
      toast("Kelas tersimpan", "ok");
      this.muatKelas();
      Dash.muat();
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },

  editWaliId: null,

  async muatWali() {
    const semua = document.getElementById("w-semua").checked;
    const list = await API.ListWaliKelas(semua);
    const tb = document.querySelector("#tabel-wali tbody");
    tb.innerHTML = "";
    if (!list.length) {
      tb.innerHTML = `<tr><td colspan="4">Belum ada data wali kelas</td></tr>`;
      return;
    }
    list.forEach((w) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${esc(w.nama)}</td><td>${esc(w.no_wa)}</td><td>${esc(w.daftar_kelas || "-")}</td><td class="aksi-sel"><div class="aksi-baris"></div></td>`;
      const wadah = tr.querySelector(".aksi-baris");
      if (w.aktif) {
        const ub = document.createElement("button");
        ub.textContent = "Ubah";
        ub.onclick = () => {
          this.editWaliId = w.id;
          setv("w-nama", w.nama);
          setv("w-wa", w.no_wa);
          document.getElementById("w-simpan").textContent = "Simpan";
          document.getElementById("w-batal").classList.remove("hidden");
        };
        const nb = document.createElement("button");
        nb.textContent = "Nonaktifkan";
        nb.className = "bahaya";
        nb.onclick = async () => {
          const h = await Dash.konfirmasi({
            mode: "Nonaktifkan",
            judul: "Nonaktifkan wali kelas?",
            label: "Wali kelas",
            nama: w.nama,
            detail: "Kelas diampu: " + (w.daftar_kelas || "-"),
            pesan: "Tautan ke kelasnya akan dilepas.",
            aksi: "Nonaktifkan",
            alasan: true,
          });
          if (!h) return;
          try {
            await API.NonaktifkanWaliKelas(w.id, h.alasan || "dari layar master");
            this.muatWali();
            this.muatKelas();
            toast("Wali kelas dinonaktifkan", "ok", "hapus");
          } catch (e) {
            toast(String(e.message || e), "err");
          }
        };
        wadah.append(ub, nb);
      } else {
        tr.style.opacity = "0.55";
        const ab = document.createElement("button");
        ab.textContent = "Aktifkan";
        ab.onclick = async () => {
          try {
            await API.AktifkanWaliKelas(w.id, "dari layar master");
            toast("Wali kelas diaktifkan", "ok");
            this.muatWali();
            this.muatKelas();
          } catch (e) {
            toast(String(e.message || e), "err");
          }
        };
        wadah.append(ab);
      }
      tb.appendChild(tr);
    });
  },
  resetFormWali() {
    this.editWaliId = null;
    setv("w-nama", "");
    setv("w-wa", "");
    document.getElementById("w-simpan").textContent = "Tambah";
    document.getElementById("w-batal").classList.add("hidden");
  },
  async simpanWali() {
    try {
      const n = val("w-nama"),
        wa = val("w-wa");
      if (this.editWaliId) await API.UbahWaliKelas(this.editWaliId, n, wa);
      else await API.TambahWaliKelas(n, wa);
      this.resetFormWali();
      toast("Wali kelas tersimpan", "ok");
      this.muatWali();
      this.muatKelas();
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },

  async muatKartu() {
    const kid = parseInt(val("c-kelas"), 10);
    const tb = document.querySelector("#tabel-kartu tbody");
    tb.innerHTML = "";
    if (!kid) return;
    (await API.ListKartu(kid)).forEach((c) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${esc(c.uid)}</td><td>${esc(c.nisn || "-")}</td><td>${esc(c.nama_siswa || "-")}</td><td class="aksi-sel"><div class="aksi-baris"></div></td>`;
      const b = document.createElement("button");
      b.textContent = "Hapus";
      b.className = "bahaya";
      b.onclick = async () => {
        const h = await Dash.konfirmasi({
          mode: "Hapus",
          judul: "Hapus kartu?",
          label: "Kartu " + c.uid,
          nama: c.nama_siswa || "(tanpa nama)",
          detail: c.nisn ? "NISN " + c.nisn : "",
          pesan: "Daftarkan ulang bila kartu ini ingin dipakai lagi.",
          aksi: "Hapus",
        });
        if (!h) return;
        try {
          await API.HapusKartu(c.uid);
          this.muatKartu();
          toast("Kartu dihapus", "ok", "hapus");
        } catch (e) {
          toast(String(e.message || e), "err");
        }
      };
      tr.querySelector(".aksi-baris").appendChild(b);
      tb.appendChild(tr);
    });
  },
  async daftarKartu() {
    try {
      const kid = parseInt(val("c-kelas"), 10);
      if (!kid) throw new Error("Pilih kelas dulu");
      await API.DaftarkanKartu(val("c-uid"), kid, val("c-nisn").trim(), val("c-nama").trim());
      setv("c-uid", "");
      setv("c-nisn", "");
      setv("c-nama", "");
      toast("Kartu didaftarkan", "ok");
      this.muatKartu();
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },
};

function val(id) {
  return document.getElementById(id).value;
}
function setv(id, v) {
  document.getElementById(id).value = v;
}
