const Pembaruan = {
  info: null,

  tampilBadge(i) {
    const b = _r("badge-pembaruan");
    b.textContent = "Pembaruan v" + i.versi_baru;
    b.classList.remove("hidden");
  },

  // dipanggil beberapa detik setelah aplikasi terbuka; gagal diam-diam
  async otomatis() {
    try {
      const i = await API.PeriksaPembaruan(true);
      if (i && i.ada) {
        this.info = i;
        this.tampilBadge(i);
      }
    } catch (e) {}
  },

  async manual() {
    try {
      const i = await API.PeriksaPembaruan(false);
      this.info = i;
      if (i.ada) {
        this.tampilBadge(i);
        this.tampil(i);
      } else toast(i.pesan || "Aplikasi sudah versi terbaru", "ok");
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },

  async muatOpsi() {
    try {
      const o = await API.GetOpsiPembaruan();
      _r("pb-otomatis").checked = o.otomatis;
      _r("pb-pra").checked = o.pra_rilis;
    } catch (e) {}
  },

  async simpanOpsi() {
    try {
      await API.SimpanOpsiPembaruan({
        otomatis: _r("pb-otomatis").checked,
        pra_rilis: _r("pb-pra").checked,
      });
      toast("Pengaturan pembaruan disimpan", "ok");
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },

  tutup() {
    Dash.tutupModal();
    pasangHandler();
  },

  tampil(i) {
    Tap.handler = () => toast("Selesaikan pembaruan dulu", "info");
    const manual = i.cara === "MANUAL";
    const mb = i.ukuran ? ` (${(i.ukuran / 1048576).toFixed(1)} MB)` : "";
    Dash.bukaModal(`
          <h2>Pembaruan tersedia: v${esc(i.versi_baru)}</h2>
          <p>Versi terpasang: v${esc(i.versi_saat_ini)}${manual ? "" : " &middot; Berkas: " + esc(i.nama_aset) + mb}</p>
          <pre class="catatan-rilis">${esc(i.catatan || "(tanpa catatan rilis)")}</pre>
          ${manual ? `<div class="peringatan">Tidak ada berkas pemasang yang cocok untuk sistem ini. Unduh manual dari halaman rilis.</div>` : ""}
          ${!manual && !i.checksum_ada ? `<div class="peringatan">Rilis ini tanpa checksums.txt: hanya ukuran berkas yang diperiksa.</div>` : ""}
          <p class="kecil">Database dicadangkan otomatis sebelum pembaruan dimulai.</p>
          <progress id="pb-prog" class="hidden" max="100" value="0"></progress>
          <div class="kecil" id="pb-status"></div>
          <div class="galat" id="pb-galat"></div>
          <div class="aksi"><button id="pb-batal">Nanti</button><button id="pb-ok">${manual ? "Buka halaman rilis" : "Unduh dan pasang"}</button></div>`);
    _r("pb-batal").onclick = () => this.tutup();
    _r("pb-ok").onclick = () => (manual ? this.bukaHalaman(i.url_halaman) : this.pasang());
  },

  bukaHalaman(u) {
    if (!/^https:\/\/github\.com\//.test(u || "")) {
      toast("Alamat rilis tidak valid", "err");
      return;
    }
    if (window.runtime && window.runtime.BrowserOpenURL) window.runtime.BrowserOpenURL(u);
    else window.open(u, "_blank");
  },

  progres(p) {
    const el = _r("pb-prog");
    if (!el) return;
    const pct = p.total > 0 ? Math.min(100, Math.floor((p.selesai * 100) / p.total)) : 0;
    el.classList.remove("hidden");
    el.value = pct;
    _r("pb-status").textContent = `Mengunduh... ${pct}%`;
  },

  async pasang() {
    _r("pb-ok").disabled = true;
    _r("pb-batal").disabled = true;
    _r("pb-galat").textContent = "";
    _r("pb-status").textContent = "Menyiapkan cadangan dan unduhan...";
    try {
      const h = await API.PasangPembaruan();
      if (h.keluar) {
        _r("pb-status").textContent = h.pesan;
        return;
      }
      this.hasilLinux(h);
    } catch (e) {
      _r("pb-galat").textContent = String(e.message || e);
      _r("pb-status").textContent = "";
      _r("pb-ok").disabled = false;
      _r("pb-batal").disabled = false;
      bunyi("err");
    }
  },

  hasilLinux(h) {
    Dash.bukaModal(`
          <h2>Siap dipasang</h2>
          <p>${esc(h.pesan)}</p>
          <input id="pb-cmd" readonly value="${esc(h.perintah)}">
          <p class="kecil">Berkas: ${esc(h.berkas)}<br>Cadangan database: ${esc(h.cadangan)}</p>
          <div class="aksi"><button id="pb-tutup">Tutup</button><button id="pb-salin">Salin perintah</button></div>`);
    _r("pb-tutup").onclick = () => this.tutup();
    _r("pb-salin").onclick = async () => {
      const el = _r("pb-cmd");
      try {
        await navigator.clipboard.writeText(el.value);
      } catch (e) {
        el.select();
        document.execCommand("copy");
      }
      toast("Perintah disalin", "ok");
    };
  },
};

// opsi pembaruan dimuat bersama tab Master > Instansi
const _muatInstansiPb = Master.muatInstansi.bind(Master);
Master.muatInstansi = async function () {
  await _muatInstansiPb();
  await Pembaruan.muatOpsi();
};

document.addEventListener("DOMContentLoaded", () => {
  _r("pb-periksa").onclick = () => Pembaruan.manual();
  _r("pb-otomatis").onchange = () => Pembaruan.simpanOpsi();
  _r("pb-pra").onchange = () => Pembaruan.simpanOpsi();
  _r("badge-pembaruan").onclick = () => {
    if (Pembaruan.info && Pembaruan.info.ada) Pembaruan.tampil(Pembaruan.info);
  };
  if (window.runtime && window.runtime.EventsOn)
    window.runtime.EventsOn("pembaruan:progres", (p) => Pembaruan.progres(p));
  setTimeout(() => Pembaruan.otomatis(), 5000);
});
