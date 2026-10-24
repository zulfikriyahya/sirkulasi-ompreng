const LABEL_STATUS = {
  BELUM_AMBIL: "Belum ambil",
  AMBIL: "Sudah ambil",
  KURANG: "Kurang",
  SELESAI: "Selesai",
  BELUM_KEMBALI: "Belum kembali",
  TIDAK_AMBIL: "Tidak mengambil",
};

// Lencana status berwarna; dipakai juga oleh riwayat.js.
function lencana(status) {
  const s = document.createElement("span");
  s.className = "lencana " + String(status).toLowerCase().replace(/_/g, "-");
  s.textContent = LABEL_STATUS[status] || status;
  return s;
}

const Dash = {
  async muat() {
    const st = await API.StatusHariIni();
    document.getElementById("info-tanggal").textContent =
      st.tanggal_indo + (st.ditutup ? " (DITUTUP)" : "");
    const tb = document.querySelector("#tabel-hari tbody");
    tb.innerHTML = "";
    const hit = { total: st.baris.length, selesai: 0, ambil: 0, kurang: 0, belum: 0 };
    const v = (x) => (x === null || x === undefined ? "-" : x);
    const sel = (teks, kelas) => {
      const td = document.createElement("td");
      if (kelas) td.className = kelas;
      td.textContent = teks;
      return td;
    };
    st.baris.forEach((b, i) => {
      if (b.status === "SELESAI") hit.selesai++;
      else if (b.status === "AMBIL") hit.ambil++;
      else if (b.status === "KURANG" || b.status === "BELUM_KEMBALI") hit.kurang++;
      else if (b.status === "BELUM_AMBIL") hit.belum++;
      const tr = document.createElement("tr");
      if (b.merah) tr.className = "merah";
      else if (b.status === "SELESAI") tr.className = "selesai";
      const tdStatus = document.createElement("td");
      tdStatus.appendChild(lencana(b.status));
      tr.append(
        sel(i + 1, "num"),
        sel(b.nama),
        sel(b.nama_ketua),
        tdStatus,
        sel(v(b.jumlah_ambil), "num"),
        sel(v(b.jumlah_kembali), "num"),
        sel(b.waktu_ambil || "-", "waktu"),
        sel(b.waktu_kembali || "-", "waktu"),
        sel(b.keterangan)
      );
      tb.appendChild(tr);
    });
    if (!st.baris.length)
      tb.innerHTML =
        '<tr><td colspan="9">Belum ada kelas. Tambahkan di Master Data &gt; Kelas &amp; Ketua.</td></tr>';

    const persen = hit.total ? Math.round((hit.selesai * 100) / hit.total) : 0;
    document.getElementById("ringkasan").textContent =
      `${hit.selesai} dari ${hit.total} kelas selesai (${persen}%)`;
    document.getElementById("progres-isi").style.width = persen + "%";
    const kartu = (kelas, label, nilai, ket) =>
      `<div class="stat ${kelas}"><div class="label">${label}</div><div class="nilai">${nilai}</div><div class="ket">${ket}</div></div>`;
    document.getElementById("statistik").innerHTML =
      kartu("", "Kelas", hit.total, `${hit.belum} belum mengambil`) +
      kartu("biru", "Sudah ambil", hit.ambil, "menunggu pengembalian") +
      kartu("hijau", "Selesai", hit.selesai, "dikembalikan lengkap") +
      kartu(hit.kurang ? "merah" : "", "Kurang", hit.kurang, "perlu ditindaklanjuti");
    Dash.terapkanHari(st);
  },

  // ---- popup ----
  tutupModal() {
    document.getElementById("modal").className = "hidden";
  },
  bukaModal(html) {
    const m = document.getElementById("modal");
    document.getElementById("modal-isi").innerHTML = html;
    m.className = "";
  },

  // mode: AMBIL | KEMBALI | SUSULAN
  popupJumlah(r, mode) {
    const ambil = mode === "AMBIL";
    const susulan = mode === "SUSULAN";
    const sudah = (r.jumlah_ambil || 0) - (r.sisa || 0);
    const judul = ambil ? "Pengambilan" : susulan ? "Pengembalian susulan" : "Pengembalian";
    const kelasMode = ambil ? "" : susulan ? "susulan" : "kembali";
    const tiles = ambil
      ? ""
      : susulan
        ? `<div class="tiles"><div class="tile"><span>Diambil</span><b>${r.jumlah_ambil}</b></div><div class="tile"><span>Sudah kembali</span><b>${sudah}</b></div><div class="tile sisa"><span>Sisa</span><b>${r.sisa}</b></div></div>`
        : `<div class="tiles"><div class="tile"><span>Diambil</span><b>${r.jumlah_ambil}</b></div></div>`;
    const min = 1;
    const max = susulan ? r.sisa : r.jumlah_ambil;
    const nama = ambil ? "diambil" : susulan ? "susulan" : "dikembalikan";
    this.bukaModal(`
      <span class="mode ${kelasMode}">${judul}</span>
      <h2>${esc(r.nama_kelas)}</h2>
      <p>Ketua: ${esc(r.nama_ketua)}</p>
      ${tiles}
      <div class="orang"><span class="peran">${ambil ? "Pengambil" : "Pengembali"}</span><b>${esc(r.nama_siswa || "(kartu tanpa nama)")}</b>${r.nisn ? `<span class="kecil">NISN ${esc(r.nisn)}</span>` : ""}</div>
      <label>Jumlah ${nama}
        <input id="in-jumlah" type="number" inputmode="numeric" autocomplete="off" min="${min}" ${ambil ? "" : `max="${max}"`} placeholder="Jumlah">
      </label>
      ${ambil ? "" : `<span class="kecil">Rentang ${min} sampai ${max}</span>`}
      <div class="galat" id="in-galat"></div>
      <div class="aksi"><button id="in-batal">Batal</button><button id="in-ok">Submit</button></div>`);
    const inp = document.getElementById("in-jumlah");
    inp.focus();
    const kirim = async () => {
      const val = inp.value.trim();
      if (!/^\d+$/.test(val)) {
        document.getElementById("in-galat").textContent = "Isi dengan bilangan bulat";
        return;
      }
      const n = parseInt(val, 10);
      if (n < 1) {
        document.getElementById("in-galat").textContent = "Jumlah minimal 1";
        return;
      }
      if (susulan && n > r.sisa) {
        document.getElementById("in-galat").textContent = `Jumlah susulan harus 1 sampai ${r.sisa}`;
        return;
      }
      try {
        if (ambil) await API.SubmitAmbil(r.kelas_id, n);
        else await API.SubmitKembali(r.kelas_id, n);
        this.tutupModal();
        toast(
          `${r.nama_kelas}: ${ambil ? "pengambilan" : susulan ? "pengembalian susulan" : "pengembalian"} tercatat`,
          "ok"
        );
        this.muat();
      } catch (err) {
        document.getElementById("in-galat").textContent = String(err.message || err);
        bunyi("err");
      }
    };
    document.getElementById("in-ok").onclick = kirim;
    inp.onkeydown = (e) => {
      if (e.key === "Enter") kirim();
    };
    document.getElementById("in-batal").onclick = async () => {
      await API.BatalInput();
      this.tutupModal();
    };
  },

  popupKonfirmasi(r) {
    this.bukaModal(`
      <span class="mode konfirmasi">Konfirmasi</span>
      <h2>${esc(r.nama_kelas)}</h2>
      <p>Kelas <b>${esc(r.nama_kelas)}</b> baru mengambil ${r.menit_sejak} menit lalu. Pengembali: <b>${esc(r.nama_siswa || "-")}</b>. Tetap proses pengembalian?</p>
      <div class="aksi"><button id="kf-tidak">Tidak</button><button id="kf-ya">Ya</button></div>`);
    document.getElementById("kf-tidak").onclick = async () => {
      await API.BatalInput();
      this.tutupModal();
    };
    document.getElementById("kf-ya").onclick = () => this.popupJumlah(r, "KEMBALI");
  },
};

function esc(s) {
  return String(s).replace(
    /[&<>"']/g,
    (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]
  );
}

async function prosesTap(uid) {
  try {
    const r = await API.ProsesTap(uid);
    if (r.aksi === "ABAIKAN") return;
    if (r.aksi === "PESAN") return toast(r.pesan, r.tipe);
    bunyi("ok");
    if (r.nama_siswa) toast(r.nama_siswa + " (" + r.nama_kelas + ")", "info");
    if (r.aksi === "AMBIL") Dash.popupJumlah(r, "AMBIL");
    else if (r.aksi === "KEMBALI") Dash.popupJumlah(r, "KEMBALI");
    else if (r.aksi === "KEMBALI_SUSULAN") Dash.popupJumlah(r, "SUSULAN");
    else if (r.aksi === "KONFIRMASI_KEMBALI") Dash.popupKonfirmasi(r);
  } catch (err) {
    toast(String(err.message || err), "err");
  }
}
