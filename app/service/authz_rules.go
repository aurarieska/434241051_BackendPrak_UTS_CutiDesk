package service

// apakahPemilik mengecek apakah user yang login adalah pemilik resource.
// Dipakai oleh service untuk data-dependent authorization.
func apakahPemilik(authUserID, pemilikID int) bool {
	return authUserID == pemilikID
}

// bolehLihatPengajuan mengecek apakah user boleh melihat pengajuan tertentu.
// Employee hanya boleh lihat miliknya sendiri. HR dan admin boleh lihat semua.
func bolehLihatPengajuan(role string, authUserID, pemilikPengajuanID int) bool {
	switch role {
	case "admin", "hr":
		return true
	case "employee":
		return authUserID == pemilikPengajuanID
	default:
		// Fail closed: role tidak dikenal → ditolak.
		return false
	}
}