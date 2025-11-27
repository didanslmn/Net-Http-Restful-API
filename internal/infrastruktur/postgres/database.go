package database

import (
	"context"
	"fmt"
	"postgresDB/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnection(dbCfg config.DBConfig) (*pgxpool.Pool, error) {
	dsn := dbCfg.ConnectionUrl
	if dsn == "" {
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			dbCfg.User,
			dbCfg.Password,
			dbCfg.Host,
			dbCfg.Port,
			dbCfg.DBName,
			dbCfg.SSLMode,
		)
	}

	// Parse the connection string
	pgxCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %v", err)
	}
	// configurasi connection pool
	pgxCfg.MaxConns = 10
	pgxCfg.MinConns = 2
	pgxCfg.MaxConnLifetime = time.Hour
	pgxCfg.MaxConnIdleTime = 30 * time.Minute

	// Create a connection pool with configured settings
	pool, err := pgxpool.NewWithConfig(context.Background(), pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("error to create connection pool: %v", err)
	}

	// Test the connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping error to connect database: %v", err)
	}

	fmt.Println("Connection Success")
	return pool, nil

}
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
