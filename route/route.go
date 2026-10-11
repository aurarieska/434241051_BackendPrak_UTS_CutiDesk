package route

import (
	"cutidesk/app/middleware"
	"cutidesk/app/service"
	"cutidesk/config"
	"cutidesk/helper"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes mendaftarkan seluruh route API.
func SetupRoutes(
	app *fiber.App,
	jwtManager *helper.JWTManager,
	setPermission *helper.SetPermission,
	authService *service.AuthService,
	userService *service.UserService,
	jenisCutiService *service.JenisCutiService,
	pengajuanCutiService *service.PengajuanCutiService,
) {
	// Root health check.
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "CutiDesk API berjalan",
		})
	})

	// Grup API v1.
	api := app.Group("/api")

	// =================================================================
	// AUTH ROUTES — sebagian public, sebagian butuh autentikasi
	// =================================================================
	auth := api.Group("/auth")

	// Public (tanpa autentikasi).
	auth.Post("/register", authService.Register)
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.RefreshToken)

	// Butuh autentikasi.
	auth.Post("/logout", middleware.Auth(jwtManager), authService.Logout)
	auth.Get("/me", middleware.Auth(jwtManager), authService.Me)

	// =================================================================
	// USER ROUTES — semua butuh autentikasi + permission
	// =================================================================
	users := api.Group("/users",
		middleware.Auth(jwtManager),
	)

	users.Get("/",
		middleware.PerluPermission(setPermission, config.PermUserRead),
		userService.Daftar,
	)
	users.Get("/:id",
		middleware.PerluPermission(setPermission, config.PermUserRead),
		userService.Detail,
	)
	users.Post("/",
		middleware.PerluPermission(setPermission, config.PermUserCreate),
		userService.Buat,
	)
	users.Put("/:id",
		middleware.PerluPermission(setPermission, config.PermUserUpdate),
		userService.UpdateFull,
	)
	users.Patch("/:id",
		middleware.PerluPermission(setPermission, config.PermUserUpdate),
		userService.UpdateParsial,
	)
	users.Patch("/:id/role",
		middleware.PerluPermission(setPermission, config.PermUserUpdate),
		userService.UbahRole,
	)

	// =================================================================
	// JENIS CUTI ROUTES
	// =================================================================
	jenisCuti := api.Group("/jenis-cuti",
		middleware.Auth(jwtManager),
	)

	jenisCuti.Get("/",
		middleware.PerluPermission(setPermission, config.PermJenisCutiRead),
		jenisCutiService.Daftar,
	)
	jenisCuti.Get("/:id",
		middleware.PerluPermission(setPermission, config.PermJenisCutiRead),
		jenisCutiService.Detail,
	)
	jenisCuti.Post("/",
		middleware.PerluPermission(setPermission, config.PermJenisCutiCreate),
		jenisCutiService.Buat,
	)
	jenisCuti.Put("/:id",
		middleware.PerluPermission(setPermission, config.PermJenisCutiUpdate),
		jenisCutiService.UpdateFull,
	)
	jenisCuti.Patch("/:id",
		middleware.PerluPermission(setPermission, config.PermJenisCutiUpdate),
		jenisCutiService.UpdateParsial,
	)
	jenisCuti.Delete("/:id",
		middleware.PerluPermission(setPermission, config.PermJenisCutiDelete),
		jenisCutiService.Hapus,
	)

	// =================================================================
	// PENGAJUAN CUTI ROUTES
	// =================================================================
	pengajuanCuti := api.Group("/pengajuan-cuti",
		middleware.Auth(jwtManager),
	)

	pengajuanCuti.Post("/",
		middleware.PerluPermission(setPermission, config.PermCutiCreate),
		pengajuanCutiService.Buat,
	)
	pengajuanCuti.Get("/",
		middleware.PerluPermission(setPermission, config.PermCutiRead),
		pengajuanCutiService.Daftar,
	)
	pengajuanCuti.Get("/:id",
		middleware.PerluPermission(setPermission, config.PermCutiRead),
		pengajuanCutiService.Detail,
	)
	pengajuanCuti.Patch("/:id/status",
		middleware.PerluPermission(setPermission, config.PermCutiApprove),
		pengajuanCutiService.UbahStatus,
	)
	pengajuanCuti.Get("/:id/riwayat",
		middleware.PerluPermission(setPermission, config.PermRiwayatRead),
		pengajuanCutiService.RiwayatStatus,
	)

	// =================================================================
	// 404 handler untuk route yang tidak ditemukan
	// =================================================================
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("Endpoint tidak ditemukan: " + c.Method() + " " + c.Path())
	})
}