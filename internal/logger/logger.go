package logger

import (
	"log/slog"
	"os"
)

// Loggers holds the scaffold and app logger instances.
type Loggers struct {
	Scaffold *slog.Logger
	App      *slog.Logger
}

// Init creates both scaffold and app loggers with JSON output to stdout.
// Level is based on ENV: "production" → Info, otherwise Debug.
// The app logger is set as the slog default so handlers can use slog.Info() directly.
func Init() Loggers {
	level := defaultLevel()
	opts := &slog.HandlerOptions{Level: level}

	l := Loggers{
		Scaffold: slog.New(slog.NewJSONHandler(os.Stdout, opts)).With("owner", "scaffold"),
		App:      slog.New(slog.NewJSONHandler(os.Stdout, opts)).With("owner", "app"),
	}

	slog.SetDefault(l.App)

	return l
}

func defaultLevel() slog.Level {
	switch os.Getenv("ENV") {
	case "production":
		return slog.LevelInfo
	default:
		return slog.LevelDebug
	}
}
