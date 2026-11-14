package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"siplab/app/model"
	"siplab/app/repository"
	"siplab/helper"
)

// AuthService menangani siklus autentikasi: Register, Login, Refresh, Logout, dan Me.
// Sesuai Clean Architecture mata kuliah, method yang menerima *fiber.Ctx hanya mem-parsing input,
// memanggil fungsi rules/repository, dan mengembalikan response atau *helper.AppError.
type AuthService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	jwt       *helper.JWTManager
	perms     *helper.PermissionSet
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwt *helper.JWTManager,
	perms *helper.PermissionSet,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwt:       jwt,
		perms:     perms,
	}
}

// Register mendaftarkan akun baru dengan role otomatis "mahasiswa".
func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	passwordHash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(err)
	}

	newUser := BuildRegisteredUser(req, passwordHash)
	created, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("username atau email sudah terdaftar")
		}
		return helper.Internal(err)
	}

	return helper.Created(c, "registrasi berhasil", created, "/api/v1/users/"+itoa(created.ID))
}

// Login memverifikasi kredensial dan menerbitkan pasangan access token (15m) + refresh token (7d).
func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Pesan sengaja disamakan agar penyerang tidak bisa melakukan enumerasi username.
			return helper.Unauthorized("username atau password salah")
		}
		return helper.Internal(err)
	}

	if !helper.CheckPassword(user.PasswordHash, req.Password) {
		return helper.Unauthorized("username atau password salah")
	}

	if !user.IsActive {
		return helper.Forbidden("akun Anda sedang dinonaktifkan")
	}

	accessToken, err := s.jwt.GenerateAccessToken(user)
	if err != nil {
		return helper.Internal(err)
	}

	rawRefresh, refreshHash, expiresAt, err := s.jwt.GenerateRefreshToken()
	if err != nil {
		return helper.Internal(err)
	}

	_, err = s.tokenRepo.Create(ctx, model.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", model.TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
		User:         user,
	})
}

// Refresh melakukan rotasi refresh token: mencabut token lama dan menerbitkan pasangan token baru.
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	tokenHash := helper.HashToken(req.RefreshToken)
	stored, err := s.tokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("refresh token tidak sah atau tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if !IsRefreshTokenUsable(stored, time.Now()) {
		return helper.Unauthorized("refresh token sudah kadaluarsa atau telah dicabut")
	}

	user, err := s.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("pemilik sesi tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if !user.IsActive {
		return helper.Forbidden("akun Anda sedang dinonaktifkan")
	}

	// Rotasi token: cabut refresh token yang baru saja dipakai
	if err := s.tokenRepo.RevokeByHash(ctx, tokenHash); err != nil && !errors.Is(err, repository.ErrNotFound) {
		return helper.Internal(err)
	}

	accessToken, err := s.jwt.GenerateAccessToken(user)
	if err != nil {
		return helper.Internal(err)
	}

	newRawRefresh, newRefreshHash, expiresAt, err := s.jwt.GenerateRefreshToken()
	if err != nil {
		return helper.Internal(err)
	}

	_, err = s.tokenRepo.Create(ctx, model.RefreshToken{
		UserID:    user.ID,
		TokenHash: newRefreshHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "token berhasil diperbarui", model.TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: newRawRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
		User:         user,
	})
}

// Logout mencabut refresh token dari database sehingga tidak dapat dipakai lagi.
func (s *AuthService) Logout(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	tokenHash := helper.HashToken(req.RefreshToken)
	if err := s.tokenRepo.RevokeByHash(ctx, tokenHash); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("refresh token tidak ditemukan atau sudah dicabut")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "logout berhasil", nil)
}

// Me mengembalikan data profil terbaru beserta daftar permission milik role pengguna.
func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	user, err := s.userRepo.FindByID(ctx, current.UserID)
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", model.MeResponse{
		User:        user,
		Permissions: s.perms.PermissionsOf(user.Role),
	})
}
