package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverOtherEnvFiles(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{".env", ".env.docker", ".env.dev", ".env.local", ".envrc", ".env.bak", ".env~", "env.txt"} {
		os.WriteFile(filepath.Join(dir, n), []byte("A=1\n"), 0o644)
	}
	os.Mkdir(filepath.Join(dir, ".env.d"), 0o755) // a folder, not a file

	var got []string
	for _, p := range DiscoverEnvFiles(dir) {
		got = append(got, filepath.Base(p))
	}
	want := []string{".env", ".env.local", ".env.dev", ".env.docker"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
