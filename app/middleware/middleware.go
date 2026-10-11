package middleware

import (
	"log/slog"
	"time"

	"cutidesk/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// SetupMiddleware memasang middleware global ke aplikasi Fiber.
func SetupMiddleware(app *fiber.App, logger *slog.Logger) {
	// 1. Request ID — setiap request punya ID unik untuk tracing.
	app.Use(requestid.New())

	// 2. Recover — tangkap panic agar server tidak crash.
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// 3. CORS — izinkan akses dari origin manapun (sesuaikan di production).
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// 4. Request logging — catat setiap request masuk dan durasi prosesnya.
	app.Use(buatLogMiddleware(logger))
}

// SetupRateLimiter memasang rate limiter global.
// Batas: maksRPS request per detik per IP.
func SetupRateLimiter(app *fiber.App, maksRPS int) {
	app.Use(limiter.New(limiter.Config{
		Max:        maksRPS,
		Expiration: 1 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return helper.TooManyRequests("Terlalu banyak permintaan, coba lagi nanti")
		},
	}))
}

// buatLogMiddleware membuat middleware yang mencatat setiap request.
// Level log: WARN untuk 4xx, ERROR untuk 5xx, INFO untuk sisanya.
func buatLogMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mulai := time.Now()

		// Lanjutkan ke handler berikutnya.
		err := c.Next()

		durasi := time.Since(mulai)
		status := c.Response().StatusCode()

		atribut := []slog.Attr{
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("durasi", durasi),
			slog.String("ip", c.IP()),
			slog.String("request_id", c.GetRespHeader(fiber.HeaderXRequestID)),
		}

		pesan := c.Method() + " " + c.Path()

		switch {
		case status >= 500:
			logger.LogAttrs(c.Context(), slog.LevelError, pesan, atribut...)
		case status >= 400:
			logger.LogAttrs(c.Context(), slog.LevelWarn, pesan, atribut...)
		default:
			logger.LogAttrs(c.Context(), slog.LevelInfo, pesan, atribut...)
		}

		return err
	}
}