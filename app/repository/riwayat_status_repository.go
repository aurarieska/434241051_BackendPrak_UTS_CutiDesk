package repository

import (
	"context"

	"cutidesk/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================================
// Interface
// =====================================================================

type RiwayatStatusRepository interface {
	// Buat mencatat perubahan status pada pengajuan cuti.
	Buat(ctx context.Context, pengajuanID int, diubahOleh int, statusSebelum *string, statusSesudah string, catatan string) (*model.RiwayatStatus, error)

	// DaftarByPengajuan mengembalikan seluruh riwayat status suatu pengajuan (urut waktu ASC).
	DaftarByPengajuan(ctx context.Context, pengajuanID int) ([]model.RiwayatStatus, error)
}

// =====================================================================
// Implementasi PostgreSQL
// =====================================================================

type riwayatStatusRepositoryPG struct {
	pool *pgxpool.Pool
}

func BuatRiwayatStatusRepository(pool *pgxpool.Pool) RiwayatStatusRepository {
	return &riwayatStatusRepositoryPG{pool: pool}
}

func (r *riwayatStatusRepositoryPG) Buat(ctx context.Context, pengajuanID int, diubahOleh int, statusSebelum *string, statusSesudah string, catatan string) (*model.RiwayatStatus, error) {
	query := `
		INSERT INTO riwayat_status (pengajuan_cuti_id, diubah_oleh, status_sebelum, status_sesudah, catatan)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, pengajuan_cuti_id, diubah_oleh, status_sebelum, status_sesudah, catatan, created_at
	`
	var rs model.RiwayatStatus
	err := r.pool.QueryRow(ctx, query,
		pengajuanID, diubahOleh, statusSebelum, statusSesudah, catatan,
	).Scan(
		&rs.ID, &rs.PengajuanCutiID, &rs.DiubahOleh,
		&rs.StatusSebelum, &rs.StatusSesudah, &rs.Catatan, &rs.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

func (r *riwayatStatusRepositoryPG) DaftarByPengajuan(ctx context.Context, pengajuanID int) ([]model.RiwayatStatus, error) {
	query := `
		SELECT rs.id, rs.pengajuan_cuti_id, rs.diubah_oleh,
		       rs.status_sebelum, rs.status_sesudah, rs.catatan, rs.created_at,
		       u.nama_lengkap AS nama_pengubah
		FROM riwayat_status rs
		JOIN users u ON u.id = rs.diubah_oleh
		WHERE rs.pengajuan_cuti_id = $1
		ORDER BY rs.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, pengajuanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daftar []model.RiwayatStatus
	for rows.Next() {
		var rs model.RiwayatStatus
		if err := rows.Scan(
			&rs.ID, &rs.PengajuanCutiID, &rs.DiubahOleh,
			&rs.StatusSebelum, &rs.StatusSesudah, &rs.Catatan, &rs.CreatedAt,
			&rs.NamaPengubah,
		); err != nil {
			return nil, err
		}
		daftar = append(daftar, rs)
	}
	return daftar, rows.Err()
}