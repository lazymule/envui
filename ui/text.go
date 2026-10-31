package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const ellipsis = "…"

// fitText shortens s with an ellipsis so that it fits in width when rendered
// with the given style.
//
// Fyne 2.4's built-in TextTruncateEllipsis miscounts multi-byte characters
// (e.g. Turkish "ç", "ğ", "ş" or "•") and cuts text far too early, so we
// measure rune-by-rune ourselves instead.
func fitText(s string, width float32, style fyne.TextStyle) string {
	size := theme.TextSize()
	if width <= 0 || fyne.MeasureText(s, size, style).Width <= width {
		return s
	}
	r := []rune(s)
	lo, hi := 0, len(r)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if fyne.MeasureText(string(r[:mid])+ellipsis, size, style).Width <= width {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return string(r[:lo]) + ellipsis
}
