package database

import (
	"context"
	"fmt"
	"time"

	"siplab/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool menginisialisasi pgxpool PostgreSQL sesuai parameter .env
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dbHost := config.GetEnv("DB_HOST", "localhost")
	dbPort := config.GetEnv("DB_PORT", "5432")
	dbUser := config.GetEnv("DB_USER", "postgres")
	dbPass := config.GetEnv("DB_PASSWORD", "postgres")
	dbName := config.GetEnv("DB_NAME", "siplab")
	sslMode := config.GetEnv("DB_SSLMODE", "disable")
	maxConns := int32(config.GetEnvInt("DB_MAX_CONNS", 10))

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPass, dbHost, dbPort, dbName, sslMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal parse dsn database: %w", err)
	}

	poolConfig.MaxConns = maxConns
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(pingCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool database: %w", err)
	}

	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("gagal ping database: %w", err)
	}

	return pool, nil
}