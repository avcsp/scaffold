package database

import (
	"fmt"
	"log/slog"

	"scaffold/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens a PostgreSQL connection using env vars and returns the *gorm.DB.
// Uses the scaffold logger for GORM query logging.
//
// Env vars: DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME, DB_SSLMODE
func Connect(log *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Getenv("DB_HOST", "localhost"),
		config.Getenv("DB_PORT", "5432"),
		config.Getenv("DB_USER", "postgres"),
		config.Getenv("DB_PASS", ""),
		config.Getenv("DB_NAME", "app"),
		config.Getenv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newSlogger(log),
	})
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	return db, nil
}

// Close closes the underlying *sql.DB connection pool.
func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	sqlDB.Close()
}
