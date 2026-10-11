package helper

import (
	"strconv"
	"time"

	"cutidesk/app/model"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager mengelola pembuatan dan parsing JWT access token
type JWTManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// BuatJWTManager membuat instance JWTManager baru
// Parameter diberikan dari luar (main.go membaca config, lalu meneruskan ke sini)
func BuatJWTManager(secret string, issuer string, ttlMenit int) *JWTManager {
	if len(secret) < 32 {
		panic("JWT_SECRET harus minimal 32 karakter")
	}
	return &JWTManager{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    time.Duration(ttlMenit) * time.Minute,
	}
}

// BuatAccessToken menghasilkan JWT access token dari identitas user
// Kembalikan: token string, masa berlaku dalam detik, error
func (j *JWTManager) BuatAccessToken(user model.AuthUser) (string, int, error) {
	sekarang := time.Now()
	kadaluarsa := sekarang.Add(j.ttl)

	claims := jwt.MapClaims{
		"sub":      strconv.Itoa(user.ID),
		"username": user.Username,
		"role":     user.Role,
		"iss":      j.issuer,
		"iat":      sekarang.Unix(),
		"exp":      kadaluarsa.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, int(j.ttl.Seconds()), nil
}

// ParseToken memvalidasi JWT dan mengembalikan identitas user dari claims
func (j *JWTManager) ParseToken(tokenString string) (*model.AuthUser, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Cegah algorithm confusion — hanya terima HMAC
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, Unauthorized("algoritma token tidak valid")
		}
		return j.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, Unauthorized("token tidak valid atau kedaluwarsa")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, Unauthorized("claims token tidak valid")
	}

	// Ambil user ID dari claim "sub"
	subStr, _ := claims["sub"].(string)
	id, err := strconv.Atoi(subStr)
	if err != nil {
		return nil, Unauthorized("sub token tidak valid")
	}

	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)

	return &model.AuthUser{
		ID:       id,
		Username: username,
		Role:     role,
	}, nil
}