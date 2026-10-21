package reader

import "strings"

// NormalisasiUID membuang semua non-digit lalu mengisi nol di kiri sampai 10 digit.
// Hasil dengan 0 atau >10 digit dianggap tidak valid.
func NormalisasiUID(s string) (string, bool) {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	if len(d) < 1 || len(d) > 10 {
		return "", false
	}
	return strings.Repeat("0", 10-len(d)) + d, true
}
