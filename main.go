package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"cutidesk/config"
	"cutidesk/database"
)

func main() {
	// 1. Muat environment dari file .env
	config.MuatEnv()

	// 2. Inisialisasi logger (console + file berrotasi)
	logger := config.SetupLogger()
	logger.Info("memulai aplikasi", "nama", config.AmbilEnv("APP_NAME", "CutiDesk"))

	// 3. Buat connection pool ke PostgreSQL
	pool := database.BuatPool()
	defer pool.Close()

	// === Tahap berikutnya (Push 2-7): repository, service, middleware, route, app ===

	port := config.AmbilEnv("APP_PORT", "3000")
	logger.Info("aplikasi siap — tekan Ctrl+C untuk berhenti", "port", port)

	// 4. Tunggu sinyal OS untuk graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("menerima sinyal shutdown", "signal", sig)

	pool.Close()
	slog.Info("aplikasi berhenti dengan aman")
}