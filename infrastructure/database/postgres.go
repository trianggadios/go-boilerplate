package database

import (
	"boilerplate-go/config"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresDB wraps the database connection.
type PostgresDB struct {
	DB *sql.DB
}

// NewPostgresConnection creates a new PostgreSQL database connection with configuration.
func NewPostgresConnection(cfg config.DatabaseConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{DB: db}, nil
}

// Close closes the database connection.
func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

// Ping checks if the database connection is alive.
func (p *PostgresDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.DB.PingContext(ctx)
}
