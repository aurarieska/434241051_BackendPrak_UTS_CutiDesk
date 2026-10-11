package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"cutidesk/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================================
// Interface — kontrak yang harus dipenuhi implementasi repository user
// =====================================================================

type UserRepository interface {
	BuatUser(ctx context.Context, req model.CreateUserRequest, hashPassword string, roleID int) (*model.User, error)
	CariByID(ctx context.Context, id int) (*model.User, error)
	CariByUsername(ctx context.Context, username string) (*model.User, error)
	CariByEmail(ctx context.Context, email string) (*model.User, error)
	DaftarUser(ctx context.Context, limit int, cursor *model.CursorParam, search string) ([]model.User, bool, error)
	UpdateUser(ctx context.Context, id int, req model.UpdateUserRequest) (*model.User, error)
	PatchUser(ctx context.Context, id int, req model.PatchUserRequest) (*model.User, error)
	AssignRole(ctx context.Context, userID int, roleName string) (*model.User, error)
	HitungUser(ctx context.Context) (int, error)
}

// =====================================================================
// Implementasi PostgreSQL
// =====================================================================

type userRepositoryPG struct {
	pool *pgxpool.Pool
}

func BuatUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepositoryPG{pool: pool}
}

func (r *userRepositoryPG) BuatUser(ctx context.Context, req model.CreateUserRequest, hashPassword string, roleID int) (*model.User, error) {
	query := `
		INSERT INTO users (username, email, password, nama_lengkap, role, role_id)
		VALUES (LOWER($1), LOWER($2), $3, $4, $5, $6)
		RETURNING id, username, email, nama_lengkap, role, is_active, created_at, updated_at
	`
	var user model.User
	err := r.pool.QueryRow(ctx, query,
		req.Username, req.Email, hashPassword, req.NamaLengkap, req.Role, roleID,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.NamaLengkap,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, terjemahkanErrorPG(err)
	}
	return &user, nil
}

func (r *userRepositoryPG) CariByID(ctx context.Context, id int) (*model.User, error) {
	query := `
		SELECT id, username, email, password, nama_lengkap, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`
	return r.scanSatuUser(ctx, query, id)
}

func (r *userRepositoryPG) CariByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, nama_lengkap, role, is_active, created_at, updated_at
		FROM users WHERE LOWER(username) = LOWER($1)
	`
	return r.scanSatuUser(ctx, query, username)
}

func (r *userRepositoryPG) CariByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, nama_lengkap, role, is_active, created_at, updated_at
		FROM users WHERE LOWER(email) = LOWER($1)
	`
	return r.scanSatuUser(ctx, query, email)
}

func (r *userRepositoryPG) DaftarUser(ctx context.Context, limit int, cursor *model.CursorParam, search string) ([]model.User, bool, error) {
	args := []interface{}{}
	kondisi := []string{}
	nomor := 1

	if search != "" {
		kondisi = append(kondisi,
			"(LOWER(username) LIKE $"+strconv.Itoa(nomor)+
				" OR LOWER(nama_lengkap) LIKE $"+strconv.Itoa(nomor)+")")
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

	query := `SELECT id, username, email, nama_lengkap, role, is_active, created_at, updated_at
		FROM users ` + where + `
		ORDER BY created_at DESC, id DESC
		LIMIT $` + strconv.Itoa(nomor)

	args = append(args, limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var daftar []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.NamaLengkap,
			&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, false, err
		}
		daftar = append(daftar, u)
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

func (r *userRepositoryPG) UpdateUser(ctx context.Context, id int, req model.UpdateUserRequest) (*model.User, error) {
	query := `
		UPDATE users
		SET email        = LOWER($1),
		    nama_lengkap = $2,
		    is_active    = $3,
		    updated_at   = NOW()
		WHERE id = $4
		RETURNING id, username, email, nama_lengkap, role, is_active, created_at, updated_at
	`
	var user model.User
	err := r.pool.QueryRow(ctx, query, req.Email, req.NamaLengkap, *req.IsActive, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.NamaLengkap,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, terjemahkanErrorPG(err)
	}
	return &user, nil
}

func (r *userRepositoryPG) PatchUser(ctx context.Context, id int, req model.PatchUserRequest) (*model.User, error) {
	sets := []string{}
	args := []interface{}{}
	nomor := 1

	if req.Email != nil {
		sets = append(sets, "email = LOWER($"+strconv.Itoa(nomor)+")")
		args = append(args, *req.Email)
		nomor++
	}
	if req.NamaLengkap != nil {
		sets = append(sets, "nama_lengkap = $"+strconv.Itoa(nomor))
		args = append(args, *req.NamaLengkap)
		nomor++
	}
	if req.IsActive != nil {
		sets = append(sets, "is_active = $"+strconv.Itoa(nomor))
		args = append(args, *req.IsActive)
		nomor++
	}

	if len(sets) == 0 {
		return r.CariByID(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")

	query := `UPDATE users SET ` + strings.Join(sets, ", ") +
		` WHERE id = $` + strconv.Itoa(nomor) +
		` RETURNING id, username, email, nama_lengkap, role, is_active, created_at, updated_at`
	args = append(args, id)

	var user model.User
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Username, &user.Email, &user.NamaLengkap,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, terjemahkanErrorPG(err)
	}
	return &user, nil
}

func (r *userRepositoryPG) AssignRole(ctx context.Context, userID int, roleName string) (*model.User, error) {
	var roleID int
	err := r.pool.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, roleName).Scan(&roleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, err
	}

	query := `
		UPDATE users
		SET role       = $1,
		    role_id    = $2,
		    updated_at = NOW()
		WHERE id = $3
		RETURNING id, username, email, nama_lengkap, role, is_active, created_at, updated_at
	`
	var user model.User
	err = r.pool.QueryRow(ctx, query, roleName, roleID, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.NamaLengkap,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, terjemahkanErrorPG(err)
	}
	return &user, nil
}

func (r *userRepositoryPG) HitungUser(ctx context.Context) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	return total, err
}

// =====================================================================
// Helper internal
// =====================================================================

func (r *userRepositoryPG) scanSatuUser(ctx context.Context, query string, args ...interface{}) (*model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.NamaLengkap, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, err
	}
	return &user, nil
}

func terjemahkanErrorPG(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return model.ErrDuplikat
		}
	}
	return err
}