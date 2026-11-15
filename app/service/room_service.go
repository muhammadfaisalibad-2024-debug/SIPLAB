package service

import (
	"context"
	"errors"

	"siplab/app/model"
	"siplab/app/repository"
)

type RoomService struct {
	roomRepo *repository.RoomRepository
}

func NewRoomService(roomRepo *repository.RoomRepository) *RoomService {
	return &RoomService{roomRepo: roomRepo}
}

func (s *RoomService) FetchRooms(ctx context.Context, search string, page, limit int) ([]model.Room, map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rooms, total, err := s.roomRepo.GetAllRooms(ctx, search, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	meta := map[string]interface{}{
		"page":        page,
		"limit":       limit,
		"total_data":  total,
		"total_pages": (total + limit - 1) / limit,
	}

	return rooms, meta, nil
}

func (s *RoomService) AddRoom(ctx context.Context, req *model.CreateRoomRequest) (*model.Room, error) {
	if req.Code == "" || req.Name == "" {
		return nil, errors.Println("kode dan nama ruangan tidak boleh kosong")
	}
	if req.Capacity <= 0 {
		return nil, errors.New("kapasitas ruangan harus lebih besar dari 0")
	}
	if req.Status == "" {
		req.Status = "AVAILABLE"
	}
	return s.roomRepo.CreateRoom(ctx, req)
}

func (s *RoomService) FetchTools(ctx context.Context, search string, page, limit int) ([]model.Tool, map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	tools, total, err := s.roomRepo.GetAllTools(ctx, search, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	meta := map[string]interface{}{
		"page":        page,
		"limit":       limit,
		"total_data":  total,
		"total_pages": (total + limit - 1) / limit,
	}

	return tools, meta, nil
}

func (s *RoomService) AddTool(ctx context.Context, req *model.CreateToolRequest) (*model.Tool, error) {
	if req.Code == "" || req.Name == "" {
		return nil, errors.New("kode dan nama peralatan tidak boleh kosong")
	}
	if req.TotalQuantity <= 0 {
		return nil, errors.New("jumlah total peralatan harus lebih besar dari 0")
	}
	if req.Condition == "" {
		req.Condition = "GOOD"
	}
	return s.roomRepo.CreateTool(ctx, req)
}