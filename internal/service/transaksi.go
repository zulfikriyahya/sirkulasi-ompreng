package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mbg/internal/model"
	"mbg/internal/reader"
	"mbg/internal/timeutil"
)

func itoa(i int64) string { return strconv.FormatInt(i, 10) }

const (
	debounceTap   = 1 * time.Second
	cooldownKelas = 10 * time.Second
	// cooldownSusulan: setelah pengembalian yang masih kurang (status KURANG), cooldown kelas
	// dipendekkan supaya teman sekelas yang menyusul membawa sisa ompreng tidak ditolak
	// "Baru saja tercatat" selama 30 detik. 5 detik tetap menahan tap ganda dari kartu yang menempel.
	cooldownSusulan = 5 * time.Second
)

func (s *Service) jedaMenit() int {
	var m int
	if err := s.db.QueryRow(`SELECT jeda_minimum_menit FROM pengaturan WHERE id=1`).Scan(&m); err != nil {
		return 10
	}
	return m
}

func pesan(tipe, teks string) model.TapResult {
	return model.TapResult{Aksi: "PESAN", Tipe: tipe, Pesan: teks}
}

// ProsesTap: satu pintu untuk semua sumber (serial, HID, manual, simulator).
func (s *Service) ProsesTap(rawUID string) model.TapResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	uid, ok := reader.NormalisasiUID(rawUID)
	if !ok {
		return pesan("err", "UID tidak valid")
	}
	now := s.now()

	if s.pending != nil {
		return pesan("info", "Selesaikan input kelas "+s.pending.Nama+" dulu")
	}
	if uid == s.lastUID && now.Sub(s.lastUIDAt) < debounceTap {
		return model.TapResult{Aksi: "ABAIKAN"}
	}
	s.lastUID, s.lastUIDAt = uid, now

	var k model.Kelas
	var kartuAktif, kelasAktif int
	var nisn, namaSiswa string
	err := s.db.QueryRow(`SELECT k.id,k.nama,k.nama_ketua,c.aktif,k.aktif,COALESCE(c.nisn,''),COALESCE(c.nama_siswa,'') FROM kartu c JOIN kelas k ON k.id=c.kelas_id WHERE c.uid=?`, uid).
		Scan(&k.ID, &k.Nama, &k.NamaKetua, &kartuAktif, &kelasAktif, &nisn, &namaSiswa)
	if err != nil || kartuAktif != 1 || kelasAktif != 1 {
		return pesan("err", "Kartu tidak terdaftar")
	}

	tgl := timeutil.Tanggal(now)
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n)
	if n > 0 {
		return pesan("err", "Hari ini sudah ditutup")
	}

	base := model.TapResult{KelasID: k.ID, NamaKelas: k.Nama, NamaKetua: k.NamaKetua, Tipe: "ok", NISN: nisn, NamaSiswa: namaSiswa}
	var status, waktuAmbil string
	var jAmbil, jKembali int
	adaTrx := true
	err = s.db.QueryRow(`SELECT status,COALESCE(waktu_ambil,''),COALESCE(jumlah_ambil,0),COALESCE(jumlah_kembali,0) FROM transaksi WHERE tanggal=? AND kelas_id=?`, tgl, k.ID).
		Scan(&status, &waktuAmbil, &jAmbil, &jKembali)
	if errors.Is(err, sql.ErrNoRows) {
		adaTrx = false
	} else if err != nil {
		return pesan("err", "Kesalahan database")
	}

	// Penolakan final ditempatkan sebelum cooldown agar pesannya selalu jelas,
	// untuk kartu siswa mana pun di kelas yang sama.
	if adaTrx && status == "SELESAI" {
		if jKembali >= jAmbil {
			return pesan("info", "Kelas "+k.Nama+" sudah mengembalikan ompreng dengan lengkap")
		}
		return pesan("info", "Kelas "+k.Nama+" sudah selesai hari ini") // data lama (SELESAI kurang) dianggap final
	}
	if adaTrx && status != "AMBIL" && status != "KURANG" {
		return pesan("info", "Kelas "+k.Nama+" sudah selesai hari ini")
	}

	if t, ok := s.lastSubmit[k.ID]; ok && now.Sub(t) < cooldownKelas {
		return pesan("info", "Baru saja tercatat")
	}

	if !adaTrx {
		s.pending = &pendingInput{k.ID, k.Nama}
		s.siswaNISN, s.siswaNama = nisn, namaSiswa
		base.Aksi = "AMBIL"
		return base
	}

	base.JumlahAmbil = jAmbil
	s.pending = &pendingInput{k.ID, k.Nama}
	s.siswaNISN, s.siswaNama = nisn, namaSiswa

	if status == "KURANG" {
		// Susulan: tanpa peringatan jeda minimum.
		base.Aksi = "KEMBALI_SUSULAN"
		base.Sisa = jAmbil - jKembali
		return base
	}

	if t, err := time.ParseInLocation("2006-01-02 15:04", tgl+" "+waktuAmbil, timeutil.WIB); err == nil {
		menit := int(now.Sub(t).Minutes())
		if menit < s.jedaMenit() {
			base.Aksi = "KONFIRMASI_KEMBALI"
			base.MenitSejak = menit
			return base
		}
	}
	base.Aksi = "KEMBALI"
	return base
}

// BatalInput: dipanggil UI saat popup/konfirmasi dibatalkan, melepas penguncian.
func (s *Service) BatalInput() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending = nil
	s.siswaNISN, s.siswaNama = "", ""
}

func (s *Service) SubmitAmbil(kelasID int64, jumlah int) error {
	if jumlah < 1 {
		return errors.New("Jumlah ambil harus bilangan bulat minimal 1")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var nama, ketua string
	if err := tx.QueryRow(`SELECT nama,nama_ketua FROM kelas WHERE id=?`, kelasID).Scan(&nama, &ketua); err != nil {
		return errors.New("Kelas tidak ditemukan")
	}
	res, err := tx.Exec(`INSERT INTO transaksi (tanggal,kelas_id,nama_kelas,nama_ketua,waktu_ambil,jumlah_ambil,status,nisn_ambil,nama_ambil) VALUES (?,?,?,?,?,?,'AMBIL',?,?)`,
		timeutil.Tanggal(now), kelasID, nama, ketua, timeutil.Jam(now), jumlah, s.siswaNISN, s.siswaNama)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("Kelas ini sudah tercatat mengambil hari ini")
		}
		return err
	}
	id, _ := res.LastInsertId()
	if err := s.audit(tx, now, "TAMBAH", "transaksi", itoa(id), nil,
		map[string]any{"kelas": nama, "jumlah_ambil": jumlah, "status": "AMBIL", "pengambil": s.siswaNama}, ""); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.antreanWA(s.db, "AMBIL", "mbg-ambil-"+itoa(id), kelasID, map[string]string{
		"pengambil":    nonKosong(s.siswaNama),
		"jumlah_ambil": strconv.Itoa(jumlah),
		"tanggal":      timeutil.HariTanggalIndo(now),
		"jam":          timeutil.Jam(now),
	})
	s.lastSubmit[kelasID] = now
	s.pending = nil
	s.siswaNISN, s.siswaNama = "", ""
	return nil
}

// namaPengembali: nama unik semua pengembali (urut waktu), dipisah koma.
func namaPengembali(tx *sql.Tx, id int64) (string, error) {
	rows, err := tx.Query(`SELECT COALESCE(nama_siswa,'') FROM pengembalian WHERE transaksi_id=? ORDER BY id`, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var out []string
	ada := map[string]bool{}
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return "", err
		}
		n = strings.TrimSpace(n)
		if n == "" || ada[n] {
			continue
		}
		ada[n] = true
		out = append(out, n)
	}
	return strings.Join(out, ", "), rows.Err()
}

// SubmitKembali: pengembalian awal (status AMBIL, 0..jumlah_ambil) atau susulan (status KURANG, 1..sisa).
// Setiap pemanggilan menambah satu baris pengembalian; transaksi menyimpan total akumulasi.
func (s *Service) SubmitKembali(kelasID int64, jumlah int) error {
	if jumlah < 1 {
		return errors.New("Jumlah kembali harus minimal 1")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int64
	var status, nama string
	var jAmbil, jKembali int
	err = tx.QueryRow(`SELECT id,status,nama_kelas,COALESCE(jumlah_ambil,0),COALESCE(jumlah_kembali,0) FROM transaksi WHERE tanggal=? AND kelas_id=?`,
		timeutil.Tanggal(now), kelasID).Scan(&id, &status, &nama, &jAmbil, &jKembali)
	if err != nil {
		return errors.New("Belum ada pengambilan hari ini untuk kelas ini")
	}
	susulan := false
	switch status {
	case "AMBIL":
		jKembali = 0
		if jumlah > jAmbil {
			return fmt.Errorf("Jumlah kembali (%d) tidak boleh melebihi jumlah ambil (%d)", jumlah, jAmbil)
		}
	case "KURANG":
		susulan = true
		sisa := jAmbil - jKembali
		if jumlah < 1 {
			return fmt.Errorf("Jumlah susulan harus minimal 1 (sisa %d)", sisa)
		}
		if jumlah > sisa {
			return fmt.Errorf("Jumlah susulan (%d) tidak boleh melebihi sisa (%d)", jumlah, sisa)
		}
	default:
		return errors.New("Transaksi kelas ini sudah selesai")
	}

	total := jKembali + jumlah
	statusBaru := "KURANG"
	if total >= jAmbil {
		statusBaru = "SELESAI"
	}
	jam := timeutil.Jam(now)
	resP, err := tx.Exec(`INSERT INTO pengembalian (transaksi_id,waktu,jumlah,nisn,nama_siswa) VALUES (?,?,?,?,?)`,
		id, jam, jumlah, s.siswaNISN, s.siswaNama)
	if err != nil {
		return err
	}
	pid, _ := resP.LastInsertId()
	namaGabung, err := namaPengembali(tx, id)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE transaksi SET waktu_kembali=?,jumlah_kembali=?,status=?,nisn_kembali=?,nama_kembali=? WHERE id=?`,
		jam, total, statusBaru, s.siswaNISN, namaGabung, id); err != nil {
		return err
	}

	lama := map[string]any{"status": status}
	baru := map[string]any{"status": statusBaru, "jumlah_kembali": total, "pengembali": s.siswaNama}
	alasan := ""
	if susulan {
		lama["jumlah_kembali"] = jKembali
		baru["jumlah_susulan"] = jumlah
		alasan = "pengembalian susulan"
	}
	if err := s.audit(tx, now, "UBAH", "transaksi", itoa(id), lama, baru, alasan); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	sisa := jAmbil - total
	jenisKembali := "Pengembalian"
	if susulan {
		jenisKembali = "Pengembalian susulan"
	}
	s.antreanWA(s.db, "KEMBALI", "mbg-kembali-"+itoa(pid), kelasID, map[string]string{
		"jenis":          jenisKembali,
		"pengembali":     nonKosong(s.siswaNama),
		"jumlah_ambil":   strconv.Itoa(jAmbil),
		"jumlah_kembali": strconv.Itoa(jumlah),
		"total_kembali":  strconv.Itoa(total),
		"sisa":           strconv.Itoa(sisa),
		"status":         teksStatusKembali(sisa),
		"tanggal":        timeutil.HariTanggalIndo(now),
		"jam":            jam,
	})
	if statusBaru == "KURANG" {
		// dimundurkan agar efektif hanya cooldownSusulan (lihat komentar konstanta)
		s.lastSubmit[kelasID] = now.Add(-(cooldownKelas - cooldownSusulan))
	} else {
		s.lastSubmit[kelasID] = now
	}
	s.pending = nil
	s.siswaNISN, s.siswaNama = "", ""
	return nil
}

// StatusHariIni: data dashboard (keterangan dihitung saat tampil).
func (s *Service) StatusHariIni() (model.StatusHari, error) {
	now := s.now()
	tgl := timeutil.Tanggal(now)
	out := model.StatusHari{Tanggal: tgl, TanggalIndo: timeutil.HariTanggalIndo(now), Baris: []model.BarisHari{}}

	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n)
	out.Ditutup = n > 0

	rows, err := s.db.Query(`SELECT k.id,k.nama,k.nama_ketua,t.status,t.jumlah_ambil,t.jumlah_kembali,
		COALESCE(t.waktu_ambil,''),COALESCE(t.waktu_kembali,''),
		(SELECT COUNT(*) FROM pengembalian p WHERE p.transaksi_id=t.id)
		FROM kelas k LEFT JOIN transaksi t ON t.kelas_id=k.id AND t.tanggal=?
		WHERE k.aktif=1 ORDER BY k.urutan,k.nama`, tgl)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	semua := true
	for rows.Next() {
		var b model.BarisHari
		var st sql.NullString
		var ja, jk sql.NullInt64
		var np int
		if err := rows.Scan(&b.KelasID, &b.Nama, &b.NamaKetua, &st, &ja, &jk, &b.WaktuAmbil, &b.WaktuKembali, &np); err != nil {
			return out, err
		}
		b.Status = "BELUM_AMBIL"
		if st.Valid {
			b.Status = st.String
		}
		if ja.Valid {
			v := int(ja.Int64)
			b.JumlahAmbil = &v
		}
		if jk.Valid {
			v := int(jk.Int64)
			b.JumlahKembali = &v
		}
		b.Keterangan, b.Merah = Keterangan(b.Status, b.JumlahAmbil, b.JumlahKembali, "", np > 1)
		if b.Status != "SELESAI" {
			semua = false
		}
		out.Baris = append(out.Baris, b)
	}
	out.SemuaSelesai = semua && len(out.Baris) > 0
	out.BolehCetak = out.Ditutup || out.SemuaSelesai
	return out, rows.Err()
}

// Keterangan: aturan 4.3. Dipakai dashboard, riwayat, dan ekspor.
// susulan = true bila transaksi punya lebih dari satu pengembalian.
func Keterangan(status string, ambil, kembali *int, catatan string, susulan bool) (string, bool) {
	var ket string
	merah := false
	kurang := ""
	if ambil != nil && kembali != nil && *kembali < *ambil {
		kurang = fmt.Sprintf("Kurang %d Ompreng", *ambil-*kembali)
	}
	switch status {
	case "SELESAI":
		switch {
		case kurang != "":
			ket, merah = kurang, true
		case susulan:
			ket = "Lengkap (susulan)"
		default:
			ket = "Lengkap"
		}
	case "KURANG":
		ket, merah = kurang, true
		if ket == "" {
			ket = "Kurang"
		}
	case "BELUM_KEMBALI":
		ket, merah = "Belum kembali", true
		if kurang != "" {
			ket = kurang // sudah ada pengembalian sebagian sebelum hari ditutup
		}
	case "TIDAK_AMBIL":
		ket, merah = "Tidak mengambil", true
	}
	if catatan != "" {
		if ket != "" {
			ket += " - "
		}
		ket += catatan
	}
	return ket, merah
}
