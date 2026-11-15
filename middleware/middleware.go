package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"siplab/helper"
)

// RequireJSON memastikan request dengan method POST, PUT, atau PATCH mengirimkan header
// Content-Type: application/json (mengembalikan 415 UNSUPPORTED_MEDIA_TYPE jika tidak sesuai).
func RequireJSON(c *fiber.Ctx) error {
	switch c.Method() {
	case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch:
		contentType := strings.ToLower(strings.TrimSpace(c.Get(fiber.HeaderContentType)))
		if !strings.HasPrefix(contentType, fiber.MIMEApplicationJSON) {
			return helper.UnsupportedMediaType("header Content-Type wajib application/json")
		}
	}
	return c.Next()
}

type rateBucket struct {
	count   int
	resetAt time.Time
}

// LoginRateLimiter membatasi percobaan login maksimal 5 kali per menit per alamat IP
// tanpa bergantung pada package eksternal di luar standar library & Fiber core.
func LoginRateLimiter() fiber.Handler {
	var (
		mu      sync.Mutex
		buckets = make(map[string]rateBucket)
	)
	const (
		maxAttempts = 5
		window      = 1 * time.Minute
	)

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()

		mu.Lock()
		b, exists := buckets[ip]
		if !exists || now.After(b.resetAt) {
			b = rateBucket{
				count:   1,
				resetAt: now.Add(window),
			}
			buckets[ip] = b
			mu.Unlock()
			return c.Next()
		}

		if b.count >= maxAttempts {
			mu.Unlock()
			return helper.TooManyRequests("terlalu banyak percobaan login, silakan coba lagi setelah 1 menit")
		}

		b.count++
		buckets[ip] = b
		mu.Unlock()
		return c.Next()
	}
}