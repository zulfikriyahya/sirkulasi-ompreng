package service

import (
	"sync"

	"mbg/internal/reader"
)

type readerState struct {
	mu sync.Mutex
	st reader.Status
}

func (s *Service) SetStatusReader(st reader.Status) {
	s.rdr.mu.Lock()
	s.rdr.st = st
	s.rdr.mu.Unlock()
}

// StatusReader: dipanggil frontend saat start (event pertama bisa terlewat).
func (s *Service) StatusReader() reader.Status {
	s.rdr.mu.Lock()
	defer s.rdr.mu.Unlock()
	return s.rdr.st
}
