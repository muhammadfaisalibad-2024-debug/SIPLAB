package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"siplab/app/model"
)

// AccessClaims menyimpan muatan klaim di dalam JWT access token.
type AccessClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager mengelola pembuatan dan verifikasi token autentikasi.
// Access token berumur 15 menit (stateless), sedangkan refresh token berumur 7 hari (disimpan hash-nya di DB).
type JWTManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTManager membuat instance JWTManager dengan TTL bawaan (15 menit access, 7 hari refresh).
func NewJWTManager(secret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &JWTManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *JWTManager) AccessTTL() time.Duration {
	return m.accessTTL
}

func (m *JWTManager) RefreshTTL() time.Duration {
	return m.refreshTTL
}

// GenerateAccessToken menerbitkan JWT access token bertanda tangan HMAC-SHA256.
func (m *JWTManager) GenerateAccessToken(user model.User) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// VerifyAccessToken memvalidasi tanda tangan, algoritma, dan masa berlaku access token.
func (m *JWTManager) VerifyAccessToken(tokenString string) (model.AuthUser, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode penandatanganan token tidak sah")
		}
		return m.secret, nil
	})
	if err != nil {
		return model.AuthUser{}, err
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return model.AuthUser{}, errors.New("klaim token tidak sah")
	}

	return model.AuthUser{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

// GenerateRefreshToken membuat string acak kriptografis 32 byte beserta hash SHA-256-nya.
// Token mentah dikirim ke client, sedangkan hash-nya disimpan di tabel refresh_tokens.
func (m *JWTManager) GenerateRefreshToken() (rawToken string, tokenHash string, expiresAt time.Time, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", time.Time{}, err
	}
	rawToken = hex.EncodeToString(b)
	tokenHash = HashToken(rawToken)
	expiresAt = time.Now().Add(m.refreshTTL)
	return rawToken, tokenHash, expiresAt, nil
}

// HashToken menghitung digest SHA-256 heksadesimal dari refresh token mentah.
func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
