package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envFile := &EnvFile{
		Path: envPath,
		Variables: []Variable{
			{Key: "DATABASE_URL", Value: "postgres://localhost", Line: 1},
			{Key: "API_KEY", Value: "secret123", Line: 2},
			{Key: "DEBUG", Value: "true", Line: 3},
		},
		Comments: map[int]string{
			0: "# Configuration",
		},
	}

	if err := WriteEnvFile(envFile); err != nil {
		t.Fatalf("WriteEnvFile failed: %v", err)
	}

	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	// Check comment is preserved
	if !strings.Contains(string(content), "# Configuration") {
		t.Error("comment not preserved in output")
	}

	// Check variables are written
	if !strings.Contains(string(content), "DATABASE_URL=postgres://localhost") {
		t.Error("DATABASE_URL not found in output")
	}
	if !strings.Contains(string(content), "API_KEY=secret123") {
		t.Error("API_KEY not found in output")
	}
	if !strings.Contains(string(content), "DEBUG=true") {
		t.Error("DEBUG not found in output")
	}
}

func TestWriteEnvFile_QuotesSpecialChars(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envFile := &EnvFile{
		Path: envPath,
		Variables: []Variable{
			{Key: "MESSAGE", Value: "hello world", Line: 0}, // has space
			{Key: "SIMPLE", Value: "nospaceshere", Line: 1},
		},
		Comments: map[int]string{},
	}

	if err := WriteEnvFile(envFile); err != nil {
		t.Fatalf("WriteEnvFile failed: %v", err)
	}

	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	// Value with space should be quoted
	if !strings.Contains(string(content), `MESSAGE="hello world"`) {
		t.Errorf("expected MESSAGE to be quoted, got: %s", string(content))
	}

	// Simple value should not be quoted
	if !strings.Contains(string(content), "SIMPLE=nospaceshere") {
		t.Errorf("expected SIMPLE without quotes, got: %s", string(content))
	}
}

func TestWriteEnvFileSimple(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.new")

	variables := []Variable{
		{Key: "ZEBRA", Value: "last"},
		{Key: "APPLE", Value: "first"},
		{Key: "MANGO", Value: "middle"},
	}

	if err := WriteEnvFileSimple(envPath, variables); err != nil {
		t.Fatalf("WriteEnvFileSimple failed: %v", err)
	}

	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Should be sorted alphabetically
	if !strings.HasPrefix(lines[0], "APPLE=") {
		t.Errorf("expected APPLE first, got: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "MANGO=") {
		t.Errorf("expected MANGO second, got: %s", lines[1])
	}
	if !strings.HasPrefix(lines[2], "ZEBRA=") {
		t.Errorf("expected ZEBRA third, got: %s", lines[2])
	}
}
