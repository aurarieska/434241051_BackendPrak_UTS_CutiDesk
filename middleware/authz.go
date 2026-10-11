package middleware

import (
	"cutidesk/helper"

	"github.com/gofiber/fiber/v2"
)

// PerluPermission membuat middleware yang memeriksa apakah user yang
// terautentikasi memiliki permission tertentu berdasarkan role-nya.
//
// Middleware ini HARUS dipasang setelah Auth middleware.
// Prinsip fail-closed: jika role tidak dikenal atau permission tidak
// ditemukan, akses ditolak.
func PerluPermission(setPermission *helper.SetPermission, namaPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil data user dari context (sudah disimpan oleh Auth middleware).
		authUser := helper.AmbilAuthUser(c)
		if authUser == nil {
			return helper.Unauthorized("Autentikasi diperlukan")
		}

		// Cek apakah role dikenal.
		if !setPermission.RoleDikenal(authUser.Role) {
			return helper.Forbidden("Role tidak dikenal, akses ditolak")
		}

		// Cek apakah role punya permission yang diminta.
		if !setPermission.Boleh(authUser.Role, namaPermission) {
			return helper.Forbidden("Anda tidak memiliki izin untuk aksi ini")
		}

		return c.Next()
	}
}

// PerluRole membuat middleware yang memeriksa apakah user memiliki
// salah satu dari role yang diizinkan.
// Berguna untuk endpoint yang tidak perlu granular permission,
// cukup cek level role saja.
func PerluRole(roleDiizinkan ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authUser := helper.AmbilAuthUser(c)
		if authUser == nil {
			return helper.Unauthorized("Autentikasi diperlukan")
		}

		for _, role := range roleDiizinkan {
			if authUser.Role == role {
				return c.Next()
			}
		}

		return helper.Forbidden("Role Anda tidak memiliki akses ke endpoint ini")
	}
}