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

type JenisCutiRepository interface {
	Buat(ctx context.Context, req model.CreateJenisCutiRequest) (*model.JenisCuti, error)
	CariByID(ctx context.Context, id int) (*model.JenisCuti, error)
	Daftar(ctx context.Context, limit int, cursor *model.CursorParam, search string) ([]model.JenisCuti, bool, error)
	Update(ctx context.Context, id int, req model.UpdateJenisCutiRequest) (*model.JenisCuti, error)
	Patch(ctx context.Context, id int, req model.PatchJenisCutiRequest) (*model.JenisCuti, error)
	Hapus(ctx context.Context, id int) error
}

// =====================================================================
// Implementasi PostgreSQL
// =====================================================================

type jenisCutiRepositoryPG struct {
	pool *pgxpool.Pool
}

func BuatJenisCutiRepository(pool *pgxpool.Pool) JenisCutiRepository {
	return &jenisCutiRepositoryPG{pool: pool}
}

func (r *jenisCutiRepositoryPG) Buat(ctx context.Context, req model.CreateJenisCutiRequest) (*model.JenisCuti, error) {
	query := `
		INSERT INTO jenis_cuti (nama, deskripsi)
		VALUES ($1, $2)
		RETURNING id, nama, deskripsi, created_at, updated_at
	`
	var jc model.JenisCuti
	err := r.pool.QueryRow(ctx, query, req.Nama, req.Deskripsi).Scan(
		&jc.ID, &jc.Nama, &jc.Deskripsi, &jc.CreatedAt, &jc.UpdatedAt,
	)
	if err != nil {
		return nil, terjemahkanErrorPG(err)
	}
	return &jc, nil
}

func (r *jenisCutiRepositoryPG) CariByID(ctx context.Context, id int) (*model.JenisCuti, error) {
	query := `
		SELECT id, nama, deskripsi, created_at, updated_at
		FROM jenis_cuti WHERE id = $1
	`
	var jc model.JenisCuti
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&jc.ID, &jc.Nama, &jc.Deskripsi, &jc.CreatedAt, &jc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, err
	}
	return &jc, nil
}

func (r *jenisCutiRepositoryPG) Daftar(ctx context.Context, limit int, cursor *model.CursorParam, search string) ([]model.JenisCuti, bool, error) {
	args := []interface{}{}
	kondisi := []string{}
	nomor := 1

	if search != "" {
		kondisi = append(kondisi, "LOWER(nama) LIKE $"+strconv.Itoa(nomor))
		args = append(args, "%"+strings.ToLower(search)+"%")
		nomor++
	}

	if cursor != nil {
		kondisi = append(kondisi,
			"(created_at, id) < ($"+strconv.Itoa(nomor)+", $"+strconv.Itoa(nomor+1)+")")
		args = append(args, cursor.WaktuDibuat, cursor.ID)
		nomor += 2
	}

	where := ""
	if len(kondisi) > 0 {
		where = "WHERE " + strings.Join(kondisi, " AND ")
	}

	query := `SELECT id, nama, deskripsi, created_at, updated_at
		FROM jenis_cuti ` + where + `
		ORDER BY created_at DESC, id DESC
		LIMIT $` + strconv.Itoa(nomor)

	args = append(args, limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var daftar []model.JenisCuti
	for rows.Next() {
		var jc model.JenisCuti
		if err := rows.Scan(&jc.ID, &jc.Nama, &jc.Deskripsi, &jc.CreatedAt, &jc.UpdatedAt); err != nil {
			return nil, false, err
		}
		daftar = append(daftar, jc)
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

func (r *jenisCutiRepositoryPG) Update(ctx context.Context, id int, req model.UpdateJenisCutiRequest) (*model.JenisCuti, error) {
	query := `
		UPDATE jenis_cuti
		SET nama       = $1,
		    deskripsi   = $2,
		    updated_at  = NOW()
		WHERE id = $3
		RETURNING id, nama, deskripsi, created_at, updated_at
	`
	var jc model.JenisCuti
	err := r.pool.QueryRow(ctx, query, req.Nama, req.Deskripsi, id).Scan(
		&jc.ID, &jc.Nama, &jc.Deskripsi, &jc.CreatedAt, &jc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, terjemahkanErrorPG(err)
	}
	return &jc, nil
}

func (r *jenisCutiRepositoryPG) Patch(ctx context.Context, id int, req model.PatchJenisCutiRequest) (*model.JenisCuti, error) {
	sets := []string{}
	args := []interface{}{}
	nomor := 1

	if req.Nama != nil {
		sets = append(sets, "nama = $"+strconv.Itoa(nomor))
		args = append(args, *req.Nama)
		nomor++
	}
	if req.Deskripsi != nil {
		sets = append(sets, "deskripsi = $"+strconv.Itoa(nomor))
		args = append(args, *req.Deskripsi)
		nomor++
	}

	if len(sets) == 0 {
		return r.CariByID(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")

	query := `UPDATE jenis_cuti SET ` + strings.Join(sets, ", ") +
		` WHERE id = $` + strconv.Itoa(nomor) +
		` RETURNING id, nama, deskripsi, created_at, updated_at`
	args = append(args, id)

	var jc model.JenisCuti
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&jc.ID, &jc.Nama, &jc.Deskripsi, &jc.CreatedAt, &jc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, terjemahkanErrorPG(err)
	}
	return &jc, nil
}

func (r *jenisCutiRepositoryPG) Hapus(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM jenis_cuti WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrTidakDitemukan
	}
	return nil
}