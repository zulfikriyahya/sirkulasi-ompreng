package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"mbg/internal/timeutil"
	"mbg/internal/versi"
)

type pendingInput struct {
	KelasID int64
	Nama    string
}

type Service struct {
	db  *sql.DB
	dir string
	now func() time.Time

	mu         sync.Mutex
	lastUID    string
	lastUIDAt  time.Time
	pending    *pendingInput
	lastSubmit map[int64]time.Time // cooldown per kelas
	rdr        readerState
	app        appCtx
	siswaNISN  string
	siswaNama  string
	pemb       pembState
}

func New(db *sql.DB, dir string) *Service {
	return NewWithClock(db, dir, timeutil.Now)
}

func NewWithClock(db *sql.DB, dir string, now func() time.Time) *Service {
	return &Service{db: db, dir: dir, now: now, lastSubmit: map[int64]time.Time{}}
}

type InfoSistem struct {
	Versi   string `json:"versi"`
	DataDir string `json:"data_dir"`
}

// InfoSistem: versi dibaca dari internal/versi (disuntikkan lewat -ldflags saat build).
func (s *Service) InfoSistem() InfoSistem { return InfoSistem{Versi: versi.Versi, DataDir: s.dir} }

func js(v any) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func (s *Service) audit(tx *sql.Tx, now time.Time, aksi, entitas, id string, lama, baru any, alasan string) error {
	_, err := tx.Exec(`INSERT INTO audit_log (waktu,aksi,entitas,entitas_id,nilai_lama,nilai_baru,alasan) VALUES (?,?,?,?,?,?,?)`,
		now.In(timeutil.WIB).Format(time.RFC3339), aksi, entitas, id, js(lama), js(baru), alasan)
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	return nil
}
