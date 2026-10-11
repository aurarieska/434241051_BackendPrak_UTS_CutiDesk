package model

import "time"

// PengajuanCuti adalah entitas pengajuan cuti di database
type PengajuanCuti struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	JenisCutiID    int       `json:"jenis_cuti_id"`
	TanggalMulai   string    `json:"tanggal_mulai"`   // format YYYY-MM-DD
	TanggalSelesai string    `json:"tanggal_selesai"`  // format YYYY-MM-DD
	Alasan         string    `json:"alasan"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// Field dari JOIN (terisi saat query gabungan)
	NamaPengaju   string `json:"nama_pengaju,omitempty"`
	NamaJenisCuti string `json:"nama_jenis_cuti,omitempty"`
}

// CreatePengajuanCutiRequest adalah body untuk POST /cuti
type CreatePengajuanCutiRequest struct {
	JenisCutiID    int    `json:"jenis_cuti_id"    validate:"required"`
	TanggalMulai   string `json:"tanggal_mulai"    validate:"required"`
	TanggalSelesai string `json:"tanggal_selesai"  validate:"required"`
	Alasan         string `json:"alasan"           validate:"required,min=10,max=500"`
}

// UbahStatusRequest adalah body untuk PATCH /cuti/:id/status
type UbahStatusRequest struct {
	Status  string  `json:"status"  validate:"required,oneof=approved rejected"`
	Catatan *string `json:"catatan" validate:"omitnil,max=500"`
}