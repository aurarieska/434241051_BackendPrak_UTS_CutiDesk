package config

import (
	"encoding/json"
	"log/slog"

	"cutidesk/helper"

	"github.com/gofiber/fiber/v2"
)

// =====================================================================
// Konstanta nama permission — dipakai di routing untuk middleware authz.
// Harus sama persis dengan isi tabel permissions di database.
// =====================================================================

const (
	PermUserCreate  = "user:create"
	PermUserRead    = "user:read"
	PermUserUpdate  = "user:update"
	PermUserDelete  = "user:delete"

	PermJenisCutiCreate = "jenis_cuti:create"
	PermJenisCutiRead   = "jenis_cuti:read"
	PermJenisCutiUpdate = "jenis_cuti:update"
	PermJenisCutiDelete = "jenis_cuti:delete"

	PermCutiCreate  = "cuti:create"
	PermCutiRead    = "cuti:read"
	PermCutiApprove = "cuti:approve"

	PermRiwayatRead = "riwayat:read"

	PermLaporanRead   = "laporan:read"
	PermProfilUpdate  = "profil:update"
)

// BuatFiberApp membuat instance Fiber dengan konfigurasi standar.
// Centralized error handler: semua error dari service (AppError) dikonversi
// ke HTTP response di satu tempat.
func BuatFiberApp(logger *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		// Body limit 1 MB (tidak ada upload file, cukup untuk JSON).
		BodyLimit: 1 * 1024 * 1024,

		// Matikan startup banner Fiber.
		DisableStartupMessage: true,

		// Centralized error handler.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Cek apakah error bertipe AppError.
			if appErr, ok := err.(*helper.AppError); ok {
				// Log berdasarkan status code.
				if appErr.StatusCode >= 500 {
					logger.Error("Server error",
						"status", appErr.StatusCode,
						"code", appErr.Code,
						"message", appErr.Message,
						"path", c.Path(),
						"method", c.Method(),
						"cause", appErr.Cause,
					)
				}

				// Bangun response body.
				respBody := fiber.Map{
					"success":    false,
					"code":       appErr.Code,
					"message":    appErr.Message,
					"request_id": c.GetRespHeader(fiber.HeaderXRequestID),
				}

				// Tambahkan fields jika ada (validation errors).
				if appErr.Fields != nil {
					respBody["fields"] = appErr.Fields
				}

				return c.Status(appErr.StatusCode).JSON(respBody)
			}

			// Error bukan AppError (misal dari Fiber internal).
			// Cek apakah Fiber error.
			if fiberErr, ok := err.(*fiber.Error); ok {
				logger.Warn("Fiber error",
					"status", fiberErr.Code,
					"message", fiberErr.Message,
					"path", c.Path(),
				)
				return c.Status(fiberErr.Code).JSON(fiber.Map{
					"success":    false,
					"code":       "FIBER_ERROR",
					"message":    fiberErr.Message,
					"request_id": c.GetRespHeader(fiber.HeaderXRequestID),
				})
			}

			// Error tidak dikenal — 500 Internal Server Error.
			logger.Error("Unhandled error",
				"error", err.Error(),
				"path", c.Path(),
				"method", c.Method(),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success":    false,
				"code":       "INTERNAL_ERROR",
				"message":    "Terjadi kesalahan internal pada server",
				"request_id": c.GetRespHeader(fiber.HeaderXRequestID),
			})
		},

		// Custom JSON encoder (opsional, untuk konsistensi).
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	return app
}