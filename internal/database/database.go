package database

import (
	"fmt"
	"log/slog"

	"scaffold/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens a PostgreSQL connection if DATABASE_DSN is set in the environment.
// Returns nil, nil if DATABASE_DSN is not present (database is optional).
// Uses the scaffold logger for GORM query logging.
//
// DSN format: host=localhost port=5432 user=postgres password=pass dbname=app sslmode=disable
// Or: postgres://user:pass@host:port/dbname?sslmode=disable
func Connect(log *slog.Logger) (*gorm.DB, error) {
	dsn := config.Getenv("DATABASE_DSN", "")
	if dsn == "" {
		log.Info("DATABASE_DSN not set, skipping database connection")
		return nil, nil
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newSlogger(log),
	})
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	log.Info("database connected")
	return db, nil
}

// Close closes the underlying *sql.DB connection pool.
// Safe to call with nil.
func Close(db *gorm.DB) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	sqlDB.Close()
}
