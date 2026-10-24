// Binding Wails: window.go.<package>.<Type>.<Method>
const API = new Proxy(
  {},
  {
    get:
      (_, name) =>
      (...args) => {
        const svc = window.go && window.go.service && window.go.service.Service;
        if (!svc || !svc[name]) return Promise.reject(new Error("Binding belum siap: " + name));
        return svc[name](...args);
      },
  }
);

// =====================================================================
// SUARA: katalog nada terpusat. Setiap aksi punya bunyi sendiri.
// Format nada: [frekuensi, mulai(detik), durasi(detik), gelombang, volume, frekuensiAkhir?]
// =====================================================================
const POLA = {
  // umpan balik klik
  klik: [[1000, 0, 0.025, "square", 0.05]],
  utama: [[1400, 0, 0.03, "square", 0.06]],
  bahaya: [[260, 0, 0.045, "sawtooth", 0.08]],
  tab: [[720, 0, 0.03, "square", 0.06]],
  buka: [
    [520, 0, 0.04, "square", 0.07],
    [780, 0.05, 0.05, "square", 0.07],
  ],
  // kartu ditempel: dua bip cepat seperti scanner
  tap: [
    [1800, 0, 0.05, "square", 0.1],
    [1800, 0.07, 0.05, "square", 0.1],
  ],
  // hasil
  ok: [
    [880, 0, 0.07, "square", 0.12],
    [1320, 0.09, 0.1, "square", 0.12],
  ],
  info: [[660, 0, 0.08, "square", 0.1]],
  err: [
    [150, 0, 0.18, "sawtooth", 0.2],
    [110, 0.22, 0.28, "sawtooth", 0.2],
  ],
  peringatan: [
    [440, 0, 0.1, "sawtooth", 0.15],
    [440, 0.16, 0.1, "sawtooth", 0.15],
  ],
  hapus: [[600, 0, 0.25, "sawtooth", 0.15, 140]],
  // ompreng: naik = ambil, turun = kembali
  ambil: [
    [523, 0, 0.07, "square", 0.12],
    [659, 0.08, 0.07, "square", 0.12],
    [784, 0.16, 0.12, "square", 0.12],
  ],
  kembali: [
    [784, 0, 0.07, "square", 0.12],
    [659, 0.08, 0.07, "square", 0.12],
    [523, 0.16, 0.12, "square", 0.12],
  ],
  selesai: [
    [523, 0, 0.09, "square", 0.12],
    [659, 0.1, 0.09, "square", 0.12],
    [784, 0.2, 0.09, "square", 0.12],
    [1047, 0.3, 0.3, "square", 0.12],
  ],
};

const Suara = {
  volume: 1, // pengali global; naikkan (mis. 1.5) untuk ruangan ramai
  _ctx: null,
  _hingga: 0,
  // Tahan bunyi hasil muat data (ok, galat, info, dst.) selama berpindah tab atau menu.
  // Toast tetap tampil; hanya suaranya yang ditahan. Klik tab dan aksi pengguna tidak terpengaruh.
  tahan(ms = 900) {
    this._hingga = Math.max(this._hingga, performance.now() + ms);
  },
  // Izin suara hasil: hanya aksi pengguna (klik tombol, ubah pilihan, Enter, tap kartu)
  // yang membukanya; klik tab atau menu langsung mencabutnya.
  _izin: 0,
  izinkan(ms = 20000) {
    this._izin = performance.now() + ms;
  },
  cabut() {
    this._izin = 0;
  },
  get nyala() {
    try {
      return localStorage.getItem("mbg.suara") !== "mati";
    } catch (e) {
      return true;
    }
  },
  set nyala(v) {
    try {
      localStorage.setItem("mbg.suara", v ? "nyala" : "mati");
    } catch (e) {}
    document.documentElement.setAttribute("data-suara", v ? "nyala" : "mati");
  },
  ctx() {
    this._ctx = this._ctx || new (window.AudioContext || window.webkitAudioContext)();
    if (this._ctx.state === "suspended") this._ctx.resume();
    return this._ctx;
  },
};

function nada(ctx, freq, mulai, durasi, gelombang, vol, freqAkhir) {
  const o = ctx.createOscillator();
  const g = ctx.createGain();
  const t = ctx.currentTime + mulai;
  const v = Math.min(0.5, vol * Suara.volume);
  o.type = gelombang;
  o.frequency.setValueAtTime(freq, t);
  if (freqAkhir) o.frequency.exponentialRampToValueAtTime(freqAkhir, t + durasi);
  // serang langsung, potong cepat agar tidak ada bunyi "klik"
  g.gain.setValueAtTime(0.0001, t);
  g.gain.exponentialRampToValueAtTime(v, t + 0.004);
  g.gain.setValueAtTime(v, t + Math.max(0.005, durasi - 0.012));
  g.gain.exponentialRampToValueAtTime(0.0001, t + durasi);
  o.connect(g);
  g.connect(ctx.destination);
  o.start(t);
  o.stop(t + durasi + 0.01);
}

