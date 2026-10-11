package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"cutidesk/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================================
// Interface
// =====================================================================

type PengajuanCutiRepository interface {
	Buat(ctx context.Context, userID int, req model.CreatePengajuanCutiRequest) (*model.PengajuanCuti, error)
	CariByID(ctx context.Context, id int) (*model.PengajuanCuti, error)
	Daftar(ctx context.Context, limit int, cursor *model.CursorParam, filter PengajuanFilter) ([]model.PengajuanCuti, bool, error)
	UbahStatus(ctx context.Context, id int, statusBaru string) (*model.PengajuanCuti, error)
}

// PengajuanFilter menyimpan filter opsional untuk daftar pengajuan cuti.
type PengajuanFilter struct {
	UserID      *int    // filter berdasarkan pemilik (employee melihat miliknya sendiri)
	Status      string  // filter berdasarkan status: pending/approved/rejected
	JenisCutiID *int    // filter berdasarkan jenis cuti
	Search      string  // pencarian di nama pengaju / alasan
}

// =====================================================================
// Implementasi PostgreSQL
// =====================================================================

type pengajuanCutiRepositoryPG struct {
	pool *pgxpool.Pool
}

func BuatPengajuanCutiRepository(pool *pgxpool.Pool) PengajuanCutiRepository {
	return &pengajuanCutiRepositoryPG{pool: pool}
}

func (r *pengajuanCutiRepositoryPG) Buat(ctx context.Context, userID int, req model.CreatePengajuanCutiRequest) (*model.PengajuanCuti, error) {
	query := `
		INSERT INTO pengajuan_cuti (user_id, jenis_cuti_id, tanggal_mulai, tanggal_selesai, alasan)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, jenis_cuti_id, tanggal_mulai, tanggal_selesai,
		          alasan, status, created_at, updated_at
	`
	var pc model.PengajuanCuti
	err := r.pool.QueryRow(ctx, query,
		userID, req.JenisCutiID, req.TanggalMulai, req.TanggalSelesai, req.Alasan,
	).Scan(
		&pc.ID, &pc.UserID, &pc.JenisCutiID, &pc.TanggalMulai, &pc.TanggalSelesai,
		&pc.Alasan, &pc.Status, &pc.CreatedAt, &pc.UpdatedAt,
	)
	if err != nil {
		return nil, terjemahkanErrorPG(err)
	}
	return &pc, nil
}

func (r *pengajuanCutiRepositoryPG) CariByID(ctx context.Context, id int) (*model.PengajuanCuti, error) {
	query := `
		SELECT pc.id, pc.user_id, pc.jenis_cuti_id, pc.tanggal_mulai, pc.tanggal_selesai,
		       pc.alasan, pc.status, pc.created_at, pc.updated_at,
		       u.nama_lengkap AS nama_pengaju,
		       jc.nama AS nama_jenis_cuti
		FROM pengajuan_cuti pc
		JOIN users u ON u.id = pc.user_id
		JOIN jenis_cuti jc ON jc.id = pc.jenis_cuti_id
		WHERE pc.id = $1
	`
	var pc model.PengajuanCuti
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pc.ID, &pc.UserID, &pc.JenisCutiID, &pc.TanggalMulai, &pc.TanggalSelesai,
		&pc.Alasan, &pc.Status, &pc.CreatedAt, &pc.UpdatedAt,
		&pc.NamaPengaju, &pc.NamaJenisCuti,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, err
	}
	return &pc, nil
}

func (r *pengajuanCutiRepositoryPG) Daftar(ctx context.Context, limit int, cursor *model.CursorParam, filter PengajuanFilter) ([]model.PengajuanCuti, bool, error) {
	args := []interface{}{}
	kondisi := []string{}
	nomor := 1

	if filter.UserID != nil {
		kondisi = append(kondisi, "pc.user_id = $"+strconv.Itoa(nomor))
		args = append(args, *filter.UserID)
		nomor++
	}

	if filter.Status != "" {
		kondisi = append(kondisi, "pc.status = $"+strconv.Itoa(nomor))
		args = append(args, filter.Status)
		nomor++
	}

	if filter.JenisCutiID != nil {
		kondisi = append(kondisi, "pc.jenis_cuti_id = $"+strconv.Itoa(nomor))
		args = append(args, *filter.JenisCutiID)
		nomor++
	}

	if filter.Search != "" {
		kondisi = append(kondisi,
			"(LOWER(u.nama_lengkap) LIKE $"+strconv.Itoa(nomor)+
				" OR LOWER(pc.alasan) LIKE $"+strconv.Itoa(nomor)+")")
		args = append(args, "%"+strings.ToLower(filter.Search)+"%")
		nomor++
	}

	if cursor != nil {
		kondisi = append(kondisi,
			"(pc.created_at, pc.id) < ($"+strconv.Itoa(nomor)+", $"+strconv.Itoa(nomor+1)+")")
		args = append(args, cursor.WaktuDibuat, cursor.ID)
		nomor += 2
	}

	where := ""
	if len(kondisi) > 0 {
		where = "WHERE " + strings.Join(kondisi, " AND ")
	}

	query := `
		SELECT pc.id, pc.user_id, pc.jenis_cuti_id, pc.tanggal_mulai, pc.tanggal_selesai,
		       pc.alasan, pc.status, pc.created_at, pc.updated_at,
		       u.nama_lengkap AS nama_pengaju,
		       jc.nama AS nama_jenis_cuti
		FROM pengajuan_cuti pc
		JOIN users u ON u.id = pc.user_id
		JOIN jenis_cuti jc ON jc.id = pc.jenis_cuti_id
		` + where + `
		ORDER BY pc.created_at DESC, pc.id DESC
		LIMIT $` + strconv.Itoa(nomor)

	args = append(args, limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var daftar []model.PengajuanCuti
	for rows.Next() {
		var pc model.PengajuanCuti
		if err := rows.Scan(
			&pc.ID, &pc.UserID, &pc.JenisCutiID, &pc.TanggalMulai, &pc.TanggalSelesai,
			&pc.Alasan, &pc.Status, &pc.CreatedAt, &pc.UpdatedAt,
			&pc.NamaPengaju, &pc.NamaJenisCuti,
		); err != nil {
			return nil, false, err
		}
		daftar = append(daftar, pc)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	adaLagi := len(daftar) > limit
	if adaLagi {
		daftar = daftar[:limit]
	}
	return daftar, adaLagi, nil
}

func (r *pengajuanCutiRepositoryPG) UbahStatus(ctx context.Context, id int, statusBaru string) (*model.PengajuanCuti, error) {
	query := `
		UPDATE pengajuan_cuti
		SET status     = $1,
		    updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, jenis_cuti_id, tanggal_mulai, tanggal_selesai,
		          alasan, status, created_at, updated_at
	`
	var pc model.PengajuanCuti
	err := r.pool.QueryRow(ctx, query, statusBaru, id).Scan(
		&pc.ID, &pc.UserID, &pc.JenisCutiID, &pc.TanggalMulai, &pc.TanggalSelesai,
		&pc.Alasan, &pc.Status, &pc.CreatedAt, &pc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, err
	}
	return &pc, nil
}