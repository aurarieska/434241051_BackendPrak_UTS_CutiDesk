package service

// validasiNamaJenisCutiTidakKosong mengecek apakah nama jenis cuti
// pada PATCH request benar-benar diisi (bukan string kosong).
// Return true jika MELANGGAR aturan.
func namaJenisCutiKosong(nama *string) bool {
	if nama == nil {
		return false // tidak diubah, tidak melanggar
	}
	return *nama == ""
}