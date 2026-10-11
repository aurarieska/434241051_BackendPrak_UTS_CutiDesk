package model

import "errors"

var (
	ErrTidakDitemukan   = errors.New("data tidak ditemukan")
	ErrDuplikat         = errors.New("data sudah ada")
	ErrTokenKedaluwarsa = errors.New("token kedaluwarsa")
	ErrTokenDicabut     = errors.New("token sudah dicabut")
)