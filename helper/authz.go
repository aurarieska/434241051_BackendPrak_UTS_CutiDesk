package helper

// SetPermission menyimpan pemetaan role → daftar permission
// Dimuat sekali saat aplikasi menyala, dari database
type SetPermission struct {
	peta map[string]map[string]bool // role → permission → true
}

// BuatSetPermission membuat SetPermission dari data yang sudah dimuat
// Format input: map[role] = []permission
func BuatSetPermission(data map[string][]string) *SetPermission {
	peta := make(map[string]map[string]bool)
	for role, permissions := range data {
		peta[role] = make(map[string]bool)
		for _, p := range permissions {
			peta[role][p] = true
		}
	}
	return &SetPermission{peta: peta}
}

// Boleh memeriksa apakah role memiliki permission tertentu
// Fail closed: role tidak dikenal → ditolak
func (sp *SetPermission) Boleh(role, permission string) bool {
	perms, ada := sp.peta[role]
	if !ada {
		return false
	}
	return perms[permission]
}

// DaftarPermission mengembalikan semua permission milik role tertentu
func (sp *SetPermission) DaftarPermission(role string) []string {
	perms, ada := sp.peta[role]
	if !ada {
		return nil
	}
	hasil := make([]string, 0, len(perms))
	for p := range perms {
		hasil = append(hasil, p)
	}
	return hasil
}

// RoleDikenal memeriksa apakah nama role terdaftar dalam sistem
func (sp *SetPermission) RoleDikenal(role string) bool {
	_, ada := sp.peta[role]
	return ada
}