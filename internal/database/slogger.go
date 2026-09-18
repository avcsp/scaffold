package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// slogger adapts *slog.Logger to GORM's logger.Interface.
type slogger struct {
	log   *slog.Logger
	level gormlogger.LogLevel
}

func newSlogger(log *slog.Logger) *slogger {
	return &slogger{
		log:   log,
		level: gormlogger.Info,
	}
}

func (s *slogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &slogger{log: s.log, level: level}
}

func (s *slogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if s.level >= gormlogger.Info {
		s.log.Info(fmt.Sprintf(msg, args...))
	}
}

func (s *slogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if s.level >= gormlogger.Warn {
		s.log.Warn(fmt.Sprintf(msg, args...))
	}
}

func (s *slogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if s.level >= gormlogger.Error {
		s.log.Error(fmt.Sprintf(msg, args...))
	}
}

func (s *slogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if s.level <= gormlogger.Silent {
		return
	}

	duration := time.Since(begin)
	sql, rows := fc()

	attrs := []any{
		"duration", duration.String(),
		"rows", rows,
		"sql", sql,
	}

	switch {
	case err != nil:
		s.log.Error("query error", append(attrs, "error", err.Error())...)
	case duration > 200*time.Millisecond:
		s.log.Warn("slow query", attrs...)
	default:
		s.log.Debug("query", attrs...)
	}
}
