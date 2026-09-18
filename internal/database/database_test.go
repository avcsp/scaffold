package database

import (
	"io"
	"log/slog"
	"os"
	"testing"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestConnectSkipsWhenNoDSN(t *testing.T) {
	os.Unsetenv("DATABASE_DSN")

	db, err := Connect(discardLogger())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if db != nil {
		t.Fatal("expected nil db when DATABASE_DSN is not set")
	}
}

func TestConnectFailsWithInvalidDSN(t *testing.T) {
	os.Setenv("DATABASE_DSN", "host=invalid port=99999 user=nobody dbname=nope")
	defer os.Unsetenv("DATABASE_DSN")

	db, err := Connect(discardLogger())
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
	if db != nil {
		t.Fatal("expected nil db on connection failure")
	}
}

func TestCloseNilIsSafe(t *testing.T) {
	// Should not panic
	Close(nil)
}
