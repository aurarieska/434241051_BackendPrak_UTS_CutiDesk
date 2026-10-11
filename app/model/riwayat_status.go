package model

import "time"

// RiwayatStatus mencatat setiap perubahan status pengajuan cuti
type RiwayatStatus struct {
	ID              int       `json:"id"`
	PengajuanCutiID int       `json:"pengajuan_cuti_id"`
	StatusSebelum   *string   `json:"status_sebelum"` // null pada pembuatan awal
	StatusSesudah   string    `json:"status_sesudah"`
	Catatan         *string   `json:"catatan"`
	DiubahOleh      int       `json:"diubah_oleh"`
	NamaPengubah    string    `json:"nama_pengubah,omitempty"` // dari JOIN
	CreatedAt       time.Time `json:"created_at"`
}