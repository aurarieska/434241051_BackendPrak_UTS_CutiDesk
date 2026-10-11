package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// MuatEnv memuat variabel dari file .env ke environment
func MuatEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("file .env tidak ditemukan, menggunakan environment sistem")
	}
}

// AmbilEnv mengambil nilai environment variable, return fallback jika kosong
func AmbilEnv(kunci, fallback string) string {
	nilai := os.Getenv(kunci)
	if nilai == "" {
		return fallback
	}
	return nilai
}

// AmbilEnvInt mengambil nilai environment variable sebagai integer
func AmbilEnvInt(kunci string, fallback int) int {
	nilai := os.Getenv(kunci)
	if nilai == "" {
		return fallback
	}
	hasil, err := strconv.Atoi(nilai)
	if err != nil {
		return fallback
	}
	return hasil
}

// AmbilEnvWajib mengambil nilai env, aplikasi berhenti jika kosong
func AmbilEnvWajib(kunci string) string {
	nilai := os.Getenv(kunci)
	if nilai == "" {
		log.Fatalf("FATAL: environment variable %s wajib diisi", kunci)
	}
	return nilai
}