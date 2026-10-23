package service

import (
	"errors"
	"strings"

	"mbg/internal/model"
	"mbg/internal/reader"
)

// ---------- Pengaturan ----------

func (s *Service) GetPengaturan() (model.Pengaturan, error) {
	var p model.Pengaturan
	err := s.db.QueryRow(`SELECT COALESCE(nama_instansi,''),COALESCE(alamat_instansi,''),
		COALESCE(penanggung_jawab,''),COALESCE(nip_penanggung_jawab,''),COALESCE(jabatan_penanggung_jawab,''),
		jeda_minimum_menit,hid_ambang_ms,COALESCE(port_serial,'') FROM pengaturan WHERE id=1`).
		Scan(&p.NamaInstansi, &p.AlamatInstansi, &p.PenanggungJawab, &p.NIPPenanggungJawab,
			&p.JabatanPenanggungJawab, &p.JedaMinimumMenit, &p.HIDAmbangMs, &p.PortSerial)
	return p, err
}

func (s *Service) SimpanPengaturan(p model.Pengaturan) error {
	if p.JedaMinimumMenit < 0 || p.JedaMinimumMenit > 600 {
		return errors.New("Jeda minimum harus 0-600 menit")
	}
	if p.HIDAmbangMs < 10 || p.HIDAmbangMs > 500 {
		return errors.New("Ambang HID harus 10-500 ms")
	}
	lama, _ := s.GetPengaturan()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`UPDATE pengaturan SET nama_instansi=?,alamat_instansi=?,penanggung_jawab=?,
		nip_penanggung_jawab=?,jabatan_penanggung_jawab=?,jeda_minimum_menit=?,hid_ambang_ms=?,port_serial=? WHERE id=1`,
		p.NamaInstansi, p.AlamatInstansi, p.PenanggungJawab, p.NIPPenanggungJawab,
		p.JabatanPenanggungJawab, p.JedaMinimumMenit, p.HIDAmbangMs, p.PortSerial)
	if err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "pengaturan", "1", lama, p, ""); err != nil {
		return err
	}
	return tx.Commit()
}

// ---------- Kelas ----------

func (s *Service) ListKelas(termasukNonaktif bool) ([]model.Kelas, error) {
	q := `SELECT id,nama,nama_ketua,urutan,aktif FROM kelas`
	if !termasukNonaktif {
		q += ` WHERE aktif=1`
	}
	q += ` ORDER BY urutan,nama`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Kelas{}
	for rows.Next() {
		var k model.Kelas
		var a int
		if err := rows.Scan(&k.ID, &k.Nama, &k.NamaKetua, &k.Urutan, &a); err != nil {
			return nil, err
		}
		k.Aktif = a == 1
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *Service) TambahKelas(nama, ketua string, urutan int) (int64, error) {
	nama, ketua = strings.TrimSpace(nama), strings.TrimSpace(ketua)
	if nama == "" || ketua == "" {
		return 0, errors.New("Nama kelas dan nama ketua wajib diisi")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO kelas (nama,nama_ketua,urutan) VALUES (?,?,?)`, nama, ketua, urutan)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, errors.New("Nama kelas sudah ada")
		}
		return 0, err
	}
	id, _ := res.LastInsertId()
	if err := s.audit(tx, s.now(), "TAMBAH", "kelas", itoa(id), nil, map[string]any{"nama": nama, "ketua": ketua, "urutan": urutan}, ""); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (s *Service) UbahKelas(id int64, nama, ketua string, urutan int) error {
	nama, ketua = strings.TrimSpace(nama), strings.TrimSpace(ketua)
	if nama == "" || ketua == "" {
		return errors.New("Nama kelas dan nama ketua wajib diisi")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var lama model.Kelas
	if err := tx.QueryRow(`SELECT id,nama,nama_ketua,urutan FROM kelas WHERE id=?`, id).
		Scan(&lama.ID, &lama.Nama, &lama.NamaKetua, &lama.Urutan); err != nil {
		return errors.New("Kelas tidak ditemukan")
	}
	if _, err := tx.Exec(`UPDATE kelas SET nama=?,nama_ketua=?,urutan=? WHERE id=?`, nama, ketua, urutan, id); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("Nama kelas sudah ada")
		}
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "kelas", itoa(id), lama, map[string]any{"nama": nama, "ketua": ketua, "urutan": urutan}, ""); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) NonaktifkanKelas(id int64, alasan string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE kelas SET aktif=0 WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "kelas", itoa(id), map[string]any{"aktif": 1}, map[string]any{"aktif": 0}, alasan); err != nil {
		return err
	}
	return tx.Commit()
}

// ---------- Kartu ----------

func (s *Service) ListKartu(kelasID int64) ([]model.Kartu, error) {
	rows, err := s.db.Query(`SELECT uid,kelas_id,aktif,COALESCE(nisn,''),COALESCE(nama_siswa,'') FROM kartu WHERE kelas_id=? ORDER BY nama_siswa,uid`, kelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Kartu{}
	for rows.Next() {
		var k model.Kartu
		var a int
		if err := rows.Scan(&k.UID, &k.KelasID, &a, &k.NISN, &k.NamaSiswa); err != nil {
			return nil, err
		}
		k.Aktif = a == 1
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *Service) DaftarkanKartu(uid string, kelasID int64, nisn, nama string) error {
	u, ok := reader.NormalisasiUID(uid)
	if !ok {
		return errors.New("UID tidak valid (harus 1-10 digit angka)")
	}
	nisn, nama = strings.TrimSpace(nisn), strings.TrimSpace(nama)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var namaKelas string
	if err := tx.QueryRow(`SELECT k.nama FROM kartu c JOIN kelas k ON k.id=c.kelas_id WHERE c.uid=?`, u).Scan(&namaKelas); err == nil {
		return errors.New("Kartu " + u + " sudah terdaftar untuk kelas " + namaKelas)
	}
	if _, err := tx.Exec(`INSERT INTO kartu (uid,kelas_id,nisn,nama_siswa) VALUES (?,?,?,?)`, u, kelasID, nisn, nama); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "TAMBAH", "kartu", u, nil, map[string]any{"kelas_id": kelasID, "nisn": nisn, "nama": nama}, ""); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) HapusKartu(uid string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var kid int64
	if err := tx.QueryRow(`SELECT kelas_id FROM kartu WHERE uid=?`, uid).Scan(&kid); err != nil {
		return errors.New("Kartu tidak ditemukan")
	}
	if _, err := tx.Exec(`DELETE FROM kartu WHERE uid=?`, uid); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "HAPUS", "kartu", uid, map[string]any{"kelas_id": kid}, nil, ""); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) AktifkanKelas(id int64, alasan string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE kelas SET aktif=1 WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "UBAH", "kelas", itoa(id), map[string]any{"aktif": 0}, map[string]any{"aktif": 1}, alasan); err != nil {
		return err
	}
	return tx.Commit()
}
