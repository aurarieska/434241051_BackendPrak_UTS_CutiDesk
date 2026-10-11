package service

import (
	"time"

	"cutidesk/app/model"
)

// roleDefaultRegister mengembalikan role default untuk user yang mendaftar sendiri.
// Mass assignment prevention: role ditentukan server, bukan dari request body.
func roleDefaultRegister() string {
	return "employee"
}

// apakahTokenKedaluwarsa mengecek apakah refresh token sudah melewati masa berlaku.
func apakahTokenKedaluwarsa(data *model.RefreshTokenData) bool {
	return time.Now().After(data.ExpiresAt)
}

// apakahTokenDicabut mengecek apakah refresh token sudah di-revoke.
func apakahTokenDicabut(data *model.RefreshTokenData) bool {
	return data.RevokedAt != nil
}

// apakahUserAktif mengecek apakah user masih aktif (boleh login).
func apakahUserAktif(user *model.User) bool {
	return user.IsActive
}