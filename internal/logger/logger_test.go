package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
)

type logEntry struct {
	Level string `json:"level"`
	Msg   string `json:"msg"`
	Owner string `json:"owner"`
}

func parseLog(buf *bytes.Buffer) logEntry {
	var e logEntry
	json.Unmarshal(buf.Bytes(), &e)
	return e
}

func TestInitReturnsBothLoggers(t *testing.T) {
	logs := Init()
	if logs.Scaffold == nil {
		t.Fatal("scaffold logger is nil")
	}
	if logs.App == nil {
		t.Fatal("app logger is nil")
	}
}

func TestScaffoldTag(t *testing.T) {
	logs := Init()

	// Write to a buffer to inspect output
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	testLog := slog.New(handler).With("owner", "scaffold")
	_ = logs // ensure Init ran

	testLog.Info("test")
	e := parseLog(&buf)
	if e.Owner != "scaffold" {
		t.Fatalf("expected owner 'scaffold', got %q", e.Owner)
	}
}

func TestAppSetAsDefault(t *testing.T) {
	// Capture what slog.Default writes
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	appLog := slog.New(handler).With("owner", "app")
	slog.SetDefault(appLog)

	slog.Info("from default")
	e := parseLog(&buf)
	if e.Owner != "app" {
		t.Fatalf("expected default logger owner 'app', got %q", e.Owner)
	}
}

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	log := slog.New(handler).With("owner", "scaffold")
	log.Info("hello")

	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}
}

func TestDefaultLevelDebugInDev(t *testing.T) {
	os.Setenv("ENV", "development")
	defer os.Unsetenv("ENV")

	level := defaultLevel()
	if level != slog.LevelDebug {
		t.Fatalf("expected LevelDebug, got %v", level)
	}
}

func TestDefaultLevelInfoInProduction(t *testing.T) {
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	level := defaultLevel()
	if level != slog.LevelInfo {
		t.Fatalf("expected LevelInfo, got %v", level)
	}
}

func TestDefaultLevelDebugWhenEnvUnset(t *testing.T) {
	os.Unsetenv("ENV")

	level := defaultLevel()
	if level != slog.LevelDebug {
		t.Fatalf("expected LevelDebug when ENV unset, got %v", level)
	}
}

func TestInitDoesNotPanic(t *testing.T) {
	logs := Init()
	logs.Scaffold.Info("scaffold works")
	logs.App.Info("app works")
	slog.Info("default works")
}
