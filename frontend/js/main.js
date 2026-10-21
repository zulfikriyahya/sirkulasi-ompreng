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
});