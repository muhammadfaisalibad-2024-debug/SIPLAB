package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"siplab/app/model"
	"siplab/app/repository"
	"siplab/helper"
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
		return nil, errors.New("kode dan nama ruangan tidak boleh kosong")
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

// List menangani GET /api/v1/rooms (daftar ruang lab berpaginasi).
func (s *RoomService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	rooms, meta, err := s.FetchRooms(ctx, strings.TrimSpace(c.Query("search")), queryInt(c, "page", 1), queryInt(c, "limit", 10))
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "daftar ruang laboratorium berhasil diambil", fiber.Map{
		"items":      rooms,
		"pagination": meta,
	})
}

// Get menangani GET /api/v1/rooms/:id (detail satu ruang lab).
func (s *RoomService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	room, err := s.roomRepo.GetRoomByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("data ruangan tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "detail ruangan ditemukan", room)
}

// Create menangani POST /api/v1/rooms (khusus Admin/Laboran dengan permission room:create).
func (s *RoomService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.Location = strings.TrimSpace(req.Location)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	room, err := s.AddRoom(ctx, &req)
	if err != nil {
		if isUniqueViolation(err) {
			return helper.Conflict("kode ruangan sudah dipakai")
		}
		return helper.BadRequest(err.Error())
	}

	return helper.Created(c, "ruang laboratorium berhasil ditambahkan", room, "/api/v1/rooms/"+strconv.Itoa(room.ID))
}

// Replace menangani PUT /api/v1/rooms/:id (khusus Admin/Laboran dengan permission room:update).
func (s *RoomService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Location = strings.TrimSpace(req.Location)

	if req.Name == "" {
		return helper.Validation(map[string]string{"name": "nama ruangan tidak boleh kosong"})
	}
	if req.Capacity <= 0 {
		return helper.Validation(map[string]string{"capacity": "kapasitas ruangan harus lebih besar dari 0"})
	}
	if req.Status == "" {
		req.Status = "AVAILABLE"
	}

	room, err := s.roomRepo.UpdateRoom(ctx, id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("data ruangan tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "ruang laboratorium berhasil diperbarui", room)
}

// Delete menangani DELETE /api/v1/rooms/:id (khusus Admin dengan permission room:delete).
func (s *RoomService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.roomRepo.DeleteRoom(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("data ruangan tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}

// Availability menangani GET /api/v1/rooms/availability?date=YYYY-MM-DD.
func (s *RoomService) Availability(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	dateQuery := strings.TrimSpace(c.Query("date"))
	if dateQuery == "" {
		dateQuery = time.Now().Format("2006-01-02")
	}

	day, err := time.ParseInLocation("2006-01-02", dateQuery, time.Local)
	if err != nil {
		return helper.BadRequest("format tanggal harus YYYY-MM-DD")
	}

	items, err := s.roomRepo.GetAvailability(ctx, day, day.AddDate(0, 0, 1))
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "ketersediaan ruang laboratorium berhasil diambil", fiber.Map{
		"date":  dateQuery,
		"items": items,
	})
}

// ListTools menangani GET /api/v1/tools (daftar alat lab berpaginasi).
func (s *RoomService) ListTools(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	tools, meta, err := s.FetchTools(ctx, strings.TrimSpace(c.Query("search")), queryInt(c, "page", 1), queryInt(c, "limit", 10))
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "daftar alat laboratorium berhasil diambil", fiber.Map{
		"items":      tools,
		"pagination": meta,
	})
}

// CreateTool menangani POST /api/v1/tools (khusus Admin/Laboran dengan permission tool:create).
func (s *RoomService) CreateTool(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateToolRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)

	tool, err := s.AddTool(ctx, &req)
	if err != nil {
		if isUniqueViolation(err) {
			return helper.Conflict("kode peralatan sudah dipakai")
		}
		return helper.BadRequest(err.Error())
	}

	return helper.Created(c, "alat laboratorium berhasil ditambahkan", tool, "/api/v1/tools/"+strconv.Itoa(tool.ID))
}

// queryInt membaca query string bertipe angka dengan nilai bawaan bila tidak valid.
func queryInt(c *fiber.Ctx, key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(c.Query(key)))
	if err != nil {
		return fallback
	}
	return value
}

// isUniqueViolation mendeteksi pelanggaran constraint UNIQUE dari PostgreSQL.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
