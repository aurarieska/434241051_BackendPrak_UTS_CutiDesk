package model

import "time"

// JenisCuti adalah entitas jenis cuti di database
type JenisCuti struct {
	ID        int       `json:"id"`
	Nama      string    `json:"nama"`
	Deskripsi *string   `json:"deskripsi"` // nullable
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateJenisCutiRequest adalah body untuk POST /jenis-cuti
type CreateJenisCutiRequest struct {
	Nama      string  `json:"nama"      validate:"required,min=3,max=100"`
	Deskripsi *string `json:"deskripsi" validate:"omitnil,max=500"`
}

// UpdateJenisCutiRequest adalah body untuk PUT /jenis-cuti/:id
type UpdateJenisCutiRequest struct {
	Nama      string  `json:"nama"      validate:"required,min=3,max=100"`
	Deskripsi *string `json:"deskripsi" validate:"omitnil,max=500"`
	IsActive  *bool   `json:"is_active" validate:"required"`
}

// PatchJenisCutiRequest adalah body untuk PATCH /jenis-cuti/:id
type PatchJenisCutiRequest struct {
	Nama      *string `json:"nama"      validate:"omitnil,min=3,max=100"`
	Deskripsi *string `json:"deskripsi" validate:"omitnil,max=500"`
	IsActive  *bool   `json:"is_active" validate:"omitnil"`
}