package config

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// SetupLogger membuat logger dengan output ke console dan file berrotasi
func SetupLogger() *slog.Logger {
	// Tentukan level log dari environment
	levelTeks := strings.ToLower(AmbilEnv("LOG_LEVEL", "info"))
	var level slog.Level
	switch levelTeks {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Rotasi file log otomatis dengan lumberjack
	penulisFile := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10, // megabyte per file
		MaxBackups: 5,
		MaxAge:     30, // hari
	}

	// Gabungkan output: console + file
	gabungan := io.MultiWriter(os.Stdout, penulisFile)

	handler := slog.NewJSONHandler(gabungan, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}