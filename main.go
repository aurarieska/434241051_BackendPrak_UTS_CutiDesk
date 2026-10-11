package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cutidesk/app/middleware"
	"cutidesk/app/repository"
	"cutidesk/app/service"
	"cutidesk/config"
	"cutidesk/database"
	"cutidesk/helper"
	"cutidesk/route"
)

func main() {
	// =====================================================================
	// 1. Muat environment variables dari .env
	// =====================================================================
	config.MuatEnv()

	// =====================================================================
	// 2. Setup logger (JSON, stdout + file rotasi)
	// =====================================================================
	logger := config.SetupLogger()
	logger.Info("Memulai CutiDesk API...")

	// =====================================================================
	// 3. Koneksi database PostgreSQL
	// =====================================================================
	pool := database.BuatPool()
	defer pool.Close()
	logger.Info("Koneksi database berhasil")

	// =====================================================================
	// 4. Inisialisasi repository
	// =====================================================================
	userRepo := repository.BuatUserRepository(pool)
	tokenRepo := repository.BuatTokenRepository(pool)
	roleRepo := repository.BuatRoleRepository(pool)
	jenisCutiRepo := repository.BuatJenisCutiRepository(pool)
	pengajuanCutiRepo := repository.BuatPengajuanCutiRepository(pool)
	riwayatStatusRepo := repository.BuatRiwayatStatusRepository(pool)

	// =====================================================================
	// 5. Muat permission dari database → SetPermission (in-memory)
	// =====================================================================
	ctx := context.Background()
	dataPermission, err := roleRepo.MuatPermissions(ctx)
	if err != nil {
		log.Fatalf("Gagal memuat permissions dari database: %v", err)
	}
	setPermission := helper.BuatSetPermission(dataPermission)
	logger.Info("Permissions berhasil dimuat",
		"jumlah_role", len(dataPermission),
	)

	// =====================================================================
	// 6. Inisialisasi JWT Manager
	//    🔧 JWT_SECRET harus diisi di file .env (minimal 32 karakter)
	// =====================================================================
	jwtSecret := config.AmbilEnvWajib("JWT_SECRET")
	jwtIssuer := config.AmbilEnv("JWT_ISSUER", "cutidesk")
	jwtTTLMenit := config.AmbilEnvInt("JWT_TTL_MENIT", 15)
	jwtManager := helper.BuatJWTManager(jwtSecret, jwtIssuer, jwtTTLMenit)

	// =====================================================================
	// 7. Inisialisasi service
	// =====================================================================
	durasiRefreshHari := config.AmbilEnvInt("REFRESH_TOKEN_HARI", 7)

	authService := service.BuatAuthService(
		userRepo, tokenRepo, roleRepo,
		jwtManager, setPermission, logger,
		durasiRefreshHari,
	)

	userService := service.BuatUserService(
		userRepo, roleRepo, setPermission, logger,
	)

	jenisCutiService := service.BuatJenisCutiService(
		jenisCutiRepo, logger,
	)

	pengajuanCutiService := service.BuatPengajuanCutiService(
		pengajuanCutiRepo, jenisCutiRepo, riwayatStatusRepo, logger,
	)

	// =====================================================================
	// 8. Buat Fiber app dengan centralized error handler
	// =====================================================================
	app := config.BuatFiberApp(logger)

	// =====================================================================
	// 9. Pasang middleware global
	// =====================================================================
	middleware.SetupMiddleware(app, logger)

	maksRPS := config.AmbilEnvInt("RATE_LIMIT_RPS", 30)
	middleware.SetupRateLimiter(app, maksRPS)

	// =====================================================================
	// 10. Daftarkan semua route
	// =====================================================================
	route.SetupRoutes(
		app, jwtManager, setPermission,
		authService, userService,
		jenisCutiService, pengajuanCutiService,
	)

	// =====================================================================
	// 11. Jalankan server dengan graceful shutdown
	// =====================================================================
	port := config.AmbilEnv("APP_PORT", "3000")

	// Goroutine: jalankan server.
	go func() {
		logger.Info("Server berjalan", "port", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Gagal menjalankan server: %v", err)
		}
	}()

	// Tunggu sinyal shutdown (Ctrl+C).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Mematikan server...")
	if err := app.Shutdown(); err != nil {
		logger.Error("Gagal mematikan server", "error", err)
	}
	logger.Info("Server berhasil dimatikan")
}