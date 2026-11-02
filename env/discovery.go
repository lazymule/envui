package env

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var envFilePatterns = []string{
	".env",
	".env.local",
	".env.development",
	".env.staging",
	".env.production",
	".env.prod",
	".env.example",
	".env.sample",
	".env.template",
	".env.test",
}

// IsEnvFileName reports whether a file name looks like an env file: ".env"
// or ".env.<something>", excluding editor backups and swap files.
func IsEnvFileName(name string) bool {
	if name == ".env" {
		return true
	}
	if !strings.HasPrefix(name, ".env.") || len(name) == len(".env.") {
		return false
	}
	lower := strings.ToLower(name)
	for _, s := range []string{"~", ".bak", ".swp", ".swo", ".tmp", ".orig", ".old"} {
		if strings.HasSuffix(lower, s) {
			return false
		}
	}
	return true
}

// DiscoverEnvFiles finds all .env files in a directory: the well-known
// names first (in a fixed order), then any other ".env.*" file A-Z.
func DiscoverEnvFiles(projectPath string) []string {
	var found []string
	seen := make(map[string]bool)

	for _, pattern := range envFilePatterns {
		path := filepath.Join(projectPath, pattern)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			found = append(found, path)
			seen[pattern] = true
		}
	}

	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return found
	}
	var extra []string
	for _, e := range entries {
		if e.IsDir() || seen[e.Name()] || !IsEnvFileName(e.Name()) {
			continue
		}
		extra = append(extra, e.Name())
	}
	sort.Strings(extra)
	for _, name := range extra {
		found = append(found, filepath.Join(projectPath, name))
	}
	return found
}

// LoadProject loads a project directory and all its env files
func LoadProject(projectPath string) (*Project, error) {
	project := &Project{
		Path:      projectPath,
		ShellVars: GetShellVariables(),
	}

	filePaths := DiscoverEnvFiles(projectPath)
	for _, path := range filePaths {
		envFile, err := ParseEnvFile(path)
		if err != nil {
			continue // Skip files that can't be parsed
		}
		project.Files = append(project.Files, envFile)
	}

	// Set the first file as active if any exist
	if len(project.Files) > 0 {
		project.ActiveFile = project.Files[0]
	}

	return project, nil
}
