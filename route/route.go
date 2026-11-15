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
	RoomService *service.RoomService

	// --- SLOT ANGGOTA 3 (Booking) ---
	BookingService *service.BookingService
}

// Register memetakan seluruh URL endpoint ke method service beserta middleware penjaganya.
// Prefix resmi adalah /api/v1, sedangkan /api disediakan sebagai alias agar dokumentasi
// dan klien yang lebih lama tetap dapat memanggil endpoint yang sama.
func Register(app *fiber.App, deps Dependencies) {
	for _, prefix := range []string{"/api/v1", "/api"} {
		registerPrefix(app.Group(prefix), deps)
	}
}

func registerPrefix(api fiber.Router, deps Dependencies) {
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
	// --- 4. RUANG LAB (ROOMS) & ALAT LAB (TOOLS) — Anggota 2 ---
	// =========================================================================
	rooms := api.Group("/rooms",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)
	// Route statis /availability wajib didaftarkan sebelum /:id agar tidak tertangkap sebagai id.
	rooms.Get("/availability", deps.RoomService.Availability)
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
	tools.Get("/", deps.RoomService.ListTools)
	tools.Post("/",
		middleware.RequirePermission(perms, "tool:create"),
		deps.RoomService.CreateTool,
	)

	// =========================================================================
	// --- 5. PEMINJAMAN (BOOKINGS) — Anggota 3 ---
	// =========================================================================
	bookings := api.Group("/bookings",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)
	bookings.Get("/", deps.BookingService.List)
	bookings.Post("/",
		middleware.RequirePermission(perms, "booking:create"),
		deps.BookingService.Create,
	)
	bookings.Get("/:id", deps.BookingService.Get)

	// Verifikasi Laboran/Admin: /verify adalah nama resmi, /status disediakan sebagai alias.
	bookings.Patch("/:id/verify",
		middleware.RequirePermission(perms, "booking:verify"),
		deps.BookingService.Verify,
	)
	bookings.Patch("/:id/status",
		middleware.RequirePermission(perms, "booking:verify"),
		deps.BookingService.Verify,
	)

	// Pembatalan oleh peminjam (ownership diperiksa di service).
	bookings.Patch("/:id/cancel", deps.BookingService.Cancel)
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
