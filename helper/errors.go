package helper

import "fmt"

// AppError adalah error terstruktur yang diubah menjadi response HTTP
// oleh ErrorHandler terpusat di config/app.go
type AppError struct {
	StatusCode int               // kode HTTP (400, 401, dst.)
	Code       string            // kode stabil untuk mesin (BAD_REQUEST, dst.)
	Message    string            // pesan untuk manusia
	Fields     map[string]string // detail per field (khusus validasi)
	Cause      error             // penyebab asli — hanya untuk log, tidak dikirim ke client
}

// Error mengimplementasikan interface error
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// --- Constructor untuk setiap kode error ---

func BadRequest(pesan string) *AppError {
	return &AppError{StatusCode: 400, Code: "BAD_REQUEST", Message: pesan}
}

func Unauthorized(pesan string) *AppError {
	return &AppError{StatusCode: 401, Code: "UNAUTHORIZED", Message: pesan}
}

func Forbidden(pesan string) *AppError {
	return &AppError{StatusCode: 403, Code: "FORBIDDEN", Message: pesan}
}

func NotFound(pesan string) *AppError {
	return &AppError{StatusCode: 404, Code: "NOT_FOUND", Message: pesan}
}

func Conflict(pesan string) *AppError {
	return &AppError{StatusCode: 409, Code: "CONFLICT", Message: pesan}
}

func Validation(fields map[string]string) *AppError {
	return &AppError{
		StatusCode: 422,
		Code:       "VALIDATION_ERROR",
		Message:    "validasi gagal",
		Fields:     fields,
	}
}

func Internal(pesan string, cause error) *AppError {
	return &AppError{
		StatusCode: 500,
		Code:       "INTERNAL_ERROR",
		Message:    pesan,
		Cause:      cause,
	}
}

func UnsupportedMediaType(pesan string) *AppError {
	return &AppError{StatusCode: 415, Code: "UNSUPPORTED_MEDIA_TYPE", Message: pesan}
}

func NotAcceptable(pesan string) *AppError {
	return &AppError{StatusCode: 406, Code: "NOT_ACCEPTABLE", Message: pesan}
}

func TooManyRequests(pesan string) *AppError {
	return &AppError{StatusCode: 429, Code: "TOO_MANY_REQUESTS", Message: pesan}
}

func PayloadTooLarge(pesan string) *AppError {
	return &AppError{StatusCode: 413, Code: "PAYLOAD_TOO_LARGE", Message: pesan}
}

func ServiceUnavailable(pesan string) *AppError {
	return &AppError{StatusCode: 503, Code: "SERVICE_UNAVAILABLE", Message: pesan}
}