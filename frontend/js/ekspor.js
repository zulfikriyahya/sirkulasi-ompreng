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
                [i + 1, b.kelas, b.ketua, b.tanggal, b.waktu_ambil, b.waktu_kembali, b.jumlah_ambil, b.jumlah_kembali, b.keterangan].forEach(t => {
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
});