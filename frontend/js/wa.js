// Panel pengaturan WhatsApp: konfigurasi, uji kirim, riwayat notifikasi.
(function () {
  const S = () => window.go.service.Service;
  const esc = (s) =>
    String(s ?? "").replace(
      /[&<>"']/g,
      (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]
    );
  const waktu = (s) => (s ? new Date(s).toLocaleString("id-ID") : "-");
  let timer = null;

  const html = `
  <div class="wa-layout">
    <section class="wa-kolom">
      <h3>Konfigurasi</h3>
      <form id="wa-form" class="wa-grid">
        <label class="wa-cek"><input type="checkbox" id="wa-aktif"> Aktifkan notifikasi ke wali kelas</label>
        <label class="wa-penuh">Alamat layanan <input id="wa-url" type="url" required></label>
        <label>Key ID <input id="wa-key" autocomplete="off"></label>
        <label>Secret <input id="wa-secret" type="password" autocomplete="new-password"></label>
        <label>Template pengambilan <input id="wa-tpl-ambil"></label>
        <label>Template pengembalian <input id="wa-tpl-kembali"></label>
        <div class="wa-penuh wa-aksi"><button type="submit" class="utama">Simpan</button> <span id="wa-pesan-simpan"></span></div>
      </form>
    </section>
    <section class="wa-kolom">
      <h3>Uji kirim</h3>
      <p class="kecil">Mengirim pesan contoh memakai template pengambilan ke nomor di bawah.</p>
      <div class="wa-baris">
        <input id="wa-uji-no" class="wa-lebar" placeholder="Nomor WhatsApp, mis. 0812xxxxxxx">
        <button type="button" id="wa-uji-btn">Kirim pesan uji</button>
      </div>
      <span id="wa-pesan-uji"></span>
    </section>
  </div>
  <section class="wa-riwayat">
    <h3>Riwayat notifikasi</h3>
    <div class="wa-baris">
      <span id="wa-ringkas"></span>
      <span class="spasi"></span>
      <select id="wa-filter">
        <option value="">Semua</option><option value="ANTRE">Antre</option>
        <option value="TERKIRIM">Terkirim</option><option value="GAGAL">Gagal</option>
      </select>
      <button type="button" id="wa-segar">Segarkan</button>
    </div>
    <div class="wa-tabel"><table>
      <thead><tr><th>Waktu</th><th>Jenis</th><th>Wali</th><th>Nomor</th><th>Status</th><th>Keterangan</th><th></th></tr></thead>
      <tbody id="wa-isi"></tbody>
    </table></div>
  </section>`;

  const $ = (id) => document.getElementById(id);
  const pesan = (id, teks, galat, senyap) => {
    const e = $(id);
    e.textContent = teks;
    e.className = galat ? "wa-galat" : "wa-ok";
    if (!senyap) bunyi(galat ? "err" : "ok");
  };

  async function muatConfig() {
    const c = await S().GetWAConfig();
    $("wa-aktif").checked = c.aktif;
    $("wa-url").value = c.base_url;
    $("wa-key").value = c.key_id;
    $("wa-secret").value = "";
    $("wa-secret").placeholder = c.secret_ada
      ? "(tersimpan, kosongkan bila tidak diubah)"
      : "Secret API key";
    $("wa-tpl-ambil").value = c.template_ambil;
    $("wa-tpl-kembali").value = c.template_kembali;
  }

  async function muatRiwayat() {
    try {
      const [r, daftar] = await Promise.all([
        S().RingkasanWA(),
        S().DaftarAntreanWA($("wa-filter").value, 100),
      ]);
      $("wa-ringkas").textContent = `Antre ${r.antre} · Terkirim ${r.terkirim} · Gagal ${r.gagal}`;
      $("wa-isi").innerHTML = daftar.length
        ? daftar
            .map((a) => {
              const ket =
                a.status === "TERKIRIM"
                  ? a.status_layanan || ""
                  : [
                      a.galat,
                      a.status === "ANTRE" && a.coba_lagi ? "coba lagi " + waktu(a.coba_lagi) : "",
                    ]
                      .filter(Boolean)
                      .join(" · ");
              return `<tr>
          <td>${esc(waktu(a.dibuat))}</td><td>${esc(a.jenis)}</td><td>${esc(a.nama_wali)}</td><td>${esc(a.no_wa)}</td>
          <td><span class="wa-lencana wa-${esc(a.status.toLowerCase())}">${esc(a.status)}</span></td>
          <td>${esc(ket)}</td>
          <td>${a.status === "GAGAL" ? `<button type="button" data-ulang="${a.id}">Ulangi</button>` : ""}</td>
        </tr>`;
            })
            .join("")
        : '<tr><td colspan="7">Belum ada notifikasi.</td></tr>';
    } catch (e) {
      $("wa-isi").innerHTML = `<tr><td colspan="7" class="wa-galat">${esc(e)}</td></tr>`;
    }
  }

  function pasang(el) {
    el.innerHTML = html;
    $("wa-form").addEventListener("submit", async (ev) => {
      ev.preventDefault();
      try {
        await S().SimpanWAConfig({
          aktif: $("wa-aktif").checked,
          base_url: $("wa-url").value,
          key_id: $("wa-key").value,
          secret: $("wa-secret").value,
          secret_ada: false,
          template_ambil: $("wa-tpl-ambil").value,
          template_kembali: $("wa-tpl-kembali").value,
        });
        pesan("wa-pesan-simpan", "Tersimpan.");
        await muatConfig();
        muatRiwayat();
      } catch (e) {
        pesan("wa-pesan-simpan", String(e), true);
      }
    });
    $("wa-uji-btn").addEventListener("click", async () => {
      const btn = $("wa-uji-btn");
      btn.disabled = true;
      pesan("wa-pesan-uji", "Mengirim...", false, true);
      try {
        const u = await S().UjiKirimWA($("wa-uji-no").value);
        pesan("wa-pesan-uji", `Diterima layanan (job ${u.job_id}, status ${u.status}).`);
      } catch (e) {
        pesan("wa-pesan-uji", String(e), true);
      }
      btn.disabled = false;
    });
    $("wa-filter").addEventListener("change", muatRiwayat);
    $("wa-segar").addEventListener("click", muatRiwayat);
    $("wa-isi").addEventListener("click", async (ev) => {
      const id = ev.target.dataset && ev.target.dataset.ulang;
      if (!id) return;
      try {
        await S().UlangiWA(Number(id));
        toast("Diantrekan ulang", "info");
      } catch (e) {
        // prettier-ignore
        bunyi("err"); alert(e);
      }
      muatRiwayat();
    });
  }

  window.WA = {
    // Panggil saat tab/halaman pengaturan dibuka.
    async muat(selector = "#wa-panel") {
      const el = document.querySelector(selector);
      if (!el) return;
      if (!$("wa-form")) pasang(el);
      await muatConfig().catch((e) => pesan("wa-pesan-simpan", String(e), true, true));
      muatRiwayat();
      clearInterval(timer);
      timer = setInterval(() => {
        if (el.offsetParent !== null) muatRiwayat();
      }, 15000);
    },
  };
})();
document.addEventListener("click", (e) => {
  if (e.target.closest && e.target.closest('[data-sub="wa"]')) window.WA.muat();
});
