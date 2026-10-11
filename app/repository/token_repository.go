package repository

import (
	"context"
	"errors"
	"time"

	"cutidesk/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================================
// Interface
// =====================================================================

type TokenRepository interface {
	// SimpanRefreshToken menyimpan hash refresh token ke database.
	SimpanRefreshToken(ctx context.Context, userID int, tokenHash string, expirasiPada time.Time) error

	// CariByHash mencari refresh token berdasarkan hash SHA-256.
	CariByHash(ctx context.Context, tokenHash string) (*model.RefreshTokenData, error)

	// CabutByHash menandai token sebagai dicabut (revoked).
	CabutByHash(ctx context.Context, tokenHash string) error

	// CabutSemuaByUser mencabut semua refresh token milik user tertentu (untuk logout all devices).
	CabutSemuaByUser(ctx context.Context, userID int) error

	// HapusTokenKedaluwarsa membersihkan token yang sudah kedaluwarsa (housekeeping).
	HapusTokenKedaluwarsa(ctx context.Context) (int64, error)
}

// =====================================================================
// Implementasi PostgreSQL
// =====================================================================

type tokenRepositoryPG struct {
	pool *pgxpool.Pool
}

func BuatTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &tokenRepositoryPG{pool: pool}
}

func (r *tokenRepositoryPG) SimpanRefreshToken(ctx context.Context, userID int, tokenHash string, expirasiPada time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(ctx, query, userID, tokenHash, expirasiPada)
	return err
}

func (r *tokenRepositoryPG) CariByHash(ctx context.Context, tokenHash string) (*model.RefreshTokenData, error) {
	query := `
		SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var data model.RefreshTokenData
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&data.ID, &data.UserID, &data.ExpiresAt, &data.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTidakDitemukan
		}
		return nil, err
	}
	return &data, nil
}

func (r *tokenRepositoryPG) CabutByHash(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	tag, err := r.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrTidakDitemukan
	}
	return nil
}

func (r *tokenRepositoryPG) CabutSemuaByUser(ctx context.Context, userID int) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *tokenRepositoryPG) HapusTokenKedaluwarsa(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}