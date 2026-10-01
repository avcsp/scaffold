package database

import (
	"io"
	"log/slog"
	"net/url"
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

func TestConnectFailsWithInvalidMariaDBDSN(t *testing.T) {
	os.Setenv("DATABASE_DSN", "mariadb://root:secret@invalid:99999/app?parseTime=True")
	defer os.Unsetenv("DATABASE_DSN")

	db, err := Connect(discardLogger())
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
	if db != nil {
		t.Fatal("expected nil db on connection failure")
	}
}

func TestMysqlDSNFromURL(t *testing.T) {
	u, err := url.Parse("mariadb://root:secret@localhost:3306/app?charset=utf8mb4&parseTime=True")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	got, err := mysqlDSNFromURL(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "root:secret@tcp(localhost:3306)/app?charset=utf8mb4&parseTime=True"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
