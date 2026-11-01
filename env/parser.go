package env

import (
	"os"
	"regexp"
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
	return ParseEnv(path, string(content)), nil
}

// ParseEnv parses the contents of a .env file. Lines that are not
// KEY=VALUE pairs (comments, blank or malformed lines) are kept verbatim so
// that saving never drops anything.
func ParseEnv(path, content string) *EnvFile {
	envFile := &EnvFile{
		Path:     path,
		Comments: make(map[int]string),
		CRLF:     strings.Contains(content, "\r\n"),
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			envFile.Comments[i] = line
			continue
		}
		v, ok := parseLine(line)
		if !ok {
			envFile.Comments[i] = line // keep malformed lines untouched
			continue
		}
		v.Line = i
		envFile.Variables = append(envFile.Variables, v)
	}
	attachDocs(envFile, len(lines))
	return envFile
}

// separatorRe matches decorative comment lines such as "# =========".
var separatorRe = regexp.MustCompile(`^\s*#\s*(?:={5,}|-{5,}|\*{5,}|#{5,}|~{5,}|_{5,})\s*$`)

// inlineTitleRe matches one-line section titles such as "# === Server ===".
var inlineTitleRe = regexp.MustCompile(`^\s*#\s*[=\-]{3,}\s*(\S.*?)\s*[=\-]{3,}\s*$`)

func commentText(line string) string {
	t := strings.TrimSpace(line)
	t = strings.TrimPrefix(t, "#")
	return strings.TrimPrefix(t, " ")
}

// attachDocs fills Doc, GroupDoc and Section of every variable from the
// comment lines around it.
func attachDocs(f *EnvFile, nLines int) {
	f.varLines = make(map[int]bool, len(f.Variables))
	f.docOwner = make(map[int]int)
	for _, v := range f.Variables {
		f.varLines[v.Line] = true
	}

	// sections: "# ====" / "# Title" (one or more lines) / "# ===="
	sectionAt := make(map[int]string) // line -> title starting there
	for i := 0; i < nLines; i++ {
		c, ok := f.Comments[i]
		if !ok {
			continue
		}
		if m := inlineTitleRe.FindStringSubmatch(c); m != nil && !separatorRe.MatchString(c) {
			sectionAt[i] = m[1]
			continue
		}
		if !separatorRe.MatchString(c) {
			continue
		}
		var title []string
		j := i + 1
		for ; j < nLines && j <= i+5; j++ {
			cj, ok := f.Comments[j]
			if !ok || separatorRe.MatchString(cj) {
				break
			}
			title = append(title, strings.TrimSpace(commentText(cj)))
		}
		if j < nLines && len(title) > 0 && separatorRe.MatchString(f.Comments[j]) {
			sectionAt[i] = strings.Join(title, " ")
			i = j // the closing separator is not a new section
		}
	}

	section := ""
	idx := 0
	byLine := make(map[int]int, len(f.Variables))
	for k, v := range f.Variables {
		byLine[v.Line] = k
	}
	for i := 0; i < nLines; i++ {
		if t, ok := sectionAt[i]; ok {
			section = t
		}
		k, ok := byLine[i]
		if !ok {
			continue
		}
		v := &f.Variables[k]
		v.Section = section

		// doc: comment lines directly above (no blank line, no separator)
		var doc []string
		var docLines []int
		for j := i - 1; j >= 0; j-- {
			c, ok := f.Comments[j]
			if !ok || separatorRe.MatchString(c) || inlineTitleRe.MatchString(c) {
				break
			}
			doc = append([]string{commentText(c)}, doc...)
			docLines = append([]int{j}, docLines...)
		}
		v.Doc, v.parsedDoc, v.docLines = strings.Join(doc, "\n"), strings.Join(doc, "\n"), docLines
		for _, j := range docLines {
			f.docOwner[j] = i
		}

		// group doc: the variable right above covers this one too
		if v.Doc == "" && idx > 0 {
			prev := f.Variables[k-1]
			if prev.Line == i-1 {
				if prev.Doc != "" {
					v.GroupDoc = prev.Doc
				} else {
					v.GroupDoc = prev.GroupDoc
				}
			}
		}
		idx++
	}
}

// parseLine parses a single KEY=VALUE line. It understands an optional
// "export " prefix, single and double quotes (with \" \\ \n escapes inside
// double quotes) and trailing "# comments" after the value.
func parseLine(line string) (Variable, bool) {
	rest := strings.TrimLeft(line, " \t")
	v := Variable{Source: SourceFile, raw: line}

	if strings.HasPrefix(rest, "export ") || strings.HasPrefix(rest, "export\t") {
		v.Export = true
		rest = strings.TrimLeft(rest[len("export"):], " \t")
	}

	eq := strings.IndexByte(rest, '=')
	if eq <= 0 {
		return v, false
	}
	v.Key = strings.TrimSpace(rest[:eq])
	if v.Key == "" || strings.ContainsAny(v.Key, " \t") {
		return v, false
	}
	rest = strings.TrimLeft(rest[eq+1:], " \t")

	var after string
	switch {
	case strings.HasPrefix(rest, `"`):
		value, end, ok := scanDoubleQuoted(rest)
		if !ok {
			v.Value, v.InlineComment, v.TrailingSpace = splitUnquoted(rest)
			break
		}
		v.Value, v.Quote, after = value, '"', rest[end:]
	case strings.HasPrefix(rest, "'"):
		end := strings.IndexByte(rest[1:], '\'')
		if end < 0 {
			v.Value, v.InlineComment, v.TrailingSpace = splitUnquoted(rest)
			break
		}
		v.Value, v.Quote, after = rest[1:1+end], '\'', rest[end+2:]
	default:
		v.Value, v.InlineComment, v.TrailingSpace = splitUnquoted(rest)
	}
	if v.Quote != 0 {
		if c := strings.TrimSpace(after); strings.HasPrefix(c, "#") {
			v.InlineComment = c
		}
	}

	v.IsSecret = IsSecret(v.Key)
	v.parsedKey, v.parsedValue, v.parsedExport = v.Key, v.Value, v.Export
	return v, true
}

// scanDoubleQuoted reads a double-quoted value starting at s[0] == '"'.
// It returns the unescaped value and the index just after the closing quote.
func scanDoubleQuoted(s string) (string, int, bool) {
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s):
			i++
			switch s[i] {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '"', '\\':
				b.WriteByte(s[i])
			default: // unknown escape: keep it as written
				b.WriteByte('\\')
				b.WriteByte(s[i])
			}
		case c == '"':
			return b.String(), i + 1, true
		default:
			b.WriteByte(c)
		}
	}
	return "", 0, false
}

// splitUnquoted separates an unquoted value from a trailing " # comment".
// A '#' only starts a comment when preceded by whitespace, so values such as
// URL fragments (http://x/#top) or colours (#fff... after '=') stay intact.
func splitUnquoted(s string) (value, comment string, trailingSpace bool) {
	for i := 1; i < len(s); i++ {
		if s[i] == '#' && (s[i-1] == ' ' || s[i-1] == '\t') {
			comment = strings.TrimSpace(s[i:])
			s = s[:i]
			break
		}
	}
	value = strings.TrimRight(s, " \t")
	trailingSpace = value != s && comment == ""
	return value, comment, trailingSpace
}
