package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"siplab/app/model"
	"siplab/app/repository"
	"siplab/helper"
)

// BookingService menangani pengajuan peminjaman, daftar peminjaman (cursor pagination),
// detail peminjaman (ownership/RBAC), serta verifikasi peminjaman oleh Laboran/Admin.
type BookingService struct {
	repo  repository.BookingRepository
	perms *helper.PermissionSet
}

func NewBookingService(repo repository.BookingRepository, perms *helper.PermissionSet) *BookingService {
	return &BookingService{
		repo:  repo,
		perms: perms,
	}
}

// Create menangani pengajuan peminjaman baru (POST /api/v1/bookings).
func (s *BookingService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Purpose = strings.TrimSpace(req.Purpose)
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	startTime, endTime, ruleErrs := ValidateBookingRequest(req)
	if ruleErrs != nil {
		return helper.Validation(ruleErrs)
	}

	// Bila meminjam ruangan, pastikan tidak bentrok dengan jadwal ruangan yang sudah disetujui
	if req.RoomID != nil {
		conflict, err := s.repo.HasRoomScheduleConflict(ctx, *req.RoomID, startTime, endTime, 0)
		if err != nil {
			return helper.Internal(err)
		}
		if conflict {
			return helper.Conflict("ruangan sudah dibooking dan disetujui pada rentang waktu tersebut")
		}
	}

	created, err := s.repo.Create(ctx, model.Booking{
		UserID:    current.UserID,
		RoomID:    req.RoomID,
		ToolID:    req.ToolID,
		Quantity:  req.Quantity,
		Purpose:   req.Purpose,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    model.BookingStatusPending,
	})
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Created(c, "pengajuan peminjaman berhasil dibuat", created, "/api/v1/bookings/"+itoa(created.ID))
}

// List menampilkan daftar peminjaman dengan cursor pagination (GET /api/v1/bookings).
// Bila pemanggil tidak memiliki permission "booking:list:any", daftar otomatis disaring
// hanya untuk peminjaman milik dirinya sendiri (ownership).
func (s *BookingService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	baseQuery, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	filter := model.BookingFilter{
		After:  baseQuery.After,
		Limit:  baseQuery.Limit,
		Status: strings.TrimSpace(c.Query("status")),
	}

	if !s.perms.Can(current.Role, "booking:list:any") {
		uid := current.UserID
		filter.UserID = &uid
	}

	rows, err := s.repo.FindAfterCursor(ctx, filter)
	if err != nil {
		return helper.Internal(err)
	}

	hasMore := len(rows) > filter.Limit
	if hasMore {
		rows = rows[:filter.Limit]
	}

	meta := &model.CursorMeta{
		Limit:   filter.Limit,
		HasMore: hasMore,
	}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar peminjaman berhasil diambil", rows, meta)
}

// Get menampilkan detail satu peminjaman (GET /api/v1/bookings/:id).
func (s *BookingService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("data peminjaman tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if !CanAccessBooking(current, booking.UserID, s.perms) {
		return helper.Forbidden("tidak berhak melihat data peminjaman milik pengguna lain")
	}

	return helper.Success(c, fiber.StatusOK, "detail peminjaman ditemukan", booking)
}

// Verify memverifikasi status peminjaman oleh Laboran atau Admin (PATCH /api/v1/bookings/:id/verify).
func (s *BookingService) Verify(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.VerifyBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Status = strings.TrimSpace(req.Status)
	req.Notes = strings.TrimSpace(req.Notes)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("data peminjaman tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if !ValidateStatusTransition(existing.Status, req.Status) {
		return helper.Conflict("perubahan status dari " + existing.Status + " ke " + req.Status + " tidak diizinkan")
	}

	if req.Status == model.BookingStatusApproved && existing.RoomID != nil {
		conflict, err := s.repo.HasRoomScheduleConflict(ctx, *existing.RoomID, existing.StartTime, existing.EndTime, existing.ID)
		if err != nil {
			return helper.Internal(err)
		}
		if conflict {
			return helper.Conflict("ruangan sudah memiliki jadwal peminjaman lain yang disetujui pada waktu tersebut")
		}
	}

	updated, err := s.repo.UpdateVerification(ctx, id, req.Status, current.UserID, req.Notes)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("data peminjaman tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "status peminjaman berhasil diverifikasi", updated)
}

// Cancel membatalkan peminjaman milik peminjam sendiri (PATCH /api/v1/bookings/:id/cancel).
// Hanya pemilik peminjaman yang boleh membatalkan, dan hanya selama status masih
// pending atau approved. Perubahan status dicatat ke tabel booking_logs.
func (s *BookingService) Cancel(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("data peminjaman tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if existing.UserID != current.UserID {
		return helper.Forbidden("hanya pemilik peminjaman yang boleh membatalkan")
	}

	if !ValidateStatusTransition(existing.Status, model.BookingStatusCancelled) {
		return helper.Conflict("peminjaman dengan status " + existing.Status + " tidak dapat dibatalkan")
	}

	updated, err := s.repo.Cancel(ctx, id, current.UserID, "dibatalkan oleh peminjam")
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("data peminjaman tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "peminjaman berhasil dibatalkan", updated)
}