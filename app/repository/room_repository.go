package repository

import (
	"context"
	"fmt"

	"siplab/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	pool *pgxpool.Pool
}

func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (r *RoomRepository) GetAllRooms(ctx context.Context, search string, limit, offset int) ([]model.Room, int, error) {
	query := `SELECT id, code, name, capacity, location, status, created_at, updated_at 
			  FROM rooms 
			  WHERE name ILIKE $1 OR code ILIKE $1 
			  ORDER BY id ASC LIMIT $2 OFFSET $3`
	
	searchParam := "%" + search + "%"
	rows, err := r.pool.Query(ctx, query, searchParam, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data ruangan: %w", err)
	}
	defer rows.Close()

	var rooms []model.Room
	for rows.Next() {
		var rm model.Room
		if err := rows.Scan(&rm.ID, &rm.Code, &rm.Name, &rm.Capacity, &rm.Location, &rm.Status, &rm.CreatedAt, &rm.UpdatedAt); err != nil {
			return nil, 0, err
		}
		rooms = append(rooms, rm)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM rooms WHERE name ILIKE $1 OR code ILIKE $1`
	err = r.pool.QueryRow(ctx, countQuery, searchParam).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

func (r *RoomRepository) CreateRoom(ctx context.Context, req *model.CreateRoomRequest) (*model.Room, error) {
	query := `INSERT INTO rooms (code, name, capacity, location, status) 
			  VALUES ($1, $2, $3, $4, $5) 
			  RETURNING id, code, name, capacity, location, status, created_at, updated_at`

	var rm model.Room
	err := r.pool.QueryRow(ctx, query, req.Code, req.Name, req.Capacity, req.Location, req.Status).
		Scan(&rm.ID, &rm.Code, &rm.Name, &rm.Capacity, &rm.Location, &rm.Status, &rm.CreatedAt, &rm.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("gagal menambah ruangan: %w", err)
	}

	return &rm, nil
}

func (r *RoomRepository) GetAllTools(ctx context.Context, search string, limit, offset int) ([]model.Tool, int, error) {
	query := `SELECT id, room_id, code, name, total_quantity, available_quantity, condition, created_at, updated_at 
			  FROM tools 
			  WHERE name ILIKE $1 OR code ILIKE $1 
			  ORDER BY id ASC LIMIT $2 OFFSET $3`

	searchParam := "%" + search + "%"
	rows, err := r.pool.Query(ctx, query, searchParam, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data peralatan: %w", err)
	}
	defer rows.Close()

	var tools []model.Tool
	for rows.Next() {
		var tl model.Tool
		if err := rows.Scan(&tl.ID, &tl.RoomID, &tl.Code, &tl.Name, &tl.TotalQuantity, &tl.AvailableQuantity, &tl.Condition, &tl.CreatedAt, &tl.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tools = append(tools, tl)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM tools WHERE name ILIKE $1 OR code ILIKE $1`
	err = r.pool.QueryRow(ctx, countQuery, searchParam).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return tools, total, nil
}

func (r *RoomRepository) CreateTool(ctx context.Context, req *model.CreateToolRequest) (*model.Tool, error) {
	query := `INSERT INTO tools (room_id, code, name, total_quantity, available_quantity, condition) 
			  VALUES ($1, $2, $3, $4, $4, $5) 
			  RETURNING id, room_id, code, name, total_quantity, available_quantity, condition, created_at, updated_at`

	var tl model.Tool
	err := r.pool.QueryRow(ctx, query, req.RoomID, req.Code, req.Name, req.TotalQuantity, req.Condition).
		Scan(&tl.ID, &tl.RoomID, &tl.Code, &tl.Name, &tl.TotalQuantity, &tl.AvailableQuantity, &tl.Condition, &tl.CreatedAt, &tl.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("gagal menambah peralatan: %w", err)
	}

	return &tl, nil
}