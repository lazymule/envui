package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// envTheme is a modern, slightly more spacious theme with an indigo accent
// and an explicit light/dark variant (independent of the OS setting).
type envTheme struct {
	dark bool
}

var _ fyne.Theme = (*envTheme)(nil)

func newEnvTheme(dark bool) *envTheme { return &envTheme{dark: dark} }

func (t *envTheme) variant() fyne.ThemeVariant {
	if t.dark {
		return theme.VariantDark
	}
	return theme.VariantLight
}

func hex(r, g, b uint8) color.Color { return color.NRGBA{R: r, G: g, B: b, A: 0xff} }

func alpha(c color.Color, a uint8) color.Color {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	n.A = a
	return n
}

func (t *envTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if t.dark {
		switch name {
		case theme.ColorNameBackground:
			return hex(0x16, 0x18, 0x1d)
		case theme.ColorNameHeaderBackground:
			return hex(0x1e, 0x21, 0x28)
		case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
			return hex(0x1e, 0x21, 0x28)
		case theme.ColorNameInputBackground:
			return hex(0x22, 0x25, 0x2d)
		case theme.ColorNameInputBorder:
			return hex(0x33, 0x37, 0x41)
		case theme.ColorNameButton:
			return hex(0x2a, 0x2e, 0x37)
		case theme.ColorNameForeground:
			return hex(0xe6, 0xe8, 0xee)
		case theme.ColorNamePlaceHolder:
			return hex(0x7d, 0x83, 0x92)
		case theme.ColorNameDisabled: // also used for low-importance labels
			return hex(0x8e, 0x94, 0xa3)
		case theme.ColorNameDisabledButton:
			return hex(0x24, 0x27, 0x2f)
		case theme.ColorNameSeparator:
			return hex(0x2a, 0x2e, 0x37)
		case theme.ColorNamePrimary, theme.ColorNameFocus:
			return hex(0x81, 0x8c, 0xf8)
		case theme.ColorNameHyperlink:
			return hex(0x81, 0x8c, 0xf8)
		case theme.ColorNameSelection:
			return alpha(hex(0x81, 0x8c, 0xf8), 0x40)
		case theme.ColorNameHover:
			return alpha(hex(0xff, 0xff, 0xff), 0x0f)
		case theme.ColorNameSuccess:
			return hex(0x34, 0xd3, 0x99)
		case theme.ColorNameWarning:
			return hex(0xfb, 0xbf, 0x24)
		case theme.ColorNameError:
			return hex(0xf8, 0x71, 0x71)
		}
	} else {
		switch name {
		case theme.ColorNameBackground:
			return hex(0xf7, 0xf8, 0xfa)
		case theme.ColorNameHeaderBackground:
			return hex(0xee, 0xf0, 0xf4)
		case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
			return hex(0xff, 0xff, 0xff)
		case theme.ColorNameInputBackground:
			return hex(0xff, 0xff, 0xff)
		case theme.ColorNameInputBorder:
			return hex(0xd5, 0xd9, 0xe0)
		case theme.ColorNameButton:
			return hex(0xe9, 0xec, 0xf1)
		case theme.ColorNameForeground:
			return hex(0x1f, 0x23, 0x2b)
		case theme.ColorNamePlaceHolder:
			return hex(0x8a, 0x90, 0x9c)
		case theme.ColorNameDisabled: // also used for low-importance labels
			return hex(0x6b, 0x72, 0x80)
		case theme.ColorNameDisabledButton:
			return hex(0xe4, 0xe7, 0xec)
		case theme.ColorNameSeparator:
			return hex(0xe2, 0xe5, 0xea)
		case theme.ColorNamePrimary, theme.ColorNameFocus:
			return hex(0x4f, 0x46, 0xe5)
		case theme.ColorNameHyperlink:
			return hex(0x4f, 0x46, 0xe5)
		case theme.ColorNameSelection:
			return alpha(hex(0x4f, 0x46, 0xe5), 0x26)
		case theme.ColorNameHover:
			return alpha(hex(0x00, 0x00, 0x00), 0x0a)
		case theme.ColorNameSuccess:
			return hex(0x05, 0x96, 0x69)
		case theme.ColorNameWarning:
			return hex(0xd9, 0x77, 0x06)
		case theme.ColorNameError:
			return hex(0xdc, 0x26, 0x26)
		}
	}
	return theme.DefaultTheme().Color(name, t.variant())
}

func (t *envTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *envTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *envTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 5
	case theme.SizeNameInnerPadding:
		return 9
	case theme.SizeNameText:
		return 13.5
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 11.5
	case theme.SizeNameInputRadius:
		return 6
	case theme.SizeNameSelectionRadius:
		return 4
	case theme.SizeNameScrollBarSmall:
		return 4
	}
	return theme.DefaultTheme().Size(name)
}
