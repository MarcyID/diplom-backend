package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config конфигурация подключения к БД
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// PostgreSQL хранилище
type PostgreSQL struct {
	Pool *pgxpool.Pool
}

// NewPostgreSQL создает новое подключение к PostgreSQL
func NewPostgreSQL(cfg Config) (*PostgreSQL, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgreSQL{Pool: pool}, nil
}

// Close закрывает подключение к БД
func (p *PostgreSQL) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
