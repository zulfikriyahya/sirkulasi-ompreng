package service

import (
	"errors"
	"strings"

	"mbg/internal/timeutil"
)

// TutupHari: kelas AMBIL/KURANG -> BELUM_KEMBALI, kelas aktif tanpa transaksi -> TIDAK_AMBIL.
func (s *Service) TutupHari(catatan string) error {
	catatan = strings.TrimSpace(catatan)
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	tgl := timeutil.Tanggal(now)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tgl).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return errors.New("Hari ini sudah ditutup")
	}
	r1, err := tx.Exec(`UPDATE transaksi SET status='BELUM_KEMBALI' WHERE tanggal=? AND status IN ('AMBIL','KURANG')`, tgl)
	if err != nil {
		return err
	}
	belum, _ := r1.RowsAffected()
	r2, err := tx.Exec(`INSERT INTO transaksi (tanggal,kelas_id,nama_kelas,nama_ketua,status)
		SELECT ?,k.id,k.nama,k.nama_ketua,'TIDAK_AMBIL' FROM kelas k
		WHERE k.aktif=1 AND NOT EXISTS (SELECT 1 FROM transaksi t WHERE t.tanggal=? AND t.kelas_id=k.id)`, tgl, tgl)
	if err != nil {
		return err
	}
	tidak, _ := r2.RowsAffected()
	if _, err := tx.Exec(`INSERT INTO hari_ditutup (tanggal,waktu_tutup,catatan) VALUES (?,?,?)`,
		tgl, timeutil.Jam(now), catatan); err != nil {
		return err
	}
	if err := s.audit(tx, now, "TUTUP_HARI", "hari", tgl, nil,
		map[string]any{"belum_kembali": belum, "tidak_ambil": tidak}, catatan); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.pending = nil
	return nil
}

// BukaHari: hanya hari ini. Status otomatis dikembalikan supaya tap bisa dilanjutkan:
// BELUM_KEMBALI tanpa pengembalian -> AMBIL, BELUM_KEMBALI dengan pengembalian sebagian -> KURANG,
// baris TIDAK_AMBIL otomatis dihapus. Koreksi hari lampau dilakukan lewat Riwayat (edit/hapus).
func (s *Service) BukaHari(tanggal, alasan string) error {
	if strings.TrimSpace(alasan) == "" {
		return errors.New("Alasan wajib diisi")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if tanggal != timeutil.Tanggal(now) {
		return errors.New("Hanya hari ini yang bisa dibuka kembali; koreksi hari sebelumnya lewat menu Riwayat")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tanggal).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return errors.New("Hari ini tidak dalam keadaan ditutup")
	}
	r1, err := tx.Exec(`UPDATE transaksi SET status='AMBIL' WHERE tanggal=? AND status='BELUM_KEMBALI' AND jumlah_kembali IS NULL`, tanggal)
	if err != nil {
		return err
	}
	r3, err := tx.Exec(`UPDATE transaksi SET status='KURANG' WHERE tanggal=? AND status='BELUM_KEMBALI' AND jumlah_kembali IS NOT NULL AND jumlah_kembali < jumlah_ambil`, tanggal)
	if err != nil {
		return err
	}
	r2, err := tx.Exec(`DELETE FROM transaksi WHERE tanggal=? AND status='TIDAK_AMBIL' AND jumlah_ambil IS NULL`, tanggal)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM hari_ditutup WHERE tanggal=?`, tanggal); err != nil {
		return err
	}
	a, _ := r1.RowsAffected()
	k, _ := r3.RowsAffected()
	b, _ := r2.RowsAffected()
	if err := s.audit(tx, now, "BUKA_HARI", "hari", tanggal, nil,
		map[string]any{"dikembalikan_ke_ambil": a, "dikembalikan_ke_kurang": k, "baris_tidak_ambil_dihapus": b}, alasan); err != nil {
		return err
	}
	return tx.Commit()
}

// ApakahBolehCetak: hari ditutup, atau semua kelas aktif berstatus SELESAI pada tanggal itu.
// Status KURANG belum dihitung selesai.
func (s *Service) ApakahBolehCetak(tanggal string) (bool, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM hari_ditutup WHERE tanggal=?`, tanggal).Scan(&n); err != nil {
		return false, err
	}
	if n > 0 {
		return true, nil
	}
	var aktif, selesai int
	err := s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(CASE WHEN t.status='SELESAI' THEN 1 ELSE 0 END),0)
		FROM kelas k LEFT JOIN transaksi t ON t.kelas_id=k.id AND t.tanggal=? WHERE k.aktif=1`, tanggal).Scan(&aktif, &selesai)
	if err != nil {
		return false, err
	}
	return aktif > 0 && aktif == selesai, nil
}