// Penegasan visual: modal bergetar saat ada galat
function getarKotak() {
  const m = document.getElementById("modal");
  const k = document.getElementById("modal-isi");
  if (!m || !k || m.classList.contains("hidden")) return;
  k.classList.remove("getar");
  void k.offsetWidth;
  k.classList.add("getar");
}

// nada hasil: butuh izin dari aksi pengguna
const TERTAHAN = new Set(["ok", "err", "info", "selesai", "peringatan", "hapus"]);
// nada aksi: memberi izin
const AKSI = new Set(["klik", "utama", "bahaya", "buka", "tap", "ambil", "kembali"]);

// paksa = true untuk peringatan sistem yang harus berbunyi walau tanpa aksi pengguna
// (mis. reader terputus). Tetap tunduk pada tombol bisukan.
function bunyi(tipe, paksa) {
  if (tipe === "err") getarKotak();
  if (tipe === "tab") Suara.cabut();
  else if (AKSI.has(tipe)) Suara.izinkan();
  if (!Suara.nyala) return;
  if (TERTAHAN.has(tipe) && !paksa) {
    const t = performance.now();
    if (t >= Suara._izin || t < Suara._hingga) return;
  }
  try {
    const ctx = Suara.ctx();
    (POLA[tipe] || POLA.info).forEach((n) => nada(ctx, ...n));
  } catch (e) {}
}

// Browser menahan audio sampai ada interaksi pertama: buka kuncinya sekali
["pointerdown", "keydown"].forEach((ev) =>
  document.addEventListener(
    ev,
    () => {
      try {
        Suara.ctx();
      } catch (e) {}
    },
    { once: true, capture: true }
  )
);

// Bunyi klik untuk SEMUA tombol, dipasang sekali di sini (fase capture)
document.addEventListener(
  "click",
  (e) => {
    const b = e.target.closest && e.target.closest("button");
    if (!b || b.disabled || b.dataset.senyap) return;
    if (b.id === "btn-suara") return;
    if (b.dataset.tab || b.dataset.sub || b.id === "r-btn-trx" || b.id === "r-btn-audit")
      bunyi("tab");
    else if (b.classList.contains("bahaya")) bunyi("bahaya");
    else if (b.classList.contains("utama")) bunyi("utama");
    else bunyi("klik");
  },
  true
);

// Mengubah pilihan (select, checkbox) dan menekan Enter juga aksi pengguna
document.addEventListener("change", () => Suara.izinkan(), true);
document.addEventListener(
  "keydown",
  (e) => {
    if (e.key === "Enter") Suara.izinkan();
  },
  true
);

// Tombol bisukan di topbar
document.addEventListener("DOMContentLoaded", () => {
  document.documentElement.setAttribute("data-suara", Suara.nyala ? "nyala" : "mati");
  const btn = document.getElementById("btn-suara");
  if (!btn) return;
  const label = () => {
    const t = Suara.nyala ? "Matikan suara" : "Nyalakan suara";
    btn.title = t;
    btn.setAttribute("aria-label", t);
  };
  label();
  btn.onclick = () => {
    Suara.nyala = !Suara.nyala;
    label();
    if (Suara.nyala) bunyi("ok");
  };
});

// Jaring pengaman: galat yang tidak ditangkap tidak lagi diam
let _rejT = 0,
  _rejM = "";
window.addEventListener("unhandledrejection", (ev) => {
  const m = String((ev.reason && ev.reason.message) || ev.reason || "Terjadi galat");
  const now = Date.now();
  if (m === _rejM && now - _rejT < 10000) return;
  _rejM = m;
  _rejT = now;
  toast(m, "err");
});

// suara: nama nada, atau false untuk senyap (bawaan mengikuti tipe)
let _toastT, _toastH;
function toast(teks, tipe = "info", suara = tipe) {
  const el = document.getElementById("toast");
  clearTimeout(_toastT);
  clearTimeout(_toastH);
  el.textContent = teks;
  el.className = tipe;
  if (suara !== false) bunyi(suara);
  _toastT = setTimeout(() => {
    el.className = tipe + " keluar";
    _toastH = setTimeout(() => (el.className = "hidden"), 180);
  }, 2500);
}

// Salin teks ke clipboard; cadangan execCommand bila Clipboard API ditolak
async function salin(teks) {
  try {
    await navigator.clipboard.writeText(teks);
  } catch (e) {
    const ta = document.createElement("textarea");
    ta.value = teks;
    ta.setAttribute("readonly", "");
    ta.style.position = "fixed";
    ta.style.opacity = "0";
    document.body.appendChild(ta);
    ta.select();
    try {
      document.execCommand("copy");
    } finally {
      document.body.removeChild(ta);
    }
  }
}
