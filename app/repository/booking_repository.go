package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siplab/app/model"
)

// BookingRepository mendefinisikan kontrak akses data peminjaman laboratorium.
type BookingRepository interface {
	Create(ctx context.Context, b model.Booking) (model.Booking, error)
	FindByID(ctx context.Context, id int) (model.Booking, error)
	FindAfterCursor(ctx context.Context, filter model.BookingFilter) ([]model.Booking, error)
	UpdateVerification(ctx context.Context, id int, status string, verifiedBy int, notes string) (model.Booking, error)
	Cancel(ctx context.Context, id int, actorID int, note string) (model.Booking, error)
	HasRoomScheduleConflict(ctx context.Context, roomID int, startTime, endTime time.Time, excludeBookingID int) (bool, error)
}

type pgBookingRepository struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) BookingRepository {
	return &pgBookingRepository{pool: pool}
}

func (r *pgBookingRepository) Create(ctx context.Context, b model.Booking) (model.Booking, error) {
	const q = `
		INSERT INTO bookings (user_id, room_id, tool_id, quantity, purpose, start_time, end_time, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, room_id, tool_id, quantity, purpose, start_time, end_time, status, verified_by, notes, created_at
	`
	var out model.Booking
	err := r.pool.QueryRow(ctx, q,
		b.UserID,
		b.RoomID,
		b.ToolID,
		b.Quantity,
		b.Purpose,
		b.StartTime,
		b.EndTime,
		b.Status,
		b.Notes,
	).Scan(
		&out.ID,
		&out.UserID,
		&out.RoomID,
		&out.ToolID,
		&out.Quantity,
		&out.Purpose,
		&out.StartTime,
		&out.EndTime,
		&out.Status,
		&out.VerifiedBy,
		&out.Notes,
		&out.CreatedAt,
	)
	if err != nil {
		return model.Booking{}, err
	}
	return out, nil
}

func (r *pgBookingRepository) FindByID(ctx context.Context, id int) (model.Booking, error) {
	const q = `
		SELECT id, user_id, room_id, tool_id, quantity, purpose, start_time, end_time, status, verified_by, notes, created_at
		FROM bookings
		WHERE id = $1
	`
	var out model.Booking
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&out.ID,
		&out.UserID,
		&out.RoomID,
		&out.ToolID,
		&out.Quantity,
		&out.Purpose,
		&out.StartTime,
		&out.EndTime,
		&out.Status,
		&out.VerifiedBy,
		&out.Notes,
		&out.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Booking{}, ErrNotFound
		}
		return model.Booking{}, err
	}
	return out, nil
}

func (r *pgBookingRepository) FindAfterCursor(ctx context.Context, f model.BookingFilter) ([]model.Booking, error) {
	query := `
		SELECT id, user_id, room_id, tool_id, quantity, purpose, start_time, end_time, status, verified_by, notes, created_at
		FROM bookings
		WHERE 1=1
	`
	args := []any{}
	argPos := 1

	if f.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *f.UserID)
		argPos++
	}

	if f.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, f.Status)
		argPos++
	}

	if f.After != nil {
		query += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", argPos, argPos+1)
		args = append(args, f.After.CreatedAt, f.After.ID)
		argPos += 2
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", argPos)
	args = append(args, f.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Booking, 0, f.Limit+1)
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(
			&b.ID,
			&b.UserID,
			&b.RoomID,
			&b.ToolID,
			&b.Quantity,
			&b.Purpose,
			&b.StartTime,
			&b.EndTime,
			&b.Status,
			&b.VerifiedBy,
			&b.Notes,
			&b.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (r *pgBookingRepository) UpdateVerification(ctx context.Context, id int, status string, verifiedBy int, notes string) (model.Booking, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Booking{}, err
	}
	defer tx.Rollback(ctx)

	fromStatus, err := lockBookingStatus(ctx, tx, id)
	if err != nil {
		return model.Booking{}, err
	}

	const q = `
		UPDATE bookings
		SET status = $2, verified_by = $3, notes = $4
		WHERE id = $1
		RETURNING id, user_id, room_id, tool_id, quantity, purpose, start_time, end_time, status, verified_by, notes, created_at
	`
	var out model.Booking
	err = tx.QueryRow(ctx, q, id, status, verifiedBy, notes).Scan(
		&out.ID,
		&out.UserID,
		&out.RoomID,
		&out.ToolID,
		&out.Quantity,
		&out.Purpose,
		&out.StartTime,
		&out.EndTime,
		&out.Status,
		&out.VerifiedBy,
		&out.Notes,
		&out.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Booking{}, ErrNotFound
		}
		return model.Booking{}, err
	}

	if err := insertBookingLog(ctx, tx, id, verifiedBy, fromStatus, status, notes); err != nil {
		return model.Booking{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Booking{}, err
	}
	return out, nil
}

// Cancel membatalkan peminjaman milik peminjam dan mencatat perubahannya ke booking_logs.
func (r *pgBookingRepository) Cancel(ctx context.Context, id int, actorID int, note string) (model.Booking, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Booking{}, err
	}
	defer tx.Rollback(ctx)

	fromStatus, err := lockBookingStatus(ctx, tx, id)
	if err != nil {
		return model.Booking{}, err
	}

	const q = `
		UPDATE bookings
		SET status = 'cancelled'
		WHERE id = $1
		RETURNING id, user_id, room_id, tool_id, quantity, purpose, start_time, end_time, status, verified_by, notes, created_at
	`
	var out model.Booking
	err = tx.QueryRow(ctx, q, id).Scan(
		&out.ID,
		&out.UserID,
		&out.RoomID,
		&out.ToolID,
		&out.Quantity,
		&out.Purpose,
		&out.StartTime,
		&out.EndTime,
		&out.Status,
		&out.VerifiedBy,
		&out.Notes,
		&out.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Booking{}, ErrNotFound
		}
		return model.Booking{}, err
	}

	if err := insertBookingLog(ctx, tx, id, actorID, fromStatus, model.BookingStatusCancelled, note); err != nil {
		return model.Booking{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Booking{}, err
	}
	return out, nil
}

// lockBookingStatus mengambil status terkini dengan row lock agar dua perubahan status
// yang bersamaan tidak saling menimpa (mencegah race condition).
func lockBookingStatus(ctx context.Context, tx pgx.Tx, id int) (string, error) {
	var status string
	err := tx.QueryRow(ctx, `SELECT status FROM bookings WHERE id = $1 FOR UPDATE`, id).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return status, nil
}

// insertBookingLog menulis satu baris audit perubahan status peminjaman.
func insertBookingLog(ctx context.Context, tx pgx.Tx, bookingID, actorID int, fromStatus, toStatus, note string) error {
	const q = `
		INSERT INTO booking_logs (booking_id, actor_id, from_status, to_status, note)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.Exec(ctx, q, bookingID, actorID, fromStatus, toStatus, note)
	return err
}

func (r *pgBookingRepository) HasRoomScheduleConflict(ctx context.Context, roomID int, startTime, endTime time.Time, excludeBookingID int) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM bookings
			WHERE room_id = $1
			  AND id <> $4
			  AND status = 'approved'
			  AND start_time < $3
			  AND end_time > $2
		)
	`
	var conflict bool
	err := r.pool.QueryRow(ctx, q, roomID, startTime, endTime, excludeBookingID).Scan(&conflict)
	return conflict, err
}