package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
)

// PortableFileName is the settings file of the portable edition. If it sits
// next to the executable, envui keeps all its settings in it (instead of the
// user profile), so the folder can be carried around on a USB stick.
const PortableFileName = "envui-portable.json"

// Prefs is the subset of fyne.Preferences that envui uses.
type Prefs interface {
	Bool(key string) bool
	BoolWithFallback(key string, fallback bool) bool
	SetBool(key string, value bool)
	String(key string) string
	StringWithFallback(key, fallback string) string
	SetString(key string, value string)
}

var _ Prefs = (fyne.Preferences)(nil)

// portablePath returns the portable settings file next to the executable,
// or "" when envui is not running as the portable edition.
func portablePath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	dir := filepath.Dir(exe)
	candidates := []string{filepath.Join(dir, PortableFileName)}
	// the software-OpenGL copy in "softgl" shares the main folder's settings
	if strings.EqualFold(filepath.Base(dir), "softgl") {
		candidates = append(candidates, filepath.Join(filepath.Dir(dir), PortableFileName))
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// filePrefs stores preferences as JSON in a single file.
type filePrefs struct {
	mu     sync.Mutex
	path   string
	values map[string]any
}

func newFilePrefs(path string) *filePrefs {
	p := &filePrefs{path: path, values: map[string]any{}}
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &p.values)
	}
	return p
}

func (p *filePrefs) get(key string) (any, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	v, ok := p.values[key]
	return v, ok
}

func (p *filePrefs) set(key string, value any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.values[key] = value
	data, err := json.MarshalIndent(p.values, "", "  ")
	if err == nil {
		// write to a temp file first so a crash can't leave a broken file
		tmp := p.path + ".tmp"
		if os.WriteFile(tmp, data, 0o644) == nil {
			_ = os.Rename(tmp, p.path)
		}
	}
}

func (p *filePrefs) Bool(key string) bool { return p.BoolWithFallback(key, false) }

func (p *filePrefs) BoolWithFallback(key string, fallback bool) bool {
	if v, ok := p.get(key); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return fallback
}

func (p *filePrefs) SetBool(key string, value bool) { p.set(key, value) }

func (p *filePrefs) String(key string) string { return p.StringWithFallback(key, "") }

func (p *filePrefs) StringWithFallback(key, fallback string) string {
	if v, ok := p.get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return fallback
}

func (p *filePrefs) SetString(key string, value string) { p.set(key, value) }
