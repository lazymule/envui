//go:build !windows

package ui

// systemLanguages returns nothing on non-Windows systems; the LANG/LC_*
// environment variables are checked by detectLanguage instead.
func systemLanguages() []string { return nil }
