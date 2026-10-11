package model

import "time"

// AuthUser menyimpan identitas pengguna yang sedang login (dari JWT)
type AuthUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// LoginRequest adalah body untuk POST /auth/login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest adalah body untuk POST /auth/register
type RegisterRequest struct {
	Username    string `json:"username"     validate:"required,min=3,max=30,username"`
	Email       string `json:"email"        validate:"required,email,max=120"`
	Password    string `json:"password"     validate:"required,min=8,max=72,strongpassword"`
	NamaLengkap string `json:"nama_lengkap" validate:"required,min=3,max=150"`
}

// RefreshRequest adalah body untuk POST /auth/refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest adalah body untuk POST /auth/logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenPair adalah pasangan token yang dikembalikan saat login/refresh
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// MeResponse adalah response untuk GET /auth/me
type MeResponse struct {
	User        User     `json:"user"`
	Permissions []string `json:"permissions"`
}

// RefreshTokenData menyimpan data refresh token dari database (internal)
type RefreshTokenData struct {
	ID        int64
	UserID    int
	ExpiresAt time.Time
	RevokedAt *time.Time
}