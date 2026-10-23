package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"mbg/internal/model"
)

const kolomTrx = `id,tanggal,kelas_id,nama_kelas,nama_ketua,COALESCE(waktu_ambil,''),COALESCE(waktu_kembali,''),jumlah_ambil,jumlah_kembali,status,COALESCE(catatan,''),COALESCE(nama_ambil,''),COALESCE(nama_kembali,'')`

type pemindai interface{ Scan(dest ...any) error }

func scanTrx(r pemindai) (model.Transaksi, error) {
	var t model.Transaksi
	var ja, jk sql.NullInt64
	if err := r.Scan(&t.ID, &t.Tanggal, &t.KelasID, &t.NamaKelas, &t.NamaKetua, &t.WaktuAmbil, &t.WaktuKembali,
		&ja, &jk, &t.Status, &t.Catatan, &t.NamaAmbil, &t.NamaKembali); err != nil {
		return t, err
	}
	if ja.Valid {
		v := int(ja.Int64)
		t.JumlahAmbil = &v
	}
	if jk.Valid {
		v := int(jk.Int64)
		t.JumlahKembali = &v
	}
	t.Keterangan, t.Merah = Keterangan(t.Status, t.JumlahAmbil, t.JumlahKembali, t.Catatan)
	return t, nil
}

func nilInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func nilStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func jamValid(s string) bool { _, err := time.Parse("15:04", s); return err == nil }

func validasiEdit(e model.EditTransaksi) error {
	switch e.Status {
	case "AMBIL", "SELESAI", "BELUM_KEMBALI", "TIDAK_AMBIL":
	default:
		return errors.New("Status tidak valid")
	}
	if e.Status == "TIDAK_AMBIL" {
		return nil
	}
	if e.JumlahAmbil == nil || *e.JumlahAmbil < 1 {
		return errors.New("Jumlah ambil harus bilangan bulat minimal 1")
	}
	if !jamValid(e.WaktuAmbil) {
		return errors.New("Waktu ambil harus berformat HH:MM")
	}
	if e.Status == "SELESAI" {
		if e.JumlahKembali == nil {
			return errors.New("Status Selesai wajib punya jumlah kembali")
		}
		if *e.JumlahKembali < 0 {
			return errors.New("Jumlah kembali tidak boleh negatif")
		}
		if *e.JumlahKembali > *e.JumlahAmbil {
			return fmt.Errorf("Jumlah kembali (%d) tidak boleh melebihi jumlah ambil (%d)", *e.JumlahKembali, *e.JumlahAmbil)
		}
		if !jamValid(e.WaktuKembali) {
			return errors.New("Waktu kembali harus berformat HH:MM")
		}
	}
	return nil
}

func (s *Service) ListTransaksi(f model.FilterTransaksi) ([]model.Transaksi, error) {
	q := `SELECT ` + kolomTrx + ` FROM transaksi WHERE 1=1`
	var args []any
	if f.Dari != "" {
		q += ` AND tanggal>=?`
		args = append(args, f.Dari)
	}
	if f.Sampai != "" {
		q += ` AND tanggal<=?`
		args = append(args, f.Sampai)
	}
	if f.KelasID > 0 {
		q += ` AND kelas_id=?`
		args = append(args, f.KelasID)
	}
	if f.Status != "" {
		q += ` AND status=?`
		args = append(args, f.Status)
	}
	q += ` ORDER BY tanggal DESC, (SELECT urutan FROM kelas k WHERE k.id=transaksi.kelas_id), nama_kelas`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Transaksi{}
	for rows.Next() {
		t, err := scanTrx(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Service) UbahTransaksi(e model.EditTransaksi) error {
	if strings.TrimSpace(e.Alasan) == "" {
		return errors.New("Alasan perubahan wajib diisi")
	}
	if err := validasiEdit(e); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	lama, err := scanTrx(tx.QueryRow(`SELECT `+kolomTrx+` FROM transaksi WHERE id=?`, e.ID))
	if err != nil {
		return errors.New("Transaksi tidak ditemukan")
	}
	ja, jk, wa, wk := e.JumlahAmbil, e.JumlahKembali, e.WaktuAmbil, e.WaktuKembali
	switch e.Status {
	case "AMBIL", "BELUM_KEMBALI":
		jk, wk = nil, ""
	case "TIDAK_AMBIL":
		ja, jk, wa, wk = nil, nil, "", ""
	}
	catatan := strings.TrimSpace(e.Catatan)
	if _, err := tx.Exec(`UPDATE transaksi SET status=?,jumlah_ambil=?,jumlah_kembali=?,waktu_ambil=?,waktu_kembali=?,catatan=? WHERE id=?`,
		e.Status, nilInt(ja), nilInt(jk), nilStr(wa), nilStr(wk), nilStr(catatan), e.ID); err != nil {
		return err
	}
	baru := map[string]any{"status": e.Status, "jumlah_ambil": nilInt(ja), "jumlah_kembali": nilInt(jk),
		"waktu_ambil": wa, "waktu_kembali": wk, "catatan": catatan}
	if err := s.audit(tx, s.now(), "UBAH", "transaksi", itoa(e.ID), lama, baru, strings.TrimSpace(e.Alasan)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Service) HapusTransaksi(id int64, alasan string) error {
	alasan = strings.TrimSpace(alasan)
	if alasan == "" {
		return errors.New("Alasan penghapusan wajib diisi")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	lama, err := scanTrx(tx.QueryRow(`SELECT `+kolomTrx+` FROM transaksi WHERE id=?`, id))
	if err != nil {
		return errors.New("Transaksi tidak ditemukan")
	}
	if _, err := tx.Exec(`DELETE FROM transaksi WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.audit(tx, s.now(), "HAPUS", "transaksi", itoa(id), lama, nil, alasan); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	delete(s.lastSubmit, lama.KelasID)
	return nil
}

func (s *Service) ListAudit(f model.FilterAudit) ([]model.Audit, error) {
	q := `SELECT id,waktu,aksi,entitas,COALESCE(entitas_id,''),COALESCE(nilai_lama,''),COALESCE(nilai_baru,''),COALESCE(alasan,'') FROM audit_log WHERE 1=1`
	var args []any
	if f.Dari != "" {
		q += ` AND substr(waktu,1,10)>=?`
		args = append(args, f.Dari)
	}
	if f.Sampai != "" {
		q += ` AND substr(waktu,1,10)<=?`
		args = append(args, f.Sampai)
	}
	if f.Entitas != "" {
		q += ` AND entitas=?`
		args = append(args, f.Entitas)
	}
	lim := f.Limit
	if lim <= 0 || lim > 1000 {
		lim = 200
	}
	q += fmt.Sprintf(` ORDER BY id DESC LIMIT %d`, lim)
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Audit{}
	for rows.Next() {
		var a model.Audit
		if err := rows.Scan(&a.ID, &a.Waktu, &a.Aksi, &a.Entitas, &a.EntitasID, &a.NilaiLama, &a.NilaiBaru, &a.Alasan); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
