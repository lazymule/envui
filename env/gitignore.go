package env

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FindGitRoot returns the closest directory at or above dir that contains a
// .git entry (folder or file, for worktrees/submodules).
func FindGitRoot(dir string) (string, bool) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

type ignoreRule struct {
	base     string // directory the pattern is relative to (slash form, "" = root)
	pattern  string
	negate   bool
	anchored bool
}

// IsGitIgnored reports whether file (inside the repository at root) is
// ignored by the .gitignore files between root and the file's folder, or
// by .git/info/exclude. It implements the commonly used subset of the
// gitignore syntax (globs, **, leading /, !negation); directory-only rules
// ("dir/") are applied to the file's parent folders.
func IsGitIgnored(root, file string) bool {
	rel, err := filepath.Rel(root, file)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false
	}
	rel = filepath.ToSlash(rel)

	var rules []ignoreRule
	rules = append(rules, readIgnoreFile(filepath.Join(root, ".git", "info", "exclude"), "")...)
	dirs := strings.Split(path.Dir(rel), "/")
	base := ""
	rules = append(rules, readIgnoreFile(filepath.Join(root, ".gitignore"), "")...)
	for _, d := range dirs {
		if d == "." || d == "" {
			continue
		}
		base = path.Join(base, d)
		rules = append(rules, readIgnoreFile(filepath.Join(root, filepath.FromSlash(base), ".gitignore"), base)...)
	}

	// a file is ignored if one of its parent folders is ignored
	parts := strings.Split(rel, "/")
	for i := 1; i < len(parts); i++ {
		if matchRules(rules, strings.Join(parts[:i], "/"), true) {
			return true
		}
	}
	return matchRules(rules, rel, false)
}

func matchRules(rules []ignoreRule, rel string, isDir bool) bool {
	ignored := false
	for _, r := range rules {
		if r.matches(rel, isDir) {
			ignored = !r.negate
		}
	}
	return ignored
}

func (r ignoreRule) matches(rel string, isDir bool) bool {
	p := r.pattern
	dirOnly := strings.HasSuffix(p, "/")
	p = strings.TrimSuffix(p, "/")
	if dirOnly && !isDir {
		return false
	}
	if r.base != "" {
		if !strings.HasPrefix(rel, r.base+"/") {
			return false
		}
		rel = strings.TrimPrefix(rel, r.base+"/")
	}
	if !r.anchored {
		// matches the name at any depth
		return globMatch(p, path.Base(rel))
	}
	return globMatch(p, rel)
}

// globMatch is path.Match plus support for "**" path segments.
func globMatch(pattern, name string) bool {
	if !strings.Contains(pattern, "**") {
		ok, _ := path.Match(pattern, name)
		return ok
	}
	pp := strings.Split(pattern, "/")
	np := strings.Split(name, "/")
	var match func(i, j int) bool
	match = func(i, j int) bool {
		if i == len(pp) {
			return j == len(np)
		}
		if pp[i] == "**" {
			for k := j; k <= len(np); k++ {
				if match(i+1, k) {
					return true
				}
			}
			return false
		}
		if j == len(np) {
			return false
		}
		if ok, _ := path.Match(pp[i], np[j]); !ok {
			return false
		}
		return match(i+1, j+1)
	}
	return match(0, 0)
}

func readIgnoreFile(file, base string) []ignoreRule {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	var rules []ignoreRule
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), " \t\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		r := ignoreRule{base: base}
		if strings.HasPrefix(line, "!") {
			r.negate = true
			line = line[1:]
		}
		line = strings.TrimPrefix(line, `\`)
		trimmed := strings.TrimSuffix(line, "/")
		if strings.HasPrefix(trimmed, "**/") {
			trimmed = strings.TrimPrefix(trimmed, "**/")
			line = strings.TrimPrefix(line, "**/")
		}
		r.anchored = strings.Contains(trimmed, "/")
		r.pattern = strings.TrimPrefix(line, "/")
		rules = append(rules, r)
	}
	return rules
}

// UnignoredEnvFiles returns the project's env files (templates excluded)
// that live in a git repository but are not ignored by it. It returns nil
// when the project is not inside a git repository.
func UnignoredEnvFiles(p *Project) []*EnvFile {
	if p == nil {
		return nil
	}
	root, ok := FindGitRoot(p.Path)
	if !ok {
		return nil
	}
	var out []*EnvFile
	for _, f := range p.Files {
		if IsTemplate(f.Path) {
			continue
		}
		if !IsGitIgnored(root, f.Path) {
			out = append(out, f)
		}
	}
	return out
}

// AddToGitignore appends the given file names to the .gitignore in dir
// (creating it if needed), skipping names that are already listed.
func AddToGitignore(dir string, names []string) error {
	file := filepath.Join(dir, ".gitignore")
	existing, _ := os.ReadFile(file)
	have := make(map[string]bool)
	for _, l := range strings.Split(string(existing), "\n") {
		have[strings.TrimSpace(l)] = true
	}
	var b strings.Builder
	b.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		b.WriteString("\n")
	}
	added := false
	for _, n := range names {
		if have[n] || have["/"+n] {
			continue
		}
		if !added {
			b.WriteString("\n# Environment files (added by envui)\n")
			added = true
		}
		b.WriteString(n + "\n")
		have[n] = true
	}
	if !added {
		return nil
	}
	return os.WriteFile(file, []byte(b.String()), 0644)
}

// IsTemplate reports whether path is an example/template env file.
func IsTemplate(p string) bool {
	switch filepath.Base(p) {
	case ".env.example", ".env.sample", ".env.template":
		return true
	}
	return false
}
