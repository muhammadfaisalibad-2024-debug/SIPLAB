package config

import (
	"log"
	"os"
	"strconv"

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