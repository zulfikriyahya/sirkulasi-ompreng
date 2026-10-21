package reader

import "testing"

func TestNormalisasiUID(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"0012345678\r\n", "0012345678", true},
		{" 12345678 ", "0012345678", true},
		{"1234567890", "1234567890", true},
		{"12345678901", "", false},
		{"abc", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, ok := NormalisasiUID(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("NormalisasiUID(%q) = %q,%v; want %q,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}
