package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverEnvFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some env files
	files := []string{".env", ".env.local", ".env.example"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("TEST=1"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", f, err)
		}
	}

	// Create a non-env file that shouldn't be discovered
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create config.json: %v", err)
	}

	found := DiscoverEnvFiles(tmpDir)

	if len(found) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(found), found)
	}

	// Check all expected files are found
	foundMap := make(map[string]bool)
	for _, f := range found {
		foundMap[filepath.Base(f)] = true
	}

	for _, expected := range files {
		if !foundMap[expected] {
			t.Errorf("expected to find %s", expected)
		}
	}
}

func TestDiscoverEnvFiles_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	found := DiscoverEnvFiles(tmpDir)

	if len(found) != 0 {
		t.Errorf("expected 0 files in empty dir, got %d", len(found))
	}
}

func TestLoadProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .env file
	envContent := `DATABASE_URL=postgres://localhost
API_KEY=secret123`
	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}

	// Create .env.example file
	exampleContent := `DATABASE_URL=
API_KEY=
DEBUG=false`
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte(exampleContent), 0644); err != nil {
		t.Fatalf("failed to create .env.example: %v", err)
	}

	project, err := LoadProject(tmpDir)
	if err != nil {
		t.Fatalf("LoadProject failed: %v", err)
	}

	if project.Path != tmpDir {
		t.Errorf("Path = %q, want %q", project.Path, tmpDir)
	}

	if len(project.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(project.Files))
	}

	if project.ActiveFile == nil {
		t.Error("ActiveFile should not be nil")
	}

	// Shell vars should be populated
	if len(project.ShellVars) == 0 {
		t.Error("ShellVars should not be empty")
	}
}
