package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"cutidesk/config"
)

// BuatPool membuat connection pool ke PostgreSQL
func BuatPool() *pgxpool.Pool {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.AmbilEnv("DB_USER", "postgres"),
		config.AmbilEnvWajib("DB_PASSWORD"),
		config.AmbilEnv("DB_HOST", "localhost"),
		config.AmbilEnv("DB_PORT", "5432"),
		config.AmbilEnv("DB_NAME", "cutidesk"),
		config.AmbilEnv("DB_SSLMODE", "disable"),
	)

	konfigPool, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("gagal parse konfigurasi database", "error", err)
		panic(err)
	}

	maxConns := config.AmbilEnvInt("DB_MAX_CONNS", 10)
	konfigPool.MaxConns = int32(maxConns)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, konfigPool)
	if err != nil {
		slog.Error("gagal membuat pool koneksi database", "error", err)
		panic(err)
	}

	// Tes koneksi awal
	if err := pool.Ping(ctx); err != nil {
		slog.Error("gagal ping database", "error", err)
		panic(err)
	}

	slog.Info("koneksi database berhasil",
		"host", config.AmbilEnv("DB_HOST", "localhost"),
		"database", config.AmbilEnv("DB_NAME", "cutidesk"),
	)

	return pool
}