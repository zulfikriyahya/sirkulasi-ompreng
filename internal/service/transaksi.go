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
	debounceTap   = 2 * time.Second
	cooldownKelas = 30 * time.Second
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
	err := s.db.QueryRow(`SELECT k.id,k.nama,k.nama_ketua,c.aktif,k.aktif FROM kartu c JOIN kelas k ON k.id=c.kelas_id WHERE c.uid=?`, uid).
		Scan(&k.ID, &k.Nama, &k.NamaKetua, &kartuAktif, &kelasAktif)
	if err != nil || kartuAktif != 1 || kelasAktif != 1 {
		return pesan("err", "Kartu tidak terdaftar")
	}

	tgl := timeutil.Tanggal(now)
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n)
	if n > 0 {
		return pesan("err", "Hari ini sudah ditutup")
	}

	if t, ok := s.lastSubmit[k.ID]; ok && now.Sub(t) < cooldownKelas {
		return pesan("info", "Baru saja tercatat")
	}

	base := model.TapResult{KelasID: k.ID, NamaKelas: k.Nama, NamaKetua: k.NamaKetua, Tipe: "ok"}

	var status, waktuAmbil string
	var jAmbil int
	err = s.db.QueryRow(`SELECT status,COALESCE(waktu_ambil,''),COALESCE(jumlah_ambil,0) FROM transaksi WHERE tanggal=? AND kelas_id=?`, tgl, k.ID).
		Scan(&status, &waktuAmbil, &jAmbil)
	if errors.Is(err, sql.ErrNoRows) {
		s.pending = &pendingInput{k.ID, k.Nama}
		base.Aksi = "AMBIL"
		return base
	}
	if err != nil {
		return pesan("err", "Kesalahan database")
	}

	if status != "AMBIL" {
		return pesan("info", "Kelas "+k.Nama+" sudah selesai hari ini")
	}

	base.JumlahAmbil = jAmbil
	s.pending = &pendingInput{k.ID, k.Nama}
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
	if s.pending != nil {
	}
	s.pending = nil
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
	res, err := tx.Exec(`INSERT INTO transaksi (tanggal,kelas_id,nama_kelas,nama_ketua,waktu_ambil,jumlah_ambil,status) VALUES (?,?,?,?,?,?,'AMBIL')`,
		timeutil.Tanggal(now), kelasID, nama, ketua, timeutil.Jam(now), jumlah)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("Kelas ini sudah tercatat mengambil hari ini")
		}
		return err
	}
	id, _ := res.LastInsertId()
	if err := s.audit(tx, now, "TAMBAH", "transaksi", itoa(id), nil,
		map[string]any{"kelas": nama, "jumlah_ambil": jumlah, "status": "AMBIL"}, ""); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.lastSubmit[kelasID] = now
	s.pending = nil
	return nil
}

func (s *Service) SubmitKembali(kelasID int64, jumlah int) error {
	if jumlah < 0 {
		return errors.New("Jumlah kembali tidak boleh negatif")
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
	var jAmbil int
	err = tx.QueryRow(`SELECT id,status,nama_kelas,COALESCE(jumlah_ambil,0) FROM transaksi WHERE tanggal=? AND kelas_id=?`,
		timeutil.Tanggal(now), kelasID).Scan(&id, &status, &nama, &jAmbil)
	if err != nil {
		return errors.New("Belum ada pengambilan hari ini untuk kelas ini")
	}
	if status != "AMBIL" {
		return errors.New("Transaksi kelas ini sudah selesai")
	}
	if jumlah > jAmbil {
		return fmt.Errorf("Jumlah kembali (%d) tidak boleh melebihi jumlah ambil (%d)", jumlah, jAmbil)
	}
	if _, err := tx.Exec(`UPDATE transaksi SET waktu_kembali=?,jumlah_kembali=?,status='SELESAI' WHERE id=?`,
		timeutil.Jam(now), jumlah, id); err != nil {
		return err
	}
	if err := s.audit(tx, now, "UBAH", "transaksi", itoa(id),
		map[string]any{"status": "AMBIL"},
		map[string]any{"status": "SELESAI", "jumlah_kembali": jumlah}, ""); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.lastSubmit[kelasID] = now
	s.pending = nil
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
		COALESCE(t.waktu_ambil,''),COALESCE(t.waktu_kembali,'')
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
		if err := rows.Scan(&b.KelasID, &b.Nama, &b.NamaKetua, &st, &ja, &jk, &b.WaktuAmbil, &b.WaktuKembali); err != nil {
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
		b.Keterangan, b.Merah = Keterangan(b.Status, b.JumlahAmbil, b.JumlahKembali, "")
		if b.Status != "SELESAI" {
			semua = false
		}
		out.Baris = append(out.Baris, b)
	}
	out.SemuaSelesai = semua && len(out.Baris) > 0
	out.BolehCetak = out.Ditutup || out.SemuaSelesai
	return out, rows.Err()
}

// Keterangan: aturan 4.3. Dipakai dashboard dan nanti ekspor.
func Keterangan(status string, ambil, kembali *int, catatan string) (string, bool) {
	var ket string
	merah := false
	switch status {
	case "SELESAI":
		if ambil != nil && kembali != nil && *kembali < *ambil {
			ket, merah = fmt.Sprintf("Kurang %d Ompreng", *ambil-*kembali), true
		} else {
			ket = "Lengkap"
		}
	case "BELUM_KEMBALI":
		ket, merah = "Belum kembali", true
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
