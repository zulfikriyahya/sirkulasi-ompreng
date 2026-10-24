// Package versi menyimpan nomor versi aplikasi. Nilainya disuntikkan saat build:
//
//	wails build -ldflags "-X mbg/internal/versi.Versi=0.4.0"
package versi

var Versi = "dev"
