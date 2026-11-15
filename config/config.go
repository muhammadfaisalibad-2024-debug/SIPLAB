package config

import (
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// LoadEnv memuat file .env ke dalam environment variabel
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: File .env tidak ditemukan, menggunakan environment variabel bawaan sistem")
	}
}

// GetEnv mengambil nilai environment variabel berdasarkan key, dengan fallback jika kosong
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

// GetEnvInt mengambil nilai environment variabel sebagai integer, dengan fallback jika kosong/error
func GetEnvInt(key string, fallback int) int {
	valStr := GetEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return fallback
}

// NewLogger membuat structured logger sederhana menggunakan slog (stdout)
func NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// NewApp membuat instance Fiber app dasar untuk Dependency Injection
func NewApp(logger *slog.Logger, deps interface{}) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "SIPLAB API",
	})
	return app
}