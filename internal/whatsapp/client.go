// Package whatsapp: klien untuk POST /api/v1/messages dan GET /api/v1/messages/{job_id}
// dengan autentikasi HMAC-SHA256 (X-API-Key, X-Signature, X-Timestamp).
package whatsapp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrAutentikasi = errors.New("autentikasi ditolak (API key, secret, atau jam komputer salah)")
	ErrDibatasi    = errors.New("API key dijeda otomatis oleh layanan (429); hubungi admin instansi")
)

const maksRespons = 1 << 20

type Config struct {
	BaseURL string
	KeyID   string
	Secret  string
}

type Pesan struct {
	TemplateCode string            `json:"template_code"`
	Recipient    string            `json:"recipient"`
	Variables    map[string]string `json:"variables"`
	Media        any               `json:"media"` // nil -> null
	ReferenceID  string            `json:"reference_id,omitempty"`
}

type Hasil struct {
	JobID  string
	Status string // queued | processing | sent | delivered | read | failed
	Pesan  string
}

// Error: respons non-2xx dari layanan.
type Error struct {
	Status int
	Pesan  string
}

func (e *Error) Error() string { return fmt.Sprintf("layanan WhatsApp HTTP %d: %s", e.Status, e.Pesan) }

func (e *Error) Unwrap() error {
	switch e.Status {
	case http.StatusUnauthorized:
		return ErrAutentikasi
	case http.StatusTooManyRequests:
		return ErrDibatasi
	}
	return nil
}

// IsSementara: true bila layak dicoba ulang nanti (jaringan putus, 5xx, 408).
// 400/401/422 dan 429 tidak sementara: mengulang tidak akan membantu.
func IsSementara(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Status >= 500 || e.Status == http.StatusRequestTimeout
	}
	return !errors.Is(err, context.Canceled)
}

type Client struct {
	cfg  Config
	http *http.Client
	now  func() time.Time
}

func New(cfg Config) (*Client, error) {
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.KeyID = strings.TrimSpace(cfg.KeyID)
	cfg.Secret = strings.TrimSpace(cfg.Secret)
	u, err := url.Parse(cfg.BaseURL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return nil, errors.New("Alamat layanan WhatsApp tidak valid (contoh: https://whatsapp.mtsn1pandeglang.sch.id)")
	}
	if cfg.KeyID == "" || cfg.Secret == "" {
		return nil, errors.New("API key dan secret WhatsApp wajib diisi")
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: 15 * time.Second}, now: time.Now}, nil
}

// tandatangani: HMAC-SHA256 dari body mentah, di-hex. Body kosong untuk GET.
func (c *Client) tandatangani(body []byte) string {
	m := hmac.New(sha256.New, []byte(c.cfg.Secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

func (c *Client) do(ctx context.Context, method, path string, body []byte) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.cfg.KeyID)
	req.Header.Set("X-Signature", c.tandatangani(body))
	req.Header.Set("X-Timestamp", strconv.FormatInt(c.now().Unix(), 10))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi layanan WhatsApp: %w", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, maksRespons))
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		pesan := ambil(m, "message", "error", "detail")
		if pesan == "" {
			pesan = strings.TrimSpace(string(raw))
			if len(pesan) > 200 {
				pesan = pesan[:200]
			}
		}
		return nil, &Error{Status: res.StatusCode, Pesan: pesan}
	}
	return m, nil
}

// ambil mencari kunci di level atas, lalu di dalam objek "data".
func ambil(m map[string]any, kunci ...string) string {
	cari := func(x map[string]any) string {
		for _, k := range kunci {
			switch v := x[k].(type) {
			case string:
				if v != "" {
					return v
				}
			case float64:
				return strconv.FormatInt(int64(v), 10)
			}
		}
		return ""
	}
	if s := cari(m); s != "" {
		return s
	}
	if d, ok := m["data"].(map[string]any); ok {
		return cari(d)
	}
	return ""
}

func hasil(m map[string]any) Hasil {
	return Hasil{JobID: ambil(m, "job_id", "id"), Status: ambil(m, "status"), Pesan: ambil(m, "message")}
}

// Kirim memasukkan pesan ke antrean layanan. Body di-marshal sekali, byte yang sama
// dipakai untuk tanda tangan dan dikirim.
func (c *Client) Kirim(ctx context.Context, p Pesan) (Hasil, error) {
	if p.TemplateCode == "" || p.Recipient == "" {
		return Hasil{}, errors.New("template_code dan recipient wajib diisi")
	}
	if p.Variables == nil {
		p.Variables = map[string]string{}
	}
	body, err := json.Marshal(p)
	if err != nil {
		return Hasil{}, err
	}
	m, err := c.do(ctx, http.MethodPost, "/api/v1/messages", body)
	if err != nil {
		return Hasil{}, err
	}
	h := hasil(m)
	if h.JobID == "" {
		return h, errors.New("respons layanan tidak memuat job_id")
	}
	return h, nil
}

// Status memeriksa status pengiriman. Signature dihitung dari body kosong.
func (c *Client) Status(ctx context.Context, jobID string) (Hasil, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return Hasil{}, errors.New("job_id kosong")
	}
	m, err := c.do(ctx, http.MethodGet, "/api/v1/messages/"+url.PathEscape(jobID), nil)
	if err != nil {
		return Hasil{}, err
	}
	h := hasil(m)
	if h.JobID == "" {
		h.JobID = jobID
	}
	return h, nil
}
