package service

import "testing"

func TestNormalisasiWA(t *testing.T) {
	ok := map[string]string{
		"081234567890":      "6281234567890",
		"81234567890":       "6281234567890",
		"+62 812-3456-7890": "6281234567890",
		"6281234567890":     "6281234567890",
	}
	for in, want := range ok {
		if got, b := NormalisasiWA(in); !b || got != want {
			t.Errorf("%q => %q,%v; want %q", in, got, b, want)
		}
	}
	for _, in := range []string{"", "abc", "0712345", "12345678901", "08"} {
		if _, b := NormalisasiWA(in); b {
			t.Errorf("%q seharusnya tidak valid", in)
		}
	}
}
