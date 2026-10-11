package helper

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validatorInstance adalah singleton — satu instance untuk seluruh aplikasi
var validatorInstance *validator.Validate

func init() {
	validatorInstance = validator.New()

	// Gunakan nama dari tag json sebagai nama field di pesan error
	validatorInstance.RegisterTagNameFunc(func(fld reflect.StructField) string {
		nama := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if nama == "-" {
			return ""
		}
		return nama
	})

	// Daftarkan custom validation tags
	validatorInstance.RegisterValidation("username", validasiUsername)
	validatorInstance.RegisterValidation("strongpassword", validasiPasswordKuat)
	validatorInstance.RegisterValidation("nospace", validasiTanpaSpasi)
}

// ValidateStruct menjalankan seluruh tag validasi pada struct
// Kembalikan nil jika valid, atau map[field]pesan jika ada pelanggaran
func ValidateStruct(s interface{}) map[string]string {
	err := validatorInstance.Struct(s)
	if err == nil {
		return nil
	}

	daftarError := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		daftarError[e.Field()] = pesanUntuk(e)
	}
	return daftarError
}

// pesanUntuk menerjemahkan tag validasi menjadi pesan bahasa Indonesia
func pesanUntuk(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "wajib diisi"
	case "min":
		return "minimal " + e.Param() + " karakter"
	case "max":
		return "maksimal " + e.Param() + " karakter"
	case "email":
		return "format email tidak valid"
	case "oneof":
		return "harus salah satu dari: " + strings.ReplaceAll(e.Param(), " ", ", ")
	case "username":
		return "hanya boleh huruf, angka, titik, dan underscore"
	case "strongpassword":
		return "harus mengandung huruf dan angka, dan bukan password umum"
	case "nospace":
		return "tidak boleh mengandung spasi"
	case "gt":
		return "harus lebih dari " + e.Param()
	default:
		return "nilai tidak valid"
	}
}

// --- Custom validation functions ---

// Regex: hanya huruf, angka, titik, underscore
var regexUsername = regexp.MustCompile(`^[a-zA-Z0-9._]+$`)

func validasiUsername(fl validator.FieldLevel) bool {
	return regexUsername.MatchString(fl.Field().String())
}

func validasiPasswordKuat(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	adaHuruf := false
	adaAngka := false

	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			adaHuruf = true
		}
		if c >= '0' && c <= '9' {
			adaAngka = true
		}
	}

	// Tolak password yang terlalu umum
	passwordUmum := []string{
		"password", "12345678", "qwerty123", "admin123",
		"password123", "abcdefgh", "letmein1",
	}
	for _, umum := range passwordUmum {
		if strings.EqualFold(password, umum) {
			return false
		}
	}

	return adaHuruf && adaAngka
}

func validasiTanpaSpasi(fl validator.FieldLevel) bool {
	return !strings.Contains(fl.Field().String(), " ")
}