package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"global-event-feed/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB connects to Postgres using pgxpool and returns the pool.
func ConnectDB(cfg *config.Config) *pgxpool.Pool {
	// Build connection string.
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	// Create a context with timeout for initial connection.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to DB.
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("unable to create DB pool: %v", err)
	}

	// Ping DB to verify connection.
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("unable to ping DB: %v", err)
	}

	log.Println("Successfully connected to Postgres")
	return pool
}
