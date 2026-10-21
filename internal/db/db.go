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
	_, err := conn.Exec(`INSERT OR IGNORE INTO pengaturan (id, jabatan_penanggung_jawab) VALUES (1, 'Penanggung Jawab MBG')`)
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
