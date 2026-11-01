package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	f := ParseEnv(".env", "A=1\nB=x   \nA=2\nC=\nD=\" pad\"\n")
	tpl := ParseEnv(".env.example", "C=required\nD=\n")
	got := map[IssueKind][]string{}
	for _, is := range Validate(f, tpl) {
		got[is.Kind] = append(got[is.Kind], is.Key)
	}
	if len(got[IssueDuplicate]) != 1 || got[IssueDuplicate][0] != "A" {
		t.Errorf("duplicates: %v", got[IssueDuplicate])
	}
	if len(got[IssueSpaces]) != 2 { // B (trailing in file) and D (quoted leading space)
		t.Errorf("spaces: %v", got[IssueSpaces])
	}
	if len(got[IssueEmptyRequired]) != 1 || got[IssueEmptyRequired][0] != "C" {
		t.Errorf("empty required: %v", got[IssueEmptyRequired])
	}
}

func TestCompare(t *testing.T) {
	l := ParseEnv("a", "A=1\nB=2\nC=3\n")
	r := ParseEnv("b", "B=2\nC=x\nD=4\n")
	rows := Compare(l, r)
	want := []struct {
		key  string
		kind DiffKind
	}{{"A", DiffOnlyLeft}, {"B", DiffSame}, {"C", DiffChanged}, {"D", DiffOnlyRight}}
	if len(rows) != len(want) {
		t.Fatalf("rows: %+v", rows)
	}
	for i, w := range want {
		if rows[i].Key != w.key || rows[i].Kind != w.kind {
			t.Errorf("row %d = %+v, want %v %v", i, rows[i], w.key, w.kind)
		}
	}
}

func TestGitIgnore(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".git", "info"), 0o755)
	os.MkdirAll(filepath.Join(root, "app", "config"), 0o755)
	write := func(p, s string) { os.WriteFile(filepath.Join(root, p), []byte(s), 0o644) }

	write(".gitignore", ".env\n*.local\n!keep.local\n/build/\n")
	write("app/.gitignore", "/config/.env.staging\n")
	write(".git/info/exclude", ".env.production\n")

	cases := map[string]bool{
		".env":                    true,
		"app/.env":                true, // unanchored: any depth
		"app/.env.local":          true,
		"app/keep.local":          false, // negated
		"app/.env.staging":        false,
		"app/config/.env.staging": true, // anchored in app/.gitignore
		".env.production":         true, // info/exclude
		"build/.env.test":         true, // ignored parent folder
		".env.test":               false,
	}
	for rel, want := range cases {
		if got := IsGitIgnored(root, filepath.Join(root, filepath.FromSlash(rel))); got != want {
			t.Errorf("IsGitIgnored(%s) = %v, want %v", rel, got, want)
		}
	}

	if r, ok := FindGitRoot(filepath.Join(root, "app", "config")); !ok || r != root {
		t.Errorf("FindGitRoot = %q %v", r, ok)
	}

	// AddToGitignore appends once and doesn't duplicate
	dir := filepath.Join(root, "app")
	if err := AddToGitignore(dir, []string{".env.staging", ".env.staging"}); err != nil {
		t.Fatal(err)
	}
	AddToGitignore(dir, []string{".env.staging"})
	if !IsGitIgnored(root, filepath.Join(dir, ".env.staging")) {
		t.Error("file should be ignored after AddToGitignore")
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if n := countLines(string(data), ".env.staging"); n != 1 {
		t.Errorf(".env.staging listed %d times:\n%s", n, data)
	}
}

func countLines(s, line string) int {
	n := 0
	for _, l := range splitLines(s) {
		if l == line {
			n++
		}
	}
	return n
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}
