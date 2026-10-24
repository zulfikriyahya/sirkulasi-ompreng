const Sinkron = {
  path: "",
  hasil: null,

  async unduh(fn) {
    try {
      const p = await fn();
      if (p) toast("Tersimpan: " + p, "ok");
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },

  async mulaiImpor() {
    try {
      const p = await API.PilihBerkasImpor();
      if (!p) return;
      this.path = p;
      this.hasil = null;
      this.tampilModal();
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },

  tutup() {
    Dash.tutupModal();
    pasangHandler();
  },

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
    _r("im-hasil").innerHTML = "";
    _r("im-galat").textContent = "";
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
<option value="GANTI">Ganti (hapus kartu / nonaktifkan kelas dan wali yang tidak ada di berkas)</option>
            </select>
          </label>
          <label class="cek"><input type="checkbox" id="im-otomatis"> Buat kelas otomatis bila kelas pada kartu belum ada</label>
          <div id="im-ganti" class="hidden">
<div class="peringatan">Mode Ganti menghapus kartu serta menonaktifkan kelas dan wali kelas yang tidak ada di berkas.</div>
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
    const o = this.opsi();
    o.hash = "";
    try {
      this.hasil = await API.PratinjauImpor(this.path, o);
      this.render();
    } catch (e) {
      this.hasil = null;
      _r("im-galat").textContent = String(e.message || e);
    }
  },

  render() {
    const h = this.hasil,
      r = h.ringkasan;
    const ubah =
      r.kelas_baru +
      r.kelas_ubah +
      r.kelas_nonaktif +
      r.kelas_otomatis +
      r.kartu_baru +
      r.kartu_ubah +
      r.kartu_hapus +
      r.pengaturan_ubah +
      r.wali_baru +
      r.wali_ubah +
      r.wali_nonaktif;
    const ringkas = [
      `Wali kelas: ${r.wali_baru} baru, ${r.wali_ubah} diubah, ${r.wali_nonaktif} dinonaktifkan, ${r.wali_sama} sama`,
      `Kelas: ${r.kelas_baru} baru, ${r.kelas_ubah} diubah, ${r.kelas_otomatis} dibuat otomatis, ${r.kelas_nonaktif} dinonaktifkan, ${r.kelas_sama} sama`,
      `Kartu: ${r.kartu_baru} baru, ${r.kartu_ubah} diubah, ${r.kartu_hapus} dihapus, ${r.kartu_sama} sama`,
      `Pengaturan: ${r.pengaturan_ubah} diubah`,
      `Galat: ${r.galat} · Peringatan: ${r.peringatan}`,
    ];
    let html = `<p>${ringkas.map(esc).join("<br>")}</p>`;
    if (r.galat === 0 && ubah === 0)
      html += `<p class="kecil">Tidak ada perubahan: data sudah sama dengan berkas.</p>`;
    if (h.baris.length) {
      html +=
        `<div class="pratinjau-wadah"><table><thead><tr><th>Entitas</th><th>Baris</th><th>Kunci</th><th>Aksi</th><th>Keterangan</th></tr></thead><tbody>` +
        h.baris
          .map(
            (b) =>
              `<tr class="${b.aksi === "GALAT" ? "merah" : ""}"><td>${esc(b.entitas)}</td><td>${b.baris || "-"}</td><td>${esc(b.kunci)}</td><td>${esc(b.aksi)}</td><td>${esc(b.pesan)}</td></tr>`
          )
          .join("") +
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
      Master.muatWali();
      Master.muatKelas();
      Dash.muat();
    } catch (e) {
      _r("im-galat").textContent = String(e.message || e);
      bunyi("err");
    }
  },
};

document.addEventListener("DOMContentLoaded", () => {
  _r("sk-xlsx").onclick = () => Sinkron.unduh(() => API.EksporMasterExcel());
  _r("sk-csv-kelas").onclick = () => Sinkron.unduh(() => API.EksporMasterCSV("kelas"));
  _r("sk-csv-kartu").onclick = () => Sinkron.unduh(() => API.EksporMasterCSV("kartu"));
  _r("sk-csv-wali").onclick = () => Sinkron.unduh(() => API.EksporMasterCSV("wali"));
  _r("sk-tpl-xlsx").onclick = () => Sinkron.unduh(() => API.UnduhTemplateExcel());
  _r("sk-tpl-kelas").onclick = () => Sinkron.unduh(() => API.UnduhTemplateCSV("kelas"));
  _r("sk-tpl-wali").onclick = () => Sinkron.unduh(() => API.UnduhTemplateCSV("wali"));
  _r("sk-tpl-kartu").onclick = () => Sinkron.unduh(() => API.UnduhTemplateCSV("kartu"));
  _r("sk-impor").onclick = () => Sinkron.mulaiImpor();
});
