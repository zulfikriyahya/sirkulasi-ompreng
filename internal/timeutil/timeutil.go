package timeutil

import (
	"fmt"
	"time"
)

var WIB = time.FixedZone("WIB", 7*3600)

var hari = [...]string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
var bulan = [...]string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember"}

func Now() time.Time { return time.Now().In(WIB) }

func Tanggal(t time.Time) string { return t.In(WIB).Format("2006-01-02") }
func Jam(t time.Time) string     { return t.In(WIB).Format("15:04") }

// "12 Januari 2026"
func TanggalIndo(t time.Time) string {
	t = t.In(WIB)
	return fmt.Sprintf("%d %s %d", t.Day(), bulan[t.Month()], t.Year())
}

// "Senin, 12 Januari 2026"
func HariTanggalIndo(t time.Time) string {
	t = t.In(WIB)
	return hari[t.Weekday()] + ", " + TanggalIndo(t)
}

// "11:46 WIB"
func JamWIB(t time.Time) string { return Jam(t) + " WIB" }
