package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const biayaBcrypt = 12

// HashPassword meng-hash password dengan bcrypt cost 12
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), biayaBcrypt)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword membandingkan password mentah dengan hash bcrypt
func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// VerifyDummyPassword menjalankan hash palsu untuk menyamakan waktu respons
// Dipakai saat username tidak ditemukan — mencegah timing attack
func VerifyDummyPassword() {
	dummyHash := "$2a$12$000000000000000000000uGWDMOtz2FXMJi2bOCEojMNOi6ENfm6"
	bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte("dummy"))
}

// RandomToken menghasilkan token acak 32 byte, dikembalikan sebagai hex string (64 karakter)
func RandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SHA256Hex meng-hash string dengan SHA-256 dan mengembalikan hex string
// Dipakai untuk menyimpan refresh token di database (bukan plaintext)
func SHA256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}