package service

import (
	"encoding/base64"
	"net/http"
)

// GetLogoPreview: data URI logo untuk pratinjau, "" bila belum ada logo.
func (s *Service) GetLogoPreview() (string, error) {
	var b []byte
	if err := s.db.QueryRow(`SELECT logo FROM pengaturan WHERE id=1`).Scan(&b); err != nil {
		return "", err
	}
	if len(b) == 0 {
		return "", nil
	}
	return "data:" + http.DetectContentType(b) + ";base64," + base64.StdEncoding.EncodeToString(b), nil
}
