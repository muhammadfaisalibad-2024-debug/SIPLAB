package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siplab/app/model"
)

// TokenRepository mengelola penyimpanan dan pencabutan refresh token di database.
type TokenRepository interface {
	Create(ctx context.Context, token model.RefreshToken) (model.RefreshToken, error)
	FindByHash(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	RevokeByHash(ctx context.Context, tokenHash string) error
	RevokeAllByUser(ctx context.Context, userID int) error
}

type tokenPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &tokenPostgresRepository{pool: pool}
}

func (r *tokenPostgresRepository) Create(ctx context.Context, token model.RefreshToken) (model.RefreshToken, error) {
	var saved model.RefreshToken
	err := r.pool.QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at`,
		token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(
		&saved.ID, &saved.UserID, &saved.TokenHash,
		&saved.ExpiresAt, &saved.RevokedAt, &saved.CreatedAt,
	)
	if err != nil {
		return model.RefreshToken{}, fmt.Errorf("menyimpan refresh token: %w", err)
	}
	return saved, nil
}

func (r *tokenPostgresRepository) FindByHash(ctx context.Context, tokenHash string) (model.RefreshToken, error) {
	var t model.RefreshToken
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		 FROM refresh_tokens
		 WHERE token_hash = $1`,
		tokenHash,
	).Scan(
		&t.ID, &t.UserID, &t.TokenHash,
		&t.ExpiresAt, &t.RevokedAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrNotFound
		}
		return model.RefreshToken{}, fmt.Errorf("mencari refresh token: %w", err)
	}
	return t, nil
}

func (r *tokenPostgresRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW()
		 WHERE token_hash = $1 AND revoked_at IS NULL`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("mencabut refresh token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *tokenPostgresRepository) RevokeAllByUser(ctx context.Context, userID int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW()
		 WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("mencabut seluruh token user: %w", err)
	}
	return nil
}
