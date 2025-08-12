package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Database struct {
	Pool *pgxpool.Pool
}

func Connect(config Config) (*Database, error) {
	dsn := BuildDSN(config)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse pgx config: %w", err)
	}

	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = time.Duration(config.MaxConnLifetime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(config.MaxConnIdleTime) * time.Second
	poolConfig.HealthCheckPeriod = time.Duration(config.HealthCheckPeriod) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("PostgreSQL connection established successfully (pgx v5)")

	return &Database{Pool: pool}, nil
}

func (db *Database) Close() {
	db.Pool.Close()
}

func BuildDSN(config Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmod=%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.SSLMode,
	)
}
