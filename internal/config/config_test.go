package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	os.WriteFile(path, []byte(content), 0644)
	return path
}

func TestBasicKeyValue(t *testing.T) {
	path := writeTempEnv(t, "FOO=bar\n")
	os.Unsetenv("FOO")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("FOO"); v != "bar" {
		t.Fatalf("expected 'bar', got %q", v)
	}
}

func TestMultipleVars(t *testing.T) {
	path := writeTempEnv(t, "A=1\nB=2\nC=3\n")
	os.Unsetenv("A")
	os.Unsetenv("B")
	os.Unsetenv("C")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct{ key, want string }{
		{"A", "1"},
		{"B", "2"},
		{"C", "3"},
	} {
		if v := os.Getenv(tc.key); v != tc.want {
			t.Fatalf("%s: expected %q, got %q", tc.key, tc.want, v)
		}
	}
}

func TestSkipsComments(t *testing.T) {
	path := writeTempEnv(t, "# this is a comment\nKEY=value\n")
	os.Unsetenv("KEY")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("KEY"); v != "value" {
		t.Fatalf("expected 'value', got %q", v)
	}
}

func TestSkipsEmptyLines(t *testing.T) {
	path := writeTempEnv(t, "\n\nKEY=value\n\n")
	os.Unsetenv("KEY")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("KEY"); v != "value" {
		t.Fatalf("expected 'value', got %q", v)
	}
}

func TestDoubleQuotedValue(t *testing.T) {
	path := writeTempEnv(t, `KEY="hello world"`)
	os.Unsetenv("KEY")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("KEY"); v != "hello world" {
		t.Fatalf("expected 'hello world', got %q", v)
	}
}

func TestSingleQuotedValue(t *testing.T) {
	path := writeTempEnv(t, `KEY='hello world'`)
	os.Unsetenv("KEY")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("KEY"); v != "hello world" {
		t.Fatalf("expected 'hello world', got %q", v)
	}
}

func TestDoesNotOverrideExisting(t *testing.T) {
	os.Setenv("EXISTING", "original")
	defer os.Unsetenv("EXISTING")

	path := writeTempEnv(t, "EXISTING=overwritten\n")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("EXISTING"); v != "original" {
		t.Fatalf("expected 'original', got %q", v)
	}
}

func TestMissingFileReturnsNil(t *testing.T) {
	err := Load("/nonexistent/path/.env")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
}

func TestSkipsLinesWithoutEquals(t *testing.T) {
	path := writeTempEnv(t, "INVALID_LINE\nKEY=value\n")
	os.Unsetenv("KEY")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("KEY"); v != "value" {
		t.Fatalf("expected 'value', got %q", v)
	}
}

func TestTrimsWhitespace(t *testing.T) {
	path := writeTempEnv(t, "  KEY  =  value  \n")
	os.Unsetenv("KEY")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("KEY"); v != "value" {
		t.Fatalf("expected 'value', got %q", v)
	}
}

func TestEmptyValue(t *testing.T) {
	path := writeTempEnv(t, "KEY=\n")
	os.Unsetenv("KEY")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("KEY"); v != "" {
		t.Fatalf("expected empty string, got %q", v)
	}
}

func TestGetenvReturnsValue(t *testing.T) {
	os.Setenv("TEST_GETENV", "actual")
	defer os.Unsetenv("TEST_GETENV")

	if v := Getenv("TEST_GETENV", "default"); v != "actual" {
		t.Fatalf("expected 'actual', got %q", v)
	}
}

func TestGetenvReturnsFallback(t *testing.T) {
	os.Unsetenv("TEST_MISSING")

	if v := Getenv("TEST_MISSING", "fallback"); v != "fallback" {
		t.Fatalf("expected 'fallback', got %q", v)
	}
}

func TestGetenvEmptyValueIsNotMissing(t *testing.T) {
	os.Setenv("TEST_EMPTY", "")
	defer os.Unsetenv("TEST_EMPTY")

	if v := Getenv("TEST_EMPTY", "fallback"); v != "" {
		t.Fatalf("expected empty string, got %q", v)
	}
}

func TestValueWithEquals(t *testing.T) {
	path := writeTempEnv(t, "KEY=foo=bar=baz\n")
	os.Unsetenv("KEY")

	if err := Load(path); err != nil {
		t.Fatal(err)
	}

	if v := os.Getenv("KEY"); v != "foo=bar=baz" {
		t.Fatalf("expected 'foo=bar=baz', got %q", v)
	}
}
