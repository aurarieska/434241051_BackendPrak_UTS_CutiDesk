package service

// tidakBolehNonaktifkanDiriSendiri mengecek apakah admin mencoba
// menonaktifkan akunnya sendiri. Return true jika MELANGGAR aturan.
func tidakBolehNonaktifkanDiriSendiri(authUserID, targetUserID int, isActive *bool) bool {
	if isActive == nil {
		return false // field tidak diubah, tidak melanggar
	}
	// Melanggar jika: target adalah diri sendiri DAN is_active di-set false.
	return authUserID == targetUserID && !*isActive
}

// tidakBolehUbahRoleSendiri mengecek apakah admin mencoba mengubah
// role-nya sendiri. Return true jika MELANGGAR aturan.
func tidakBolehUbahRoleSendiri(authUserID, targetUserID int) bool {
	return authUserID == targetUserID
}