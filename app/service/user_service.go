package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"siplab/app/model"
	"siplab/app/repository"
	"siplab/helper"
)

// UserService menangani pengelolaan user dengan Cursor Pagination, Ownership Check, dan RBAC.
type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(repo repository.UserRepository, perms *helper.PermissionSet) *UserService {
	return &UserService{
		repo:  repo,
		perms: perms,
	}
}

// List menampilkan daftar user menggunakan cursor pagination (GET /api/v1/users).
func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	// Potong baris ke-(limit+1) yang hanya dipakai sebagai penanda hasMore.
	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	meta := &model.CursorMeta{
		Limit:   q.Limit,
		HasMore: hasMore,
	}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar user berhasil diambil", rows, meta)
}

// Get mengambil detail satu user (GET /api/v1/users/:id).
// Memeriksa kepemilikan (ownership) atau permission "user:read:any" SEBELUM query ke database
// untuk mencegah timing attack enumerasi ID.
func (s *UserService) Get(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Forbidden("tidak berhak mengakses data user lain")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "user ditemukan", user)
}

// Patch memperbarui sebagian data user (PATCH /api/v1/users/:id).
func (s *UserService) Patch(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Forbidden("tidak berhak mengubah data user lain")
	}

	var req model.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyUserPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(err)
	}

	updatedPayload := ApplyUserPatch(existing, req)
	result, err := s.repo.Update(ctx, updatedPayload)
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diperbarui", result)
}

// AssignRole mengubah role milik user lain (PATCH /api/v1/users/:id/role).
// Dijaga oleh middleware RequirePermission("role:assign") dan aturan bisnis anti-self-demotion.
func (s *UserService) AssignRole(c *fiber.Ctx) error {
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

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Role = strings.TrimSpace(req.Role)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if ruleErrs := ValidateAssignRole(current, id, req, s.perms); len(ruleErrs) > 0 {
		return helper.Validation(ruleErrs)
	}

	result, err := s.repo.UpdateRole(ctx, id, req.Role)
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "role user berhasil diubah", result)
}

// Delete menghapus akun user lain (DELETE /api/v1/users/:id).
func (s *UserService) Delete(c *fiber.Ctx) error {
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

	if current.UserID == id {
		return helper.Forbidden("tidak boleh menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateUserError(err)
	}

	return helper.NoContent(c)
}

// translateUserError memetakan sentinel error repository ke AppError tanpa membocorkan detail SQL.
func translateUserError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("user tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("username atau email sudah dipakai")
	default:
		return helper.Internal(err)
	}
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
