package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"global-event-feed/config"
)

// ConnectDB connects to Postgress  using pgxpool and returns the pool
func ConnectDB(config *config.Config) *pgxpool.Pool{
	// Build connection string
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	// Create a context with timeout for inital connection
	ctx, cancle := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancle()

	// Connect to DB
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to create DB pool: %v", err)
	}

	// Ping DB to verify connection
	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping DB: %v", err)
	}

	log.Println("Successfully connected to Postgres")
	return pool
}
