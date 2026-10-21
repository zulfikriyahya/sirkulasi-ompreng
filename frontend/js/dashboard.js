const Dash = {
    async muat() {
        const st = await API.StatusHariIni();
        document.getElementById("info-tanggal").textContent = st.tanggal_indo + (st.ditutup ? " (DITUTUP)" : "");
        const tb = document.querySelector("#tabel-hari tbody");
        tb.innerHTML = "";
        const label = { BELUM_AMBIL: "Belum ambil", AMBIL: "Sudah ambil", SELESAI: "Selesai", BELUM_KEMBALI: "Belum kembali", TIDAK_AMBIL: "Tidak mengambil" };
        let selesai = 0;
        st.baris.forEach((b, i) => {
            if (b.status === "SELESAI") selesai++;
            const tr = document.createElement("tr");
            if (b.merah) tr.className = "merah"; else if (b.status === "SELESAI") tr.className = "selesai";
            const v = x => (x === null || x === undefined) ? "-" : x;
            [i + 1, b.nama, b.nama_ketua, label[b.status] || b.status, v(b.jumlah_ambil), v(b.jumlah_kembali),
            b.waktu_ambil || "-", b.waktu_kembali || "-", b.keterangan].forEach(t => {
                const td = document.createElement("td"); td.textContent = t; tr.appendChild(td);
            });
            tb.appendChild(tr);
        });
        document.getElementById("ringkasan").textContent = `${selesai} dari ${st.baris.length} kelas selesai.`;
        Dash.terapkanHari(st);
    },

    // ---- popup ----
    tutupModal() { document.getElementById("modal").className = "hidden"; },
    bukaModal(html) { const m = document.getElementById("modal"); document.getElementById("modal-isi").innerHTML = html; m.className = ""; },

    popupJumlah(r, mode) {
        const ambil = mode === "AMBIL";
        this.bukaModal(`
      <h2>${ambil ? "Pengambilan" : "Pengembalian"} - ${esc(r.nama_kelas)}</h2>
      <p>Ketua: ${esc(r.nama_ketua)}${ambil ? "" : ` &middot; Diambil: <b>${r.jumlah_ambil}</b> ompreng`}</p>
      <input id="in-jumlah" type="number" min="${ambil ? 1 : 0}" ${ambil ? "" : `max="${r.jumlah_ambil}"`} placeholder="Jumlah ${ambil ? "diambil" : "dikembalikan"}">
      <div class="galat" id="in-galat"></div>
      <div class="aksi"><button id="in-batal">Batal</button><button id="in-ok">Submit</button></div>`);
        const inp = document.getElementById("in-jumlah"); inp.focus();
        const kirim = async () => {
            const val = inp.value.trim();
            if (!/^\d+$/.test(val)) { document.getElementById("in-galat").textContent = "Isi dengan bilangan bulat"; return; }
            try {
                if (ambil) await API.SubmitAmbil(r.kelas_id, parseInt(val, 10));
                else await API.SubmitKembali(r.kelas_id, parseInt(val, 10));
                this.tutupModal();
                toast(`${r.nama_kelas}: ${ambil ? "pengambilan" : "pengembalian"} tercatat`, "ok");
                this.muat();
            } catch (err) {
                document.getElementById("in-galat").textContent = String(err.message || err);
                bunyi("err");
            }
        };
        document.getElementById("in-ok").onclick = kirim;
        inp.onkeydown = e => { if (e.key === "Enter") kirim(); };
        document.getElementById("in-batal").onclick = async () => { await API.BatalInput(); this.tutupModal(); };
    },

    popupKonfirmasi(r) {
        this.bukaModal(`
      <h2>Konfirmasi</h2>
      <p>Kelas <b>${esc(r.nama_kelas)}</b> baru mengambil ${r.menit_sejak} menit lalu. Tetap proses pengembalian?</p>
      <div class="aksi"><button id="kf-tidak">Tidak</button><button id="kf-ya">Ya</button></div>`);
        document.getElementById("kf-tidak").onclick = async () => { await API.BatalInput(); this.tutupModal(); };
        document.getElementById("kf-ya").onclick = () => this.popupJumlah(r, "KEMBALI");
    },
};

function esc(s) { return String(s).replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c])); }

async function prosesTap(uid) {
    try {
        const r = await API.ProsesTap(uid);
        if (r.aksi === "ABAIKAN") return;
        if (r.aksi === "PESAN") return toast(r.pesan, r.tipe);
        bunyi("ok");
        if (r.aksi === "AMBIL") Dash.popupJumlah(r, "AMBIL");
        else if (r.aksi === "KEMBALI") Dash.popupJumlah(r, "KEMBALI");
        else if (r.aksi === "KONFIRMASI_KEMBALI") Dash.popupKonfirmasi(r);
    } catch (err) { toast(String(err.message || err), "err"); }
}