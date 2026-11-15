package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"siplab/helper"
)

// RequireAuth memverifikasi Bearer access token pada header Authorization
// dan menyimpan model.AuthUser ke dalam fiber.Ctx Locals.
func RequireAuth(jwt *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if jwt == nil {
			return helper.Unauthorized("konfigurasi autentikasi belum siap")
		}

		authHeader := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
		if authHeader == "" {
			return helper.Unauthorized("header Authorization wajib disertakan")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			return helper.Unauthorized("format header Authorization harus: Bearer <access_token>")
		}

		user, err := jwt.VerifyAccessToken(strings.TrimSpace(parts[1]))
		if err != nil {
			return helper.Unauthorized("access token tidak valid atau sudah kedaluwarsa")
		}

		helper.SetCurrentUser(c, user)
		return c.Next()
	}
}