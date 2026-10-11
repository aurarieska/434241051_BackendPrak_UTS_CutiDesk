package helper

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Cursor menyimpan posisi terakhir untuk keyset pagination
type Cursor struct {
	WaktuDibuat time.Time
	ID          int
}

// EncodeCursor mengubah pasangan (waktu, id) menjadi string base64url
func EncodeCursor(waktu time.Time, id int) string {
	raw := fmt.Sprintf("%d:%d", waktu.UnixMicro(), id)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor mengurai string base64url menjadi Cursor
// Kembalikan nil, nil jika cursor kosong (halaman pertama)
func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil // halaman pertama
	}

	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, BadRequest("cursor tidak valid")
	}

	bagian := strings.SplitN(string(decoded), ":", 2)
	if len(bagian) != 2 {
		return nil, BadRequest("format cursor tidak valid")
	}

	mikro, err := strconv.ParseInt(bagian[0], 10, 64)
	if err != nil {
		return nil, BadRequest("waktu cursor tidak valid")
	}

	id, err := strconv.Atoi(bagian[1])
	if err != nil {
		return nil, BadRequest("id cursor tidak valid")
	}

	return &Cursor{
		WaktuDibuat: time.UnixMicro(mikro),
		ID:          id,
	}, nil
}