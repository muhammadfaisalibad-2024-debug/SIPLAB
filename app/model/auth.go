package model

import "time"

// AuthUser menyimpan identitas pengguna yang diekstrak dari JWT ke dalam fiber.Ctx Locals.
type AuthUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// RegisterRequest dipakai pada POST /api/v1/auth/register.
// Perhatikan: TIDAK ADA field Role di sini untuk mencegah serangan Mass Assignment.
// Seluruh pendaftar mandiri otomatis diberi role "mahasiswa" di layer service.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,max=72,nospace,strongpassword"`
}

// LoginRequest dipakai pada POST /api/v1/auth/login.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest dipakai pada POST /api/v1/auth/refresh dan POST /api/v1/auth/logout.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenPairResponse dikirimkan saat login atau refresh berhasil.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         User   `json:"user"`
}

// MeResponse dikirimkan pada GET /api/v1/auth/me agar frontend tahu profil & daftar permission.
type MeResponse struct {
	User        User     `json:"user"`
	Permissions []string `json:"permissions"`
}

// RefreshToken merepresentasikan baris pada tabel refresh_tokens.
type RefreshToken struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
