package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================================
// Interface
// =====================================================================

type RoleRepository interface {
	// MuatPermissions mengambil semua mapping role → permission dari database.
	// Hasilnya: map[namaRole][]namaPermission.
	MuatPermissions(ctx context.Context) (map[string][]string, error)

	// CariRoleID mengembalikan ID role berdasarkan nama.
	CariRoleID(ctx context.Context, namaRole string) (int, error)
}

// =====================================================================
// Implementasi PostgreSQL
// =====================================================================

type roleRepositoryPG struct {
	pool *pgxpool.Pool
}

func BuatRoleRepository(pool *pgxpool.Pool) RoleRepository {
	return &roleRepositoryPG{pool: pool}
}

func (r *roleRepositoryPG) MuatPermissions(ctx context.Context) (map[string][]string, error) {
	query := `
		SELECT ro.name AS role_name, p.name AS permission_name
		FROM role_permissions rp
		JOIN roles ro ON ro.id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id
		ORDER BY ro.name, p.name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasil := make(map[string][]string)
	for rows.Next() {
		var namaRole, namaPermission string
		if err := rows.Scan(&namaRole, &namaPermission); err != nil {
			return nil, err
		}
		hasil[namaRole] = append(hasil[namaRole], namaPermission)
	}
	return hasil, rows.Err()
}

func (r *roleRepositoryPG) CariRoleID(ctx context.Context, namaRole string) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM roles WHERE name = $1`, namaRole,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}