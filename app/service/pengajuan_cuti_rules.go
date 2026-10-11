package service

import (
	"time"
)

// validasiTanggalCuti mengecek aturan bisnis tanggal pengajuan cuti.
// Return pesan error jika melanggar, atau string kosong jika valid.
func validasiTanggalCuti(tanggalMulai, tanggalSelesai string) string {
	mulai, err := time.Parse("2006-01-02", tanggalMulai)
	if err != nil {
		return "Format tanggal_mulai tidak valid, gunakan YYYY-MM-DD"
	}

	selesai, err := time.Parse("2006-01-02", tanggalSelesai)
	if err != nil {
		return "Format tanggal_selesai tidak valid, gunakan YYYY-MM-DD"
	}

	// Tanggal mulai tidak boleh di masa lalu.
	hari_ini := time.Now().Truncate(24 * time.Hour)
	if mulai.Before(hari_ini) {
		return "Tanggal mulai tidak boleh di masa lalu"
	}

	// Tanggal selesai harus >= tanggal mulai.
	if selesai.Before(mulai) {
		return "Tanggal selesai harus sama dengan atau setelah tanggal mulai"
	}

	return ""
}

// apakahStatusBolehDiubah mengecek apakah status pengajuan masih bisa diubah.
// Hanya pengajuan berstatus "pending" yang bisa di-approve/reject.
func apakahStatusBolehDiubah(statusSekarang string) bool {
	return statusSekarang == "pending"
}

// statusValidUntukPerubahan mengecek apakah status baru valid untuk perubahan.
func statusValidUntukPerubahan(statusBaru string) bool {
	return statusBaru == "approved" || statusBaru == "rejected"
}