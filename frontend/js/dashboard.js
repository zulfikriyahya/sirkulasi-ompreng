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
  _prev: null, // status per kelas pada muat sebelumnya, untuk kedipan baris
  _semua: null, // apakah semua kelas selesai pada muat sebelumnya
  _aksi: false, // true hanya bila muat dipicu aksi pengguna (submit ambil/kembali)

  async muat() {
    const dariAksi = this._aksi;
    this._aksi = false;
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
    const sekarang = {};
    st.baris.forEach((b, i) => {
      if (b.status === "SELESAI") hit.selesai++;
      else if (b.status === "AMBIL") hit.ambil++;
      else if (b.status === "KURANG" || b.status === "BELUM_KEMBALI") hit.kurang++;
      else if (b.status === "BELUM_AMBIL") hit.belum++;
      sekarang[b.nama] = b.status;
      const tr = document.createElement("tr");
      if (b.merah) tr.className = "merah";
      else if (b.status === "SELESAI") tr.className = "selesai";
      // baris yang statusnya baru berubah dikedipkan
      if (this._prev && this._prev[b.nama] !== undefined && this._prev[b.nama] !== b.status)
        tr.classList.add("berubah");
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
    this._prev = sekarang;
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

    // fanfare hanya untuk aksi pengguna; muat saat pindah tab atau segar berkala tidak bersuara
    const semua = hit.total > 0 && hit.selesai === hit.total;
    if (dariAksi && this._semua === false && semua) setTimeout(() => bunyi("selesai"), 450);
    this._semua = semua;

    Dash.terapkanHari(st);
  },

  // ---- popup ----
  _tm: null,
  tutupModal() {
    const m = document.getElementById("modal");
    if (m.classList.contains("hidden")) return;
    m.classList.add("keluar");
    clearTimeout(this._tm);
    this._tm = setTimeout(() => (m.className = "hidden"), 170);
  },
  // senyap = true bila pemanggil sudah memainkan nada sendiri
  bukaModal(html, senyap) {
    clearTimeout(this._tm);
    const m = document.getElementById("modal");
    const k = document.getElementById("modal-isi");
    k.className = "kotak"; // buang sisa kelas getar / bahaya dari modal sebelumnya
    k.innerHTML = html;
    m.className = "";
    if (!senyap) bunyi("buka");
  },

  // Modal konfirmasi pengganti confirm() bawaan browser.
  // Mengembalikan null bila dibatalkan, atau { alasan } bila disetujui.
  // Opsi: mode, judul, label, nama, detail, pesan, aksi, alasan (true = tampilkan kolom alasan)
  konfirmasi(o) {
    return new Promise((selesai) => {
      bunyi("peringatan");
      Tap.handler = () => toast("Selesaikan konfirmasi dulu", "info", "peringatan");
      this.bukaModal(
        `
        <span class="mode bahaya">${esc(o.mode || "Konfirmasi")}</span>
        <h2>${esc(o.judul)}</h2>
        <div class="orang"><span class="peran">${esc(o.label || "Data")}</span><b>${esc(o.nama)}</b>${o.detail ? `<span class="kecil">${esc(o.detail)}</span>` : ""}</div>
        ${o.pesan ? `<p>${esc(o.pesan)}</p>` : ""}
        ${o.alasan ? `<label>Alasan (opsional)<input id="kf-alasan" maxlength="120" placeholder="mis. kelas digabung"></label>` : ""}
        <div class="aksi"><button id="kf-batal">Batal</button><button id="kf-ok" class="bahaya-solid">${esc(o.aksi || "Ya")}</button></div>`,
        true
      );
      document.getElementById("modal-isi").classList.add("bahaya");
      const alasan = document.getElementById("kf-alasan");
      let sudah = false;
      const akhir = (hasil) => {
        if (sudah) return;
        sudah = true;
        document.removeEventListener("keydown", tombol, true);
        this.tutupModal();
        pasangHandler(); // pulihkan penangan tap sesuai tab aktif
        selesai(hasil);
      };
      const tombol = (e) => {
        if (e.key === "Escape") {
          e.preventDefault();
          akhir(null);
        }
      };
      document.addEventListener("keydown", tombol, true);
      document.getElementById("kf-batal").onclick = () => akhir(null);
      document.getElementById("kf-ok").onclick = () =>
        akhir({ alasan: alasan ? alasan.value.trim() : "" });
      if (alasan)
        alasan.onkeydown = (e) => {
          if (e.key === "Enter") document.getElementById("kf-ok").click();
        };
      // fokus awal di Batal supaya Enter tidak sengaja menyetujui aksi berbahaya
      document.getElementById("kf-batal").focus();
    });
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
    // nada naik = ambil, nada turun = kembali
    bunyi(ambil ? "ambil" : "kembali");
    this.bukaModal(
      `
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
      <div class="aksi"><button id="in-batal">Batal</button><button id="in-ok">Submit</button></div>`,
      true
    );
    const inp = document.getElementById("in-jumlah");
    inp.focus();
    const salah = (teks) => {
      document.getElementById("in-galat").textContent = teks;
      bunyi("err");
      inp.focus();
      inp.select();
    };
    const kirim = async () => {
      const val = inp.value.trim();
      if (!/^\d+$/.test(val)) return salah("Isi dengan bilangan bulat");
      const n = parseInt(val, 10);
      if (n < 1) return salah("Jumlah minimal 1");
      if (susulan && n > r.sisa) return salah(`Jumlah susulan harus 1 sampai ${r.sisa}`);
      try {
        if (ambil) await API.SubmitAmbil(r.kelas_id, n);
        else await API.SubmitKembali(r.kelas_id, n);
        this.tutupModal();
        toast(
          `${r.nama_kelas}: ${ambil ? "pengambilan" : susulan ? "pengembalian susulan" : "pengembalian"} tercatat`,
          "ok"
        );
        this._aksi = true;
        this.muat();
      } catch (err) {
        salah(String(err.message || err));
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
    bunyi("peringatan");
    this.bukaModal(
      `
      <span class="mode konfirmasi">Konfirmasi</span>
      <h2>${esc(r.nama_kelas)}</h2>
      <p>Kelas <b>${esc(r.nama_kelas)}</b> baru mengambil ${r.menit_sejak} menit lalu. Pengembali: <b>${esc(r.nama_siswa || "-")}</b>. Tetap proses pengembalian?</p>
      <div class="aksi"><button id="kf-tidak">Tidak</button><button id="kf-ya">Ya</button></div>`,
      true
    );
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
  bunyi("tap"); // kartu terbaca: umpan balik seketika sebelum backend menjawab
  try {
    const r = await API.ProsesTap(uid);
    if (r.aksi === "ABAIKAN") return;
    if (r.aksi === "PESAN") return toast(r.pesan, r.tipe);
    if (r.nama_siswa) toast(r.nama_siswa + " (" + r.nama_kelas + ")", "info", false);
    if (r.aksi === "AMBIL") Dash.popupJumlah(r, "AMBIL");
    else if (r.aksi === "KEMBALI") Dash.popupJumlah(r, "KEMBALI");
    else if (r.aksi === "KEMBALI_SUSULAN") Dash.popupJumlah(r, "SUSULAN");
    else if (r.aksi === "KONFIRMASI_KEMBALI") Dash.popupKonfirmasi(r);
  } catch (err) {
    toast(String(err.message || err), "err");
  }
}
