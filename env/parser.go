package env

import (
	"os"
	"strings"
)

// unquote removes surrounding quotes from a value
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// ParseEnvFile reads and parses a .env file
func ParseEnvFile(path string) (*EnvFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	envFile := &EnvFile{
		Path:     path,
		Comments: make(map[int]string),
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Preserve comments
		if strings.HasPrefix(trimmed, "#") {
			envFile.Comments[i] = line
			continue
		}

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Parse KEY=value
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := unquote(strings.TrimSpace(parts[1]))

		envFile.Variables = append(envFile.Variables, Variable{
			Key:      key,
			Value:    value,
			Source:   SourceFile,
			IsSecret: IsSecret(key),
			Line:     i,
		})
	}

	return envFile, nil
}
