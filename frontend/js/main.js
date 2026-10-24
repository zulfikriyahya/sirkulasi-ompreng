let tabAktif = "beranda",
  subAktif = "instansi";

function pilihTab(t) {
  Suara.tahan();
  tabAktif = t;
  document
    .querySelectorAll(".sidebar nav button")
    .forEach((b) => b.classList.toggle("aktif", b.dataset.tab === t));
  document
    .querySelectorAll(".tab")
    .forEach((s) => s.classList.toggle("aktif", s.id === "tab-" + t));
  if (t === "master") (Master.muatInstansi(), Master.muatWali(), Master.muatKelas());
  else if (t === "riwayat") Riwayat.muat();
  else if (t === "ekspor") Ekspor.muat();
  else Dash.muat();
  pasangHandler();
}

function pilihSub(s) {
  Suara.tahan();
  subAktif = s;
  document
    .querySelectorAll(".subnav button")
    .forEach((b) => b.classList.toggle("aktif", b.dataset.sub === s));
  document
    .querySelectorAll(".sub")
    .forEach((x) => x.classList.toggle("aktif", x.id === "sub-" + s));
  pasangHandler();
}
// Di tab Master > Kartu, tap mengisi kolom UID; di Riwayat/Ekspor tap ditolak; selain itu diproses sebagai tap transaksi.
function pasangHandler() {
  if (tabAktif === "master" && subAktif === "kartu") {
    Tap.handler = (uid) => {
      setv("c-uid", uid);
      toast("UID terbaca: " + uid, "info");
    };
  } else if (tabAktif === "riwayat" || tabAktif === "ekspor") {
    Tap.handler = () => toast("Pindah ke tab Beranda untuk memproses tap", "info", "peringatan");
  } else {
    Tap.handler = prosesTap;
  }
}

document.addEventListener("DOMContentLoaded", async () => {
  document
    .querySelectorAll(".sidebar nav button")
    .forEach((b) => (b.onclick = () => pilihTab(b.dataset.tab)));
  document
    .querySelectorAll(".subnav button")
    .forEach((b) => (b.onclick = () => pilihSub(b.dataset.sub)));
  document.getElementById("p-simpan").onclick = () => Master.simpanInstansi();
  document.getElementById("k-simpan").onclick = () => Master.simpanKelas();
  document.getElementById("k-batal").onclick = () => Master.resetFormKelas();
  document.getElementById("c-daftar").onclick = () => Master.daftarKartu();
  document.getElementById("c-kelas").onchange = () => Master.muatKartu();
  document.getElementById("k-semua").onchange = () => Master.muatKelas();
  document.getElementById("w-simpan").onclick = () => Master.simpanWali();
  document.getElementById("w-batal").onclick = () => Master.resetFormWali();
  document.getElementById("w-semua").onchange = () => Master.muatWali();
  document.getElementById("btn-uid-manual").onclick = () => {
    Dash.bukaModal(`<h2>Input UID manual</h2><p>Ketik UID kartu bila reader tidak tersedia.</p><input id="um" maxlength="12" placeholder="UID"><div class="galat"></div>
      <div class="aksi"><button id="um-batal">Batal</button><button id="um-ok">Proses</button></div>`);
    const i = document.getElementById("um");
    i.focus();
    const ok = () => {
      const u = i.value;
      Dash.tutupModal();
      prosesTap(u);
    };
    document.getElementById("um-ok").onclick = ok;
    i.onkeydown = (e) => {
      if (e.key === "Enter") ok();
    };
    document.getElementById("um-batal").onclick = () => Dash.tutupModal();
  };
  const bd = document.getElementById("info-data");
  if (bd)
    bd.onclick = async () => {
      const t = bd.dataset.salin;
      if (!t) return;
      await salin(t);
      toast("Lokasi data disalin", "ok");
      const lb = bd.querySelector(".jalur-label");
      bd.classList.add("tersalin");
      lb.textContent = "Tersalin";
      setTimeout(() => {
        bd.classList.remove("tersalin");
        lb.textContent = "Data";
      }, 1400);
    };

  try {
    const p = await API.GetPengaturan();
    Tap.ambangMs = p.hid_ambang_ms;
  } catch (e) {}
  pasangHandler();
  Dash.muat();
  setInterval(() => {
    if (tabAktif === "beranda" && document.getElementById("modal").className === "hidden")
      Suara.tahan(1500);
    Dash.muat();
  }, 30000);
});

