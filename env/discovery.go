package env

import (
	"os"
	"path/filepath"
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

// DiscoverEnvFiles finds all .env files in a directory
func DiscoverEnvFiles(projectPath string) []string {
	var found []string

	for _, pattern := range envFilePatterns {
		path := filepath.Join(projectPath, pattern)
		if _, err := os.Stat(path); err == nil {
			found = append(found, path)
		}
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
