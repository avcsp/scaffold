package database

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"scaffold/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var instance *gorm.DB

// Instance returns the database connection established by Connect.
// Returns nil if DATABASE_DSN was not configured.
func Instance() *gorm.DB {
	return instance
}

// Connect opens a database connection if DATABASE_DSN is set in the environment.
// Returns nil, nil if DATABASE_DSN is not present (database is optional).
// Uses the scaffold logger for GORM query logging.
//
// The driver is picked based on the DSN scheme:
//   - MariaDB/MySQL: mariadb://user:pass@host:port/dbname?charset=utf8mb4&parseTime=True
//   - Postgres: host=localhost port=5432 user=postgres password=pass dbname=app sslmode=disable
//     Or: postgres://user:pass@host:port/dbname?sslmode=disable
func Connect(log *slog.Logger) (*gorm.DB, error) {
	dsn := config.Getenv("DATABASE_DSN", "")
	if dsn == "" {
		log.Info("DATABASE_DSN not set, skipping database connection")
		return nil, nil
	}

	dialector, err := dialectorFor(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_DSN: %w", err)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: newSlogger(log),
	})
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	instance = db
	log.Info("database connected")
	return db, nil
}

// dialectorFor picks the GORM dialector based on the DSN scheme.
// A "mariadb://" or "mysql://" scheme selects the MySQL driver; everything
// else (key=value pairs or a postgres:// URL) is treated as Postgres.
func dialectorFor(dsn string) (gorm.Dialector, error) {
	if !strings.HasPrefix(dsn, "mariadb://") && !strings.HasPrefix(dsn, "mysql://") {
		return postgres.Open(dsn), nil
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	mysqlDSN, err := mysqlDSNFromURL(u)
	if err != nil {
		return nil, err
	}
	return mysql.Open(mysqlDSN), nil
}

// mysqlDSNFromURL converts a mariadb://user:pass@host:port/dbname?params URL
// into the go-sql-driver/mysql DSN format: user:pass@tcp(host:port)/dbname?params
func mysqlDSNFromURL(u *url.URL) (string, error) {
	var userinfo string
	if u.User != nil {
		userinfo = u.User.String() + "@"
	}

	host := u.Host
	if host == "" {
		return "", fmt.Errorf("missing host")
	}

	dbname := strings.TrimPrefix(u.Path, "/")

	dsn := fmt.Sprintf("%stcp(%s)/%s", userinfo, host, dbname)
	if u.RawQuery != "" {
		dsn += "?" + u.RawQuery
	}
	return dsn, nil
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