// ---- Reader serial: event dari backend ----
let _readerOk = null;
function tampilStatusReader(st) {
  const el = document.getElementById("status-reader");
  if (!el) return;
  const ok = !!(st && st.terhubung);
  bunyi(ok ? "ok" : "err", true);
  if (_readerOk !== null && _readerOk !== ok) {
    bunyi(ok ? "ok" : "err");
    toast(ok ? "Reader terhubung" : "Reader TERPUTUS", ok ? "ok" : "err", false);
  }
  _readerOk = ok;
  el.className = "pil-status " + (ok ? "ok" : "putus");
  el.querySelector(".teks").textContent = ok
    ? "Reader serial terhubung (" + st.port + ")"
    : "Reader serial terputus";
  el.title = ok
    ? "Reader serial siap menerima tap"
    : "Reader serial tidak terhubung. Reader HID (keyboard) tetap bisa dipakai.";
}

document.addEventListener("DOMContentLoaded", () => {
  if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn("card:tap", (uid) => {
      Suara.izinkan();
      if (Tap.handler) Tap.handler(uid);
    });
    window.runtime.EventsOn("reader:status", tampilStatusReader);
  }
  API.StatusReader()
    .then(tampilStatusReader)
    .catch(() => {});
});

// ---- Sidebar: ciutkan/bentangkan, diingat antar sesi ----
document.addEventListener("DOMContentLoaded", () => {
  const sb = document.getElementById("sidebar");
  const btn = document.getElementById("btn-collapse");
  if (!sb || !btn) return;
  const KUNCI = "mbg.sidebar.ciut";
  const terapkan = (ciut) => {
    sb.classList.toggle("ciut", ciut);
    btn.setAttribute("aria-expanded", String(!ciut));
    const t = ciut ? "Bentangkan sidebar" : "Ciutkan sidebar";
    btn.title = t;
    btn.setAttribute("aria-label", t);
  };
  let awal = false;
  try {
    awal = localStorage.getItem(KUNCI) === "1";
  } catch (e) {}
  terapkan(awal);
  btn.onclick = () => {
    const ciut = !sb.classList.contains("ciut");
    terapkan(ciut);
    try {
      localStorage.setItem(KUNCI, ciut ? "1" : "0");
    } catch (e) {}
  };
});

// ---- Tema: terang/gelap, awalnya mengikuti sistem, pilihan diingat ----
document.addEventListener("DOMContentLoaded", () => {
  const html = document.documentElement;
  const btn = document.getElementById("btn-tema");
  if (!btn) return;
  const KUNCI = "mbg.tema";
  const tersimpan = () => {
    try {
      return localStorage.getItem(KUNCI);
    } catch (e) {
      return null;
    }
  };
  const terapkan = (t) => {
    html.setAttribute("data-tema", t);
    const label = t === "gelap" ? "Ganti ke mode terang" : "Ganti ke mode gelap";
    btn.title = label;
    btn.setAttribute("aria-label", label);
  };
  terapkan(html.getAttribute("data-tema") || "terang");
  btn.onclick = () => {
    const baru = html.getAttribute("data-tema") === "gelap" ? "terang" : "gelap";
    html.classList.add("ganti-tema");
    terapkan(baru);
    setTimeout(() => html.classList.remove("ganti-tema"), 350);
    try {
      localStorage.setItem(KUNCI, baru);
    } catch (e) {}
  };
  // selama pengguna belum memilih sendiri, ikuti perubahan tema sistem
  matchMedia("(prefers-color-scheme: dark)").addEventListener("change", (e) => {
    if (!tersimpan()) terapkan(e.matches ? "gelap" : "terang");
  });
});
