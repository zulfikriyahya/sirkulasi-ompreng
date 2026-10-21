package reader

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

const vidEspressif = "303A"

type Status struct {
	Terhubung bool   `json:"terhubung"`
	Port      string `json:"port"`
}

// Manager memindai port tiap 2 detik, membuka reader ESP32-C3, dan membaca per baris.
type Manager struct {
	port     func() string // port manual dari Pengaturan ("" = otomatis)
	onUID    func(string)
	onStatus func(Status)
	lastErr  string
}

func NewManager(port func() string, onUID func(string), onStatus func(Status)) *Manager {
	return &Manager{port: port, onUID: onUID, onStatus: onStatus}
}

// ParseBaris mengklasifikasi satu baris dari firmware: UID | READY | ERR | ABAIKAN.
func ParseBaris(s string) (jenis, isi string) {
	s = strings.TrimSpace(s)
	switch {
	case s == "READY":
		return "READY", s
	case strings.HasPrefix(s, "ERR:"):
		return "ERR", s
	case len(s) == 10 && strings.Trim(s, "0123456789") == "":
		return "UID", s
	}
	return "ABAIKAN", s
}

func tidur(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

func (m *Manager) cari() string {
	if p := strings.TrimSpace(m.port()); p != "" {
		return p
	}
	if p := os.Getenv("MBG_PORT"); p != "" { // untuk pengujian dengan port virtual
		return p
	}
	list, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return ""
	}
	for _, d := range list {
		if d.IsUSB && strings.EqualFold(d.VID, vidEspressif) {
			return d.Name
		}
	}
	return ""
}

func (m *Manager) Run(ctx context.Context) {
	m.onStatus(Status{})
	for ctx.Err() == nil {
		name := m.cari()
		if name == "" {
			tidur(ctx, 2*time.Second)
			continue
		}
		m.baca(ctx, name)
		m.onStatus(Status{})
		tidur(ctx, 2*time.Second)
	}
}

func (m *Manager) baca(ctx context.Context, name string) {
	p, err := serial.Open(name, &serial.Mode{BaudRate: 115200})
	if err != nil {
		if msg := err.Error(); msg != m.lastErr { // hindari spam log
			log.Printf("[reader] gagal membuka %s: %v", name, err)
			m.lastErr = msg
		}
		return
	}
	m.lastErr = ""
	if err := p.SetDTR(true); err != nil { // USB CDC ESP32 hanya mengirim jika DTR aktif
	}
	_ = p.ResetInputBuffer()
	dibuka := time.Now()
	m.onStatus(Status{Terhubung: true, Port: name})

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			p.Close()
		case <-done:
		}
	}()
	sc := bufio.NewScanner(p)
	for sc.Scan() {
		jenis, isi := ParseBaris(sc.Text())
		switch jenis {
		case "UID":
			if time.Since(dibuka) < 500*time.Millisecond { // sisa buffer reader sebelum konek
				continue
			}
			m.onUID(isi)
		case "READY":
		case "ERR":
		default:
			if isi != "" {
			}
		}
	}
	close(done)
	p.Close()
}
