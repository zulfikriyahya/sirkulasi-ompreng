const _r = (id) => document.getElementById(id);

// Tombol cetak/tutup/buka di Beranda mengikuti status hari
Dash.terapkanHari = function (st) {
  _r("btn-cetak").disabled = !st.boleh_cetak;
  _r("btn-tutup").classList.toggle("hidden", st.ditutup || st.semua_selesai);
  _r("btn-buka").classList.toggle("hidden", !st.ditutup);
};

const Riwayat = {
  pilih(s) {
    _r("r-trx").classList.toggle("hidden", s !== "trx");
    _r("r-audit").classList.toggle("hidden", s !== "audit");
    _r("r-btn-trx").classList.toggle("aktif", s === "trx");
    _r("r-btn-audit").classList.toggle("aktif", s === "audit");
    if (s === "trx") this.muatTrx();
    else this.muatAudit();
  },

  async muat() {
    const st = await API.StatusHariIni();
    for (const id of ["rf-dari", "rf-sampai", "ra-dari", "ra-sampai"])
      if (!_r(id).value) _r(id).value = st.tanggal;
    const ks = await API.ListKelas(true);
    const sel = _r("rf-kelas"),
      cur = sel.value;
    sel.innerHTML =
      '<option value="0">Semua kelas</option>' +
      ks.map((k) => `<option value="${k.id}">${esc(k.nama)}</option>`).join("");
    sel.value = cur || "0";
    this.muatTrx();
  },

  async muatTrx() {
    try {
      const rows = await API.ListTransaksi({
        dari: _r("rf-dari").value,
        sampai: _r("rf-sampai").value,
        kelas_id: parseInt(_r("rf-kelas").value || "0", 10),
        status: _r("rf-status").value,
      });
      const tb = document.querySelector("#tabel-riwayat tbody");
      tb.innerHTML = "";
      const v = (x) => (x === null || x === undefined ? "-" : x);
      rows.forEach((r) => {
        const tr = document.createElement("tr");
        if (r.merah) tr.className = "merah";
        [
          r.tanggal,
          r.nama_kelas,
          r.nama_ketua,
          r.nama_ambil || "-",
          r.nama_kembali || "-",
          r.status,
          v(r.jumlah_ambil),
          v(r.jumlah_kembali),
          (r.waktu_ambil || "-") + " / " + (r.waktu_kembali || "-"),
          r.keterangan,
        ].forEach((t, i) => {
          const td = document.createElement("td");
          if (i === 5) td.appendChild(lencana(t));
          else td.textContent = t;
          if (i === 6 || i === 7) td.className = "num";
          else if (i === 8) td.className = "waktu";
          tr.appendChild(td);
        });
        if (r.pengembalian && r.pengembalian.length > 1) {
          const d = document.createElement("div");
          d.className = "kecil";
          d.textContent = r.pengembalian
            .map((p) => `${p.waktu} • ${p.jumlah} • ${p.nama_siswa || "-"}`)
            .join("  |  ");
          tr.lastElementChild.appendChild(d);
        }
        const td = document.createElement("td");
        td.className = "aksi-sel";
        const wadah = document.createElement("div");
        wadah.className = "aksi-baris";
        const be = document.createElement("button");
        be.textContent = "Ubah";
        be.onclick = () => this.dialogUbah(r);
        const bh = document.createElement("button");
        bh.textContent = "Hapus";
        bh.className = "bahaya";
        bh.onclick = () => this.dialogHapus(r);
        wadah.append(be, bh);
        td.appendChild(wadah);
        tr.appendChild(td);
        tb.appendChild(tr);
      });
      if (!rows.length) tb.innerHTML = '<tr><td colspan="11">Tidak ada data.</td></tr>';
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },

  async muatAudit() {
    try {
      const rows = await API.ListAudit({
        dari: _r("ra-dari").value,
        sampai: _r("ra-sampai").value,
        entitas: _r("ra-entitas").value,
        limit: 300,
      });
      const tb = document.querySelector("#tabel-audit tbody");
      tb.innerHTML = "";
      rows.forEach((a) => {
        const tr = document.createElement("tr");
        [
          a.waktu.replace("T", " ").slice(0, 19),
          a.aksi,
          a.entitas,
          a.entitas_id,
          a.nilai_lama,
          a.nilai_baru,
          a.alasan,
        ].forEach((t) => {
          const td = document.createElement("td");
          td.textContent = t;
          td.title = t;
          tr.appendChild(td);
        });
        tb.appendChild(tr);
      });
      if (!rows.length) tb.innerHTML = '<tr><td colspan="7">Tidak ada data.</td></tr>';
    } catch (e) {
      toast(String(e.message || e), "err");
    }
  },

  dialogUbah(r) {
    const n = (x) => (x === null || x === undefined ? "" : x);
    const opsi = ["AMBIL", "KURANG", "SELESAI", "BELUM_KEMBALI", "TIDAK_AMBIL"]
      .map((s) => `<option ${s === r.status ? "selected" : ""}>${s}</option>`)
      .join("");
    Dash.bukaModal(`
          <h2>Ubah transaksi</h2>
          <p>${esc(r.nama_kelas)} &middot; ${esc(r.tanggal)}</p>
          <label>Status <select id="e-status">${opsi}</select></label>
          <div class="dua-kolom">
            <label>Jumlah ambil <input id="e-ja" type="number" min="1" value="${n(r.jumlah_ambil)}"></label>
            <label>Jumlah kembali <input id="e-jk" type="number" min="0" value="${n(r.jumlah_kembali)}"></label>
            <label>Waktu ambil (HH:MM) <input id="e-wa" maxlength="5" value="${esc(r.waktu_ambil)}"></label>
            <label>Waktu kembali (HH:MM) <input id="e-wk" maxlength="5" value="${esc(r.waktu_kembali)}"></label>
          </div>
          <label>Catatan <input id="e-cat" value="${esc(r.catatan)}"></label>
          <label>Alasan perubahan (wajib) <input id="e-alasan"></label>
          <div class="galat" id="e-galat"></div>
          <div class="aksi"><button id="e-batal">Batal</button><button id="e-ok">Simpan</button></div>`);
    const num = (id) => {
      const t = _r(id).value.trim();
      return t === "" ? null : parseInt(t, 10);
    };
    _r("e-batal").onclick = () => Dash.tutupModal();
    _r("e-ok").onclick = async () => {
      try {
        await API.UbahTransaksi({
          id: r.id,
          status: _r("e-status").value,
          jumlah_ambil: num("e-ja"),
          jumlah_kembali: num("e-jk"),
          waktu_ambil: _r("e-wa").value.trim(),
          waktu_kembali: _r("e-wk").value.trim(),
          catatan: _r("e-cat").value,
          alasan: _r("e-alasan").value,
        });
        Dash.tutupModal();
        toast("Perubahan tersimpan", "ok");
        this.muatTrx();
      } catch (e) {
        _r("e-galat").textContent = String(e.message || e);
        bunyi("err");
      }
    };
  },

  dialogHapus(r) {
    Dash.bukaModal(`
          <h2>Hapus transaksi?</h2>
          <p>${esc(r.nama_kelas)} - ${esc(r.tanggal)} (${esc(r.status)}). Tindakan ini tercatat di audit log.</p>
          <label>Alasan penghapusan (wajib) <input id="h-alasan"></label>
          <div class="galat" id="h-galat"></div>
          <div class="aksi"><button id="h-batal">Batal</button><button id="h-ok">Hapus</button></div>`);
    _r("h-batal").onclick = () => Dash.tutupModal();
    _r("h-ok").onclick = async () => {
      try {
        await API.HapusTransaksi(r.id, _r("h-alasan").value);
        Dash.tutupModal();
        toast("Transaksi dihapus", "ok", "hapus");
        this.muatTrx();
      } catch (e) {
        _r("h-galat").textContent = String(e.message || e);
        bunyi("err");
      }
    };
  },
};

document.addEventListener("DOMContentLoaded", () => {
  _r("r-btn-trx").onclick = () => Riwayat.pilih("trx");
  _r("r-btn-audit").onclick = () => Riwayat.pilih("audit");
  _r("rf-cari").onclick = () => Riwayat.muatTrx();
  _r("ra-cari").onclick = () => Riwayat.muatAudit();

  _r("btn-tutup").onclick = () => {
    Dash.bukaModal(`
          <h2>Tutup hari dengan catatan</h2>
          <p>Kelas yang sudah mengambil tapi belum mengembalikan menjadi <b>Belum kembali</b>. Kelas yang tidak mengambil dicatat <b>Tidak mengambil</b> (tidak masuk laporan).</p>
          <label>Catatan <input id="t-cat" placeholder="mis. kelas VII C tidak masuk"></label>
          <div class="galat" id="t-galat"></div>
          <div class="aksi"><button id="t-batal">Batal</button><button id="t-ok">Tutup hari</button></div>`);
    _r("t-batal").onclick = () => Dash.tutupModal();
    _r("t-ok").onclick = async () => {
      try {
        await API.TutupHari(_r("t-cat").value);
        Dash.tutupModal();
        toast("Hari ditutup", "ok", "selesai");   
        Dash.muat();
      } catch (e) {
        _r("t-galat").textContent = String(e.message || e);
        bunyi("err");
      }
    };
  };

  _r("btn-buka").onclick = () => {
    Dash.bukaModal(`
          <h2>Buka kembali hari ini?</h2>
          <p>Status otomatis dikembalikan: Belum kembali menjadi Sudah ambil, dan baris Tidak mengambil dihapus.</p>
          <label>Alasan (wajib) <input id="b-alasan"></label>
          <div class="galat" id="b-galat"></div>
          <div class="aksi"><button id="b-batal">Batal</button><button id="b-ok">Buka</button></div>`);
    _r("b-batal").onclick = () => Dash.tutupModal();
    _r("b-ok").onclick = async () => {
      try {
        const st = await API.StatusHariIni();
        await API.BukaHari(st.tanggal, _r("b-alasan").value);
        Dash.tutupModal();
        toast("Hari dibuka kembali", "ok");
        Dash.muat();
      } catch (e) {
        _r("b-galat").textContent = String(e.message || e);
        bunyi("err");
      }
    };
  };
});
