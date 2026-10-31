package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestValidateKey(t *testing.T) {
	for _, k := range []string{"FOO", "_FOO", "foo_bar1", "A.B", "MY-KEY"} {
		if err := validateKey(k); err != nil {
			t.Errorf("validateKey(%q) = %v, want nil", k, err)
		}
	}
	for _, k := range []string{"", "1FOO", "FOO BAR", "FOO=BAR", "ŞEHİR"} {
		if err := validateKey(k); err == nil {
			t.Errorf("validateKey(%q) = nil, want error", k)
		}
	}
}

func TestTruncateIsRuneSafe(t *testing.T) {
	got := truncate("çğışöüçğışöü", 8)
	if got != "çğışö..." {
		t.Errorf("truncate = %q", got)
	}
	if truncate("kısa", 10) != "kısa" {
		t.Error("short strings must be unchanged")
	}
}

func TestFitText(t *testing.T) {
	test.NewApp()
	style := fyne.TextStyle{Monospace: true}

	short := "Çağrı"
	if got := fitText(short, 1000, style); got != short {
		t.Errorf("fitText(short) = %q, want unchanged", got)
	}

	long := strings.Repeat("Şirket Uygulaması ", 20)
	width := float32(200)
	got := fitText(long, width, style)
	if !strings.HasSuffix(got, ellipsis) {
		t.Fatalf("expected ellipsis, got %q", got)
	}
	if w := fyne.MeasureText(got, 14, style).Width; w > width+20 {
		t.Errorf("fitted text too wide: %v > %v", w, width)
	}
}

func TestTranslations(t *testing.T) {
	defer SetLanguage(LangEnglish)

	SetLanguage(LangEnglish)
	if got := Tn("stats.vars", 1); got != "1 variable" {
		t.Errorf("en one: %q", got)
	}
	if got := Tn("stats.vars", 3); got != "3 variables" {
		t.Errorf("en other: %q", got)
	}

	SetLanguage(LangTurkish)
	if got := Tn("stats.vars", 1); got != "1 değişken" {
		t.Errorf("tr one: %q", got)
	}
	if got := T("status.missing", 2, ".env.example"); got != ".env.example dosyasına göre 2 eksik" {
		t.Errorf("tr missing: %q", got)
	}

	// every English key must exist in every other language
	for lang, m := range messages {
		for key := range messages[LangEnglish] {
			if strings.HasSuffix(key, ".one") {
				continue // languages without singular forms may omit these
			}
			if _, ok := m[key]; !ok {
				t.Errorf("%s: missing translation for %q", lang, key)
			}
		}
	}
}
