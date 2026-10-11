package helper

import (
	"cutidesk/app/model"

	"github.com/gofiber/fiber/v2"
)

const kunciAuthUser = "auth_user"

// SimpanAuthUser menyimpan identitas user ke Fiber Locals (dipanggil oleh middleware)
func SimpanAuthUser(c *fiber.Ctx, user *model.AuthUser) {
	c.Locals(kunciAuthUser, user)
}

// AmbilAuthUser mengambil identitas user dari Fiber Locals (dipanggil oleh service)
// Kembalikan nil jika belum login
func AmbilAuthUser(c *fiber.Ctx) *model.AuthUser {
	user, ok := c.Locals(kunciAuthUser).(*model.AuthUser)
	if !ok {
		return nil
	}
	return user
}