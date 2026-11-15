package middleware

import (
	"github.com/gofiber/fiber/v2"
	"siplab/helper"
)

// RequirePermission memeriksa apakah role pengguna yang sedang login memiliki permission yang diminta.
// Menerapkan prinsip fail-closed: bila belum login -> 401, bila tidak punya izin -> 403.
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("autentikasi diperlukan")
		}

		if !perms.Can(user.Role, permission) {
			return helper.Forbidden("anda tidak memiliki izin (" + permission + ") untuk melakukan aksi ini")
		}

		return c.Next()
	}
}