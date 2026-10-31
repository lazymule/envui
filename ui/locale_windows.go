//go:build windows

package ui

import "golang.org/x/sys/windows"

// systemLanguages returns the user's preferred UI languages (e.g. "tr-TR").
func systemLanguages() []string {
	const muiLanguageName = 0x8
	langs, err := windows.GetUserPreferredUILanguages(muiLanguageName)
	if err != nil {
		return nil
	}
	return langs
}
