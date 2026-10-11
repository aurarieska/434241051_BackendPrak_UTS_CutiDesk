package model

import "time"

// CursorParam digunakan repository untuk menerima parameter cursor pagination.
// Struct ini ada di model agar repository tidak perlu import helper.
type CursorParam struct {
	WaktuDibuat time.Time
	ID          int
}