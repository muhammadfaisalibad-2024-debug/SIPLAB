package helper

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"siplab/app/model"
)

const authUserContextKey = "auth_user"

// RequestContext membungkus UserContext dari Fiber dengan batas waktu 5 detik
// agar operasi database tidak menggantung tanpa batas.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ParamID mengekstrak parameter URL :id dan memastikan nilainya bilangan bulat positif (> 0).
func ParamID(c *fiber.Ctx) (int, bool) {
	raw := c.Params("id")
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// SetCurrentUser menyimpan identitas pengguna terautentikasi ke dalam fiber.Ctx Locals.
func SetCurrentUser(c *fiber.Ctx, user model.AuthUser) {
	c.Locals(authUserContextKey, user)
}

// CurrentUser mengambil identitas pengguna terautentikasi dari fiber.Ctx Locals.
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	val := c.Locals(authUserContextKey)
	if val == nil {
		return model.AuthUser{}, false
	}
	switch u := val.(type) {
	case model.AuthUser:
		return u, true
	case *model.AuthUser:
		if u != nil {
			return *u, true
		}
	}
	return model.AuthUser{}, false
}