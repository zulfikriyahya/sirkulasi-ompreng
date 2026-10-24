package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite"
)

// DataDir: %APPDATA%\MBG (Windows), $XDG_DATA_HOME/mbg atau ~/.local/share/mbg (Linux).
func DataDir() string {
	if runtime.GOOS == "windows" {
		if a := os.Getenv("APPDATA"); a != "" {
			return filepath.Join(a, "MBG")
		}
	}
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "mbg")
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".local", "share", "mbg")
}

func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", filepath.ToSlash(path))
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1) // satu penulis, hindari SQLITE_BUSY
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	if err := Migrate(conn); err != nil {
		return nil, fmt.Errorf("migrasi: %w", err)
	}
	return conn, nil
}

func Migrate(conn *sql.DB) error {
	if _, err := conn.Exec(schema); err != nil {
		return err
	}
	for _, k := range []struct{ tabel, kolom, tipe string }{
		{"kartu", "nisn", "TEXT"}, {"kartu", "nama_siswa", "TEXT"},
		{"transaksi", "nisn_ambil", "TEXT"}, {"transaksi", "nama_ambil", "TEXT"},
		{"transaksi", "nisn_kembali", "TEXT"}, {"transaksi", "nama_kembali", "TEXT"},
		{"pengaturan", "periksa_otomatis", "INTEGER NOT NULL DEFAULT 1"},
		{"pengaturan", "pra_rilis", "INTEGER NOT NULL DEFAULT 0"},
	} {
		if err := tambahKolom(conn, k.tabel, k.kolom, k.tipe); err != nil {
			return err
		}
	}
	// Backfill: satu baris pengembalian per transaksi SELESAI lama yang belum punya rincian.
	// Idempoten berkat NOT EXISTS, aman dijalankan setiap aplikasi dibuka.
	if _, err := conn.Exec(`INSERT INTO pengembalian (transaksi_id,waktu,jumlah,nisn,nama_siswa)
		SELECT t.id,t.waktu_kembali,t.jumlah_kembali,t.nisn_kembali,t.nama_kembali FROM transaksi t
		WHERE t.status='SELESAI' AND t.jumlah_kembali IS NOT NULL AND t.waktu_kembali IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM pengembalian p WHERE p.transaksi_id=t.id)`); err != nil {
		return err
	}
	_, err := conn.Exec(`INSERT OR IGNORE INTO pengaturan (id, jabatan_penanggung_jawab) VALUES (1, 'Penanggung Jawab MBG')`)
	return err
}

func tambahKolom(conn *sql.DB, tabel, kolom, tipe string) error {
	rows, err := conn.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, tabel))
	if err != nil {
		return err
	}
	ada := false
	for rows.Next() {
		var cid, notnull, pk int
		var nama, tipe string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &nama, &tipe, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		if nama == kolom {
			ada = true
		}
	}
	rows.Close()
	if ada {
		return nil
	}
	_, err = conn.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, tabel, kolom, tipe))
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS pengaturan (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  nama_instansi TEXT, alamat_instansi TEXT,
  logo BLOB, logo_mime TEXT,
  penanggung_jawab TEXT, nip_penanggung_jawab TEXT,
  jabatan_penanggung_jawab TEXT,
  jeda_minimum_menit INTEGER NOT NULL DEFAULT 10,
  hid_ambang_ms INTEGER NOT NULL DEFAULT 50,
  port_serial TEXT
);
CREATE TABLE IF NOT EXISTS kelas (
  id INTEGER PRIMARY KEY,
  nama TEXT NOT NULL UNIQUE,
  nama_ketua TEXT NOT NULL,
  urutan INTEGER NOT NULL DEFAULT 0,
  aktif INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS kartu (
  uid TEXT PRIMARY KEY,
  kelas_id INTEGER NOT NULL REFERENCES kelas(id),
  aktif INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS transaksi (
  id INTEGER PRIMARY KEY,
  tanggal TEXT NOT NULL,
  kelas_id INTEGER NOT NULL REFERENCES kelas(id),
  nama_kelas TEXT NOT NULL,
  nama_ketua TEXT NOT NULL,
  waktu_ambil TEXT,
  waktu_kembali TEXT,
  jumlah_ambil INTEGER,
  jumlah_kembali INTEGER,
  status TEXT NOT NULL,
  catatan TEXT,
  UNIQUE (tanggal, kelas_id)
);
CREATE TABLE IF NOT EXISTS pengembalian (
  id INTEGER PRIMARY KEY,
  transaksi_id INTEGER NOT NULL REFERENCES transaksi(id) ON DELETE CASCADE,
  waktu TEXT NOT NULL,
  jumlah INTEGER NOT NULL,
  nisn TEXT,
  nama_siswa TEXT
);
CREATE INDEX IF NOT EXISTS idx_pengembalian_trx ON pengembalian(transaksi_id);
CREATE TABLE IF NOT EXISTS hari_ditutup (
  tanggal TEXT PRIMARY KEY,
  waktu_tutup TEXT NOT NULL,
  catatan TEXT
);
CREATE TABLE IF NOT EXISTS audit_log (
  id INTEGER PRIMARY KEY,
  waktu TEXT NOT NULL,
  aksi TEXT NOT NULL,
  entitas TEXT NOT NULL,
  entitas_id TEXT,
  nilai_lama TEXT,
  nilai_baru TEXT,
  alasan TEXT
);
`
