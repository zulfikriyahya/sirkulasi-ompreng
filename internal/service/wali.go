package service

import (
	"errors"
	"strings"

	"mbg/internal/model"
)

// NormalisasiWA mengubah 08xx / 8xx / +62 8xx / 62 8xx menjadi 628xxxxxxxxx.
func NormalisasiWA(s string) (string, bool) {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '+' || r == '(' || r == ')' || r == '.':
		default:
			return "", false
		}
	}
	d := b.String()
	switch {
	case strings.HasPrefix(d, "62"):
	case strings.HasPrefix(d, "0"):
		d = "62" + d[1:]
	case strings.HasPrefix(d, "8"):
		d = "62" + d
	default:
		return "", false
	}
	if len(d) < 10 || len(d) > 15 || d[2] != '8' {
		return "", false
	}
	return d, true
}

func (s *Service) ListWaliKelas(termasukNonaktif bool) ([]model.WaliKelas, error) {
	q := `SELECT w.id,w.nama,w.no_wa,w.aktif,
                COALESCE((SELECT group_concat(k.nama, ', ') FROM kelas k WHERE k.wali_kelas_id=w.id AND k.aktif=1),'')
                FROM wali_kelas w`
	if !termasukNonaktif {
		q += ` WHERE w.aktif=1`
	}
	q += ` ORDER BY w.nama`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.WaliKelas{}
	for rows.Next() {
		var w model.WaliKelas
		var a int
		if err := rows.Scan(&w.ID, &w.Nama, &w.NoWA, &a, &w.DaftarKelas); err != nil {
			return nil, err
		}
		w.Aktif = a == 1
		out = append(out, w)
	}
	return out, rows.Err()
}

func cekWali(nama, noWA string) (string, string, error) {
	nama = strings.TrimSpace(nama)
	if nama == "" {
		return "", "", errors.New("Nama wali kelas wajib diisi")
	}
	wa, ok := NormalisasiWA(noWA)
	if !ok {
		return "", "", errors.New("Nomor WhatsApp tidak valid (contoh: 081234567890)")
	}
	return nama, wa, nil
}

func (s *Service) TambahWaliKelas(nama, noWA string) (int64, error) {
	nama, wa, err := cekWali(nama, noWA)
	if err != nil {
		return 0, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO wali_kelas (nama,no_wa) VALUES (?,?)`, nama, wa)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	if err := s.audit(tx, s.now(), "TAMBAH", "wali_kelas", itoa(id), nil, map[string]any{"nama": nama, "no_wa": wa}, ""); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (s *Service) UbahWaliKelas(id int64, nama, noWA string) error {
	nama, wa, err := cekWali(nama, noWA)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var ln, lw string
	if err := tx.QueryRow(`SELECT nama,no_wa FROM wali_kelas WHERE id=?`, id).Scan(&ln, &lw); err != nil {
		return errors.New("Wali kelas tidak ditemukan")
	}
	if _, err := tx.Exec(`UPDATE wali_kelas SET nama=?,no_wa=? WHERE id=?`, nama, wa, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "wali_kelas", itoa(id),
		map[string]any{"nama": ln, "no_wa": lw}, map[string]any{"nama": nama, "no_wa": wa}, ""); err != nil {
		return err
	}
	return tx.Commit()
}

// NonaktifkanWaliKelas juga melepas tautan ke kelas agar notifikasi tidak terkirim ke wali nonaktif.
func (s *Service) NonaktifkanWaliKelas(id int64, alasan string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE kelas SET wali_kelas_id=NULL WHERE wali_kelas_id=?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE wali_kelas SET aktif=0 WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "wali_kelas", itoa(id), map[string]any{"aktif": 1}, map[string]any{"aktif": 0}, alasan); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) AktifkanWaliKelas(id int64, alasan string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE wali_kelas SET aktif=1 WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "wali_kelas", itoa(id), map[string]any{"aktif": 0}, map[string]any{"aktif": 1}, alasan); err != nil {
		return err
	}
	return tx.Commit()
}

// SetWaliKelas menautkan kelas ke wali. waliID 0 = lepas tautan.
func (s *Service) SetWaliKelas(kelasID, waliID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var lama int64
	if err := tx.QueryRow(`SELECT COALESCE(wali_kelas_id,0) FROM kelas WHERE id=?`, kelasID).Scan(&lama); err != nil {
		return errors.New("Kelas tidak ditemukan")
	}
	var baru any
	if waliID != 0 {
		var a int
		if err := tx.QueryRow(`SELECT aktif FROM wali_kelas WHERE id=?`, waliID).Scan(&a); err != nil || a != 1 {
			return errors.New("Wali kelas tidak ditemukan atau nonaktif")
		}
		baru = waliID
	}
	if _, err := tx.Exec(`UPDATE kelas SET wali_kelas_id=? WHERE id=?`, baru, kelasID); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "kelas", itoa(kelasID),
		map[string]any{"wali_kelas_id": lama}, map[string]any{"wali_kelas_id": waliID}, "set wali kelas"); err != nil {
		return err
	}
	return tx.Commit()
}
