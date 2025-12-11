package env

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// WriteEnvFile writes an EnvFile back to disk, preserving comments
func WriteEnvFile(envFile *EnvFile) error {
	// Find the maximum line number we need to account for
	maxLine := 0
	for _, v := range envFile.Variables {
		if v.Line > maxLine {
			maxLine = v.Line
		}
	}
	for line := range envFile.Comments {
		if line > maxLine {
			maxLine = line
		}
	}

	// Build a map of variables by line number
	varByLine := make(map[int]*Variable)
	for i := range envFile.Variables {
		varByLine[envFile.Variables[i].Line] = &envFile.Variables[i]
	}

	// Build the output preserving original structure
	var lines []string
	usedLines := make(map[int]bool)

	// First, write out lines in their original positions
	for i := 0; i <= maxLine; i++ {
		if comment, ok := envFile.Comments[i]; ok {
			lines = append(lines, comment)
			usedLines[i] = true
		} else if v, ok := varByLine[i]; ok {
			lines = append(lines, formatVariable(v))
			usedLines[i] = true
		} else {
			lines = append(lines, "")
		}
	}

	// Add any new variables (line == -1 or not in used lines)
	for i := range envFile.Variables {
		v := &envFile.Variables[i]
		if v.Line < 0 || !usedLines[v.Line] {
			lines = append(lines, formatVariable(v))
		}
	}

	// Remove trailing empty lines but keep at least one newline at end
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(envFile.Path, []byte(content), 0644)
}

// formatVariable formats a variable for writing to file
func formatVariable(v *Variable) string {
	// If value contains spaces or special chars, quote it
	if strings.ContainsAny(v.Value, " \t\"'#") {
		return fmt.Sprintf("%s=\"%s\"", v.Key, strings.ReplaceAll(v.Value, "\"", "\\\""))
	}
	return fmt.Sprintf("%s=%s", v.Key, v.Value)
}

// WriteEnvFileSimple writes variables to a file without preserving structure
// Useful for creating new files
func WriteEnvFileSimple(path string, variables []Variable) error {
	// Sort variables by key for consistent output
	sorted := make([]Variable, len(variables))
	copy(sorted, variables)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})

	var lines []string
	for _, v := range sorted {
		lines = append(lines, formatVariable(&v))
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}
