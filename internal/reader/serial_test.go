package reader

import "testing"

func TestParseBaris(t *testing.T) {
	cases := []struct{ in, jenis string }{
		{"0012345678\r", "UID"},
		{"READY", "READY"},
		{"ERR:RC522", "ERR"},
		{"12345", "ABAIKAN"},
		{"00123456789", "ABAIKAN"},
		{"abc1234567", "ABAIKAN"},
		{"", "ABAIKAN"},
	}
	for _, c := range cases {
		if j, _ := ParseBaris(c.in); j != c.jenis {
			t.Errorf("ParseBaris(%q) = %s, mau %s", c.in, j, c.jenis)
		}
	}
}
