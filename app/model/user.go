package model

import "time"

// User adalah entitas pengguna di database
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Password     string    `json:"-"` // tidak pernah keluar sebagai JSON
	NamaLengkap  string    `json:"nama_lengkap"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateUserRequest adalah body untuk POST /users (admin membuat user)
type CreateUserRequest struct {
	Username    string `json:"username"     validate:"required,min=3,max=30,username"`
	Email       string `json:"email"        validate:"required,email,max=120"`
	Password    string `json:"password"     validate:"required,min=8,max=72,strongpassword"`
	NamaLengkap string `json:"nama_lengkap" validate:"required,min=3,max=150"`
	Role        string `json:"role"         validate:"required,oneof=admin hr employee"`
}

// UpdateUserRequest adalah body untuk PUT /users/:id (penggantian penuh)
// is_active pakai *bool agar required bisa membedakan "tidak dikirim" vs "false"
type UpdateUserRequest struct {
	Username    string `json:"username"     validate:"required,min=3,max=30,username"`
	Email       string `json:"email"        validate:"required,email,max=120"`
	NamaLengkap string `json:"nama_lengkap" validate:"required,min=3,max=150"`
	IsActive    *bool  `json:"is_active"    validate:"required"`
}

// PatchUserRequest adalah body untuk PATCH /users/:id (perubahan sebagian)
// Semua field pointer — nil berarti "tidak diubah"
type PatchUserRequest struct {
	Username    *string `json:"username"     validate:"omitnil,min=3,max=30,username"`
	Email       *string `json:"email"        validate:"omitnil,email,max=120"`
	NamaLengkap *string `json:"nama_lengkap" validate:"omitnil,min=3,max=150"`
	IsActive    *bool   `json:"is_active"    validate:"omitnil"`
}

// AssignRoleRequest adalah body untuk PATCH /users/:id/role
type AssignRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin hr employee"`
}