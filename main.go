package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"siplab/app/repository"
	"siplab/app/service"
	"siplab/config"
	"siplab/database"
	"siplab/helper"
	"siplab/route"
)

// main hanya berisi urutan perakitan (Dependency Injection manual) dan graceful shutdown.
// Urutan baku: config -> logger -> pool -> repository -> permission set -> service -> app -> graceful shutdown.
func main() {
	// 1. Muat konfigurasi environment (.env)
	config.LoadEnv()

	// 2. Inisialisasi structured logger (stdout + rotating file logs/app.log)
	logger := config.NewLogger()

	// 3. Buat connection pool PostgreSQL (pgxpool)
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 4. Rakit layer Repository
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)
	roomRepository := repository.NewRoomRepository(pool)
	bookingRepository := repository.NewBookingRepository(pool)

	// 5. Muat pemetaan RBAC PermissionSet sekali saat startup (fail fast bila gagal)
	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission RBAC dari database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission RBAC berhasil dimuat", slog.Any("roles", permissions.KnownRoles()))

	// 6. Inisialisasi JWTManager (Access Token 15 menit, Refresh Token 7 hari) dan layer Service
	jwtSecret := config.GetEnv("JWT_SECRET", "siplab-rahasia-kampus-2026-minimal-32-karakter")
	accessTTL := time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute
	refreshTTL := time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour
	jwtManager := helper.NewJWTManager(jwtSecret, accessTTL, refreshTTL)

	authService := service.NewAuthService(userRepository, tokenRepository, jwtManager, permissions)
	userService := service.NewUserService(userRepository, permissions)
	roomService := service.NewRoomService(roomRepository)
	bookingService := service.NewBookingService(bookingRepository, permissions)

	// 7. Rakit instance Fiber App melalui config.NewApp
	app := config.NewApp(logger, route.Dependencies{
		Pool:           pool,
		JWT:            jwtManager,
		Permissions:    permissions,
		AuthService:    authService,
		UserService:    userService,
		RoomService:    roomService,
		BookingService: bookingService,
	})

	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti mendadak", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server SIPLAB berjalan", slog.String("port", port))

	// 8. Graceful Shutdown: tunggu sinyal SIGINT/SIGTERM dan beri waktu hingga 10 detik untuk menyelesaikan request aktif
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server SIPLAB secara rapi...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}
	logger.Info("server SIPLAB berhenti dengan rapi")
}
