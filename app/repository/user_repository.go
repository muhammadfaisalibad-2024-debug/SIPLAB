package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"siplab/app/model"
)

const userColumns = "id, username, email, password_hash, role, is_active, created_at"

// UserRepository mendefinisikan kontrak akses data untuk entitas User.
type UserRepository interface {
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	Create(ctx context.Context, user model.User) (model.User, error)
	Update(ctx context.Context, user model.User) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
	Delete(ctx context.Context, id int) error
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository mengembalikan implementasi PostgreSQL dari UserRepository.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func scanUser(row pgx.Row) (model.User, error) {
	var u model.User
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.IsActive,
		&u.CreatedAt,
	)
	return u, err
}

// FindAfterCursor mengambil satu halaman user menggunakan keyset pagination (created_at DESC, id DESC).
// Mengambil limit+1 baris untuk mendeteksi apakah masih ada halaman berikutnya (has_more) tanpa COUNT(*).
func (r *userPostgresRepository) FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.User, error) {
	args := []any{}
	where := " WHERE 1=1"

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)", len(args), len(args))
	}
	if q.Role != "" {
		args = append(args, q.Role)
		where += fmt.Sprintf(" AND role = $%d", len(args))
	}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND is_active = $%d", len(args))
	}
	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)-1, len(args))
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		"SELECT %s FROM users%s ORDER BY created_at DESC, id DESC LIMIT $%d",
		userColumns, where, len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	result := []model.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca row user: %w", err)
		}
		result = append(result, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query user: %w", err)
	}
	return result, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	u, err := scanUser(r.pool.QueryRow(ctx,
		"SELECT "+userColumns+" FROM users WHERE id = $1", id,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mencari user berdasarkan id: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	u, err := scanUser(r.pool.QueryRow(ctx,
		"SELECT "+userColumns+" FROM users WHERE LOWER(username) = LOWER($1)", username,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mencari user berdasarkan username: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) Create(ctx context.Context, user model.User) (model.User, error) {
	created, err := scanUser(r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, role, is_active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+userColumns,
		user.Username, user.Email, user.PasswordHash, user.Role, user.IsActive,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("menyimpan user baru: %w", err)
	}
	return created, nil
}

func (r *userPostgresRepository) Update(ctx context.Context, user model.User) (model.User, error) {
	updated, err := scanUser(r.pool.QueryRow(ctx,
		`UPDATE users
		 SET username = $1, email = $2, is_active = $3
		 WHERE id = $4
		 RETURNING `+userColumns,
		user.Username, user.Email, user.IsActive, user.ID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("memperbarui user: %w", err)
	}
	return updated, nil
}

// UpdateRole dipisahkan secara eksplisit dari Update biasa agar perubahan role hanya bisa lewat jalur khusus.
func (r *userPostgresRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	updated, err := scanUser(r.pool.QueryRow(ctx,
		"UPDATE users SET role = $1 WHERE id = $2 RETURNING "+userColumns,
		role, id,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengubah role user: %w", err)
	}
	return updated, nil
}

func (r *userPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("menghapus user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
