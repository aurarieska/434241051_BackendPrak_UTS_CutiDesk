package middleware

import (
	"strings"

	"cutidesk/helper"

	"github.com/gofiber/fiber/v2"
)

// Auth membuat middleware yang memvalidasi JWT access token.
// jwtManager dipakai untuk parsing & verifikasi token.
func Auth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil header Authorization.
		headerAuth := c.Get("Authorization")
		if headerAuth == "" {
			return helper.Unauthorized("Token autentikasi diperlukan")
		}

		// Format yang diharapkan: "Bearer <token>"
		bagian := strings.SplitN(headerAuth, " ", 2)
		if len(bagian) != 2 || !strings.EqualFold(bagian[0], "Bearer") {
			return helper.Unauthorized("Format token tidak valid, gunakan: Bearer <token>")
		}

		tokenString := bagian[1]

		// Parse dan validasi token.
		authUser, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			return helper.Unauthorized("Token tidak valid atau sudah kedaluwarsa")
		}

		// Simpan data user yang terautentikasi ke Fiber context.
		helper.SimpanAuthUser(c, authUser)

		return c.Next()
	}
}