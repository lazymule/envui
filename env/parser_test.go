package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	content := `# Database configuration
DATABASE_URL=postgres://localhost:5432/mydb
API_KEY="secret-key-123"
DEBUG=true

# Redis
REDIS_URL=redis://localhost:6379
EMPTY_VAR=
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	envFile, err := ParseEnvFile(envPath)
	if err != nil {
		t.Fatalf("ParseEnvFile failed: %v", err)
	}

	if envFile.Path != envPath {
		t.Errorf("Path = %q, want %q", envFile.Path, envPath)
	}

	if len(envFile.Variables) != 5 {
		t.Errorf("got %d variables, want 5", len(envFile.Variables))
	}

	// Check first variable
	if envFile.Variables[0].Key != "DATABASE_URL" {
		t.Errorf("first key = %q, want DATABASE_URL", envFile.Variables[0].Key)
	}
	if envFile.Variables[0].Value != "postgres://localhost:5432/mydb" {
		t.Errorf("first value = %q, want postgres://localhost:5432/mydb", envFile.Variables[0].Value)
	}

	// Check quoted variable (quotes should be stripped)
	if envFile.Variables[1].Key != "API_KEY" {
		t.Errorf("second key = %q, want API_KEY", envFile.Variables[1].Key)
	}
	if envFile.Variables[1].Value != "secret-key-123" {
		t.Errorf("second value = %q, want secret-key-123", envFile.Variables[1].Value)
	}
	if !envFile.Variables[1].IsSecret {
		t.Error("API_KEY should be marked as secret")
	}

	// Check comments are preserved
	if len(envFile.Comments) != 2 {
		t.Errorf("got %d comments, want 2", len(envFile.Comments))
	}
}

func TestParseEnvFile_NotFound(t *testing.T) {
	_, err := ParseEnvFile("/nonexistent/.env")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`hello`, "hello"},
		{`"hello`, `"hello`},
		{`hello"`, `hello"`},
		{`""`, ""},
		{`''`, ""},
		{`"it's fine"`, "it's fine"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := unquote(tt.input)
			if got != tt.expected {
				t.Errorf("unquote(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
