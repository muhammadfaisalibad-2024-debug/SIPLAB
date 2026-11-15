package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"siplab/app/service"
	"siplab/helper"
	"siplab/middleware"
)

// Dependencies mengumpulkan seluruh komponen yang dibutuhkan untuk mendaftarkan route.
type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.JWTManager
	Permissions *helper.PermissionSet
	AuthService *service.AuthService
	UserService *service.UserService
	// --- SLOT ANGGOTA 2 (Ruang & Alat) ---
	// RoomService *service.RoomService
	// ToolService *service.ToolService

	// --- SLOT ANGGOTA 3 (Booking) ---
	BookingService *service.BookingService
}

// Register memetakan seluruh URL endpoint ke method service beserta middleware penjaganya.
func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")
	perms := deps.Permissions

	// --- 1. PUBLIK ---
	api.Get("/health", healthCheck(deps.Pool))

	// --- 2. AUTENTIKASI (Anggota 1) ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// --- 3. MANAJEMEN USER & RBAC (Anggota 1) ---
	users := api.Group("/users",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)
	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		deps.UserService.List,
	)
	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		deps.UserService.AssignRole,
	)
	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		deps.UserService.Delete,
	)
	users.Get("/:id", deps.UserService.Get)
	users.Patch("/:id", deps.UserService.Patch)

	// =========================================================================
	// --- 4. SLOT ANGGOTA 2: RUANG LAB (ROOMS) & ALAT LAB (TOOLS) ---
	// =========================================================================
	/*
		rooms := api.Group("/rooms",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT),
		)
		rooms.Get("/", deps.RoomService.List)
		rooms.Get("/:id", deps.RoomService.Get)
		rooms.Post("/",
			middleware.RequirePermission(perms, "room:create"),
			deps.RoomService.Create,
		)
		rooms.Put("/:id",
			middleware.RequirePermission(perms, "room:update"),
			deps.RoomService.Replace,
		)
		rooms.Delete("/:id",
			middleware.RequirePermission(perms, "room:delete"),
			deps.RoomService.Delete,
		)

		tools := api.Group("/tools",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT),
		)
		tools.Get("/", deps.ToolService.List)
		tools.Get("/:id", deps.ToolService.Get)
		tools.Post("/",
			middleware.RequirePermission(perms, "tool:create"),
			deps.ToolService.Create,
		)
		tools.Put("/:id",
			middleware.RequirePermission(perms, "tool:update"),
			deps.ToolService.Replace,
		)
		tools.Delete("/:id",
			middleware.RequirePermission(perms, "tool:delete"),
			deps.ToolService.Delete,
		)
	*/

	// =========================================================================
	// --- 5. SLOT ANGGOTA 3: PEMINJAMAN (BOOKINGS) ---
	// =========================================================================
	if deps.BookingService != nil {
		bookings := api.Group("/bookings",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT),
		)
		bookings.Post("/",
			middleware.RequirePermission(perms, "booking:create"),
			deps.BookingService.Create,
		)
		bookings.Patch("/:id/verify",
			middleware.RequirePermission(perms, "booking:verify"),
			deps.BookingService.Verify,
		)
		bookings.Get("/", deps.BookingService.List)
		bookings.Get("/:id", deps.BookingService.Get)
	}
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database SIPLAB berjalan normal", nil)
	}
}