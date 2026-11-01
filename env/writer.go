package env

import (
	"os"
	"sort"
	"strings"
)

// WriteEnvFile writes an EnvFile back to disk, preserving comments, blank
// lines, unparsed lines and the exact text of every unchanged variable.
func WriteEnvFile(envFile *EnvFile) error {
	return os.WriteFile(envFile.Path, []byte(FormatEnvFile(envFile)), 0644)
}

// FormatEnvFile renders an EnvFile as it would be written to disk.
func FormatEnvFile(envFile *EnvFile) string {
	// Find the maximum line number we need to account for
	maxLine := -1
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

	// Build a map of variables by line number (the first one wins if a
	// line number is somehow shared)
	varByLine := make(map[int]*Variable)
	for i := range envFile.Variables {
		v := &envFile.Variables[i]
		if v.Line < 0 {
			continue
		}
		if _, taken := varByLine[v.Line]; !taken {
			varByLine[v.Line] = v
		}
	}

	// comment lines that are the (old) description of a variable that was
	// deleted or whose description was edited are not written back
	skip := make(map[int]bool)
	for line, owner := range envFile.docOwner {
		if v, ok := varByLine[owner]; !ok || v.DocChanged() {
			skip[line] = true
		}
	}

	var lines []string
	written := make(map[*Variable]bool)
	for i := 0; i <= maxLine; i++ {
		if skip[i] {
			continue
		}
		if comment, ok := envFile.Comments[i]; ok {
			lines = append(lines, comment)
		} else if v, ok := varByLine[i]; ok {
			if v.DocChanged() {
				lines = append(lines, docLines(v.Doc)...)
			}
			lines = append(lines, formatVariable(v))
			written[v] = true
		} else if envFile.varLines[i] {
			continue // a deleted variable: drop its line instead of leaving a gap
		} else {
			lines = append(lines, "")
		}
	}

	// New variables (and any whose line was already used) go at the end
	for i := range envFile.Variables {
		v := &envFile.Variables[i]
		if !written[v] {
			lines = append(lines, docLines(v.Doc)...)
			lines = append(lines, formatVariable(v))
		}
	}

	// Remove trailing empty lines but keep at least one newline at end
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	nl := "\n"
	if envFile.CRLF {
		nl = "\r\n"
	}
	return strings.Join(lines, nl) + nl
}

// docLines renders a description as "# " comment lines.
func docLines(doc string) []string {
	doc = strings.TrimRight(doc, "\n ")
	if doc == "" {
		return nil
	}
	var out []string
	for _, l := range strings.Split(doc, "\n") {
		l = strings.TrimRight(l, " \t\r")
		if l == "" {
			out = append(out, "#")
		} else {
			out = append(out, "# "+l)
		}
	}
	return out
}

// formatVariable formats a variable for writing to file. Unchanged
// variables are written exactly as they were read.
func formatVariable(v *Variable) string {
	if v.Unchanged() {
		return v.raw
	}
	var b strings.Builder
	if v.Export {
		b.WriteString("export ")
	}
	b.WriteString(v.Key)
	b.WriteByte('=')
	b.WriteString(quoteValue(v.Value, v.Quote))
	if v.InlineComment != "" {
		b.WriteString(" ")
		b.WriteString(v.InlineComment)
	}
	return b.String()
}

// quoteValue returns value in a form that ParseEnv and common dotenv
// loaders (DelphiMVCFramework, Docker Compose, python-dotenv, godotenv)
// read back as exactly the same string:
//
//   - plain values, including Windows paths with backslashes, stay unquoted
//   - values with # $ " or \ (plus spaces) use single quotes, which every
//     loader takes literally (no $VAR interpolation, no escapes)
//   - values that only need quoting because of spaces use double quotes
//   - values with a ' or a line break use double quotes with \ escapes
func quoteValue(value string, preferred byte) string {
	edgeSpace := value != strings.TrimSpace(value)
	needs := edgeSpace || strings.ContainsAny(value, " \t#$\"'\n\r")
	literalOnly := strings.ContainsAny(value, "#$\"\\")
	canSingle := !strings.ContainsAny(value, "'\n\r")

	switch {
	case !needs && preferred == 0:
		return value
	case !needs && preferred == '"' && !literalOnly:
		return `"` + value + `"`
	case canSingle && (literalOnly || preferred == '\''):
		return "'" + value + "'"
	case !strings.ContainsAny(value, "\n\r\"\\") && !strings.Contains(value, "$"):
		return `"` + value + `"`
	}
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`)
	return `"` + r.Replace(value) + `"`
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
	for i := range sorted {
		sorted[i].raw = "" // always format
		lines = append(lines, formatVariable(&sorted[i]))
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}
