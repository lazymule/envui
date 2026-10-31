package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// envTheme is a modern, slightly more spacious theme using the GLARK blues
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

// Brand palette, taken from the GLARK logo.
var (
	brandBlue  = hex(0x07, 0x4e, 0x8a) // main logo blue
	brandSky   = hex(0x9a, 0xd3, 0xf2) // light gradient end
	brandMid   = hex(0x63, 0x9c, 0xc4) // gradient middle
	brandNavy  = hex(0x15, 0x36, 0x54) // dark accent
	brandAzure = hex(0x5a, 0xae, 0xe3) // brighter blue for dark mode
)

func (t *envTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if t.dark {
		// navy surfaces with sky-blue accents
		switch name {
		case theme.ColorNameBackground:
			return hex(0x0c, 0x18, 0x26)
		case theme.ColorNameHeaderBackground:
			return hex(0x11, 0x22, 0x36)
		case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
			return hex(0x12, 0x25, 0x3a)
		case theme.ColorNameInputBackground:
			return hex(0x10, 0x20, 0x33)
		case theme.ColorNameInputBorder:
			return hex(0x24, 0x40, 0x5e)
		case theme.ColorNameButton:
			return hex(0x1a, 0x33, 0x4f)
		case theme.ColorNameForeground:
			return hex(0xe8, 0xf1, 0xf8)
		case theme.ColorNamePlaceHolder:
			return hex(0x7a, 0x93, 0xab)
		case theme.ColorNameDisabled: // also used for low-importance labels
			return hex(0x8f, 0xa8, 0xc0)
		case theme.ColorNameDisabledButton:
			return hex(0x16, 0x2b, 0x42)
		case theme.ColorNameSeparator:
			return hex(0x1d, 0x35, 0x50)
		case theme.ColorNamePrimary:
			return brandAzure
		case theme.ColorNameFocus, theme.ColorNameHyperlink:
			return brandSky
		case theme.ColorNameSelection:
			return alpha(brandAzure, 0x45)
		case theme.ColorNameHover:
			return alpha(brandSky, 0x12)
		case theme.ColorNamePressed:
			return alpha(brandSky, 0x25)
		case theme.ColorNameScrollBar:
			return alpha(brandMid, 0x99)
		case theme.ColorNameSuccess:
			return hex(0x4a, 0xd6, 0x9d)
		case theme.ColorNameWarning:
			return hex(0xf5, 0xb9, 0x42)
		case theme.ColorNameError:
			return hex(0xf2, 0x76, 0x76)
		}
	} else {
		// white/ice-blue surfaces with the GLARK blue as accent
		switch name {
		case theme.ColorNameBackground:
			return hex(0xf5, 0xf9, 0xfc)
		case theme.ColorNameHeaderBackground:
			return hex(0xe9, 0xf2, 0xf9)
		case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
			return hex(0xff, 0xff, 0xff)
		case theme.ColorNameInputBackground:
			return hex(0xff, 0xff, 0xff)
		case theme.ColorNameInputBorder:
			return hex(0xc9, 0xdb, 0xea)
		case theme.ColorNameButton:
			return hex(0xe1, 0xec, 0xf5)
		case theme.ColorNameForeground:
			return hex(0x1c, 0x2a, 0x38)
		case theme.ColorNamePlaceHolder:
			return hex(0x86, 0x98, 0xa9)
		case theme.ColorNameDisabled: // also used for low-importance labels
			return hex(0x5f, 0x74, 0x88)
		case theme.ColorNameDisabledButton:
			return hex(0xe3, 0xea, 0xf0)
		case theme.ColorNameSeparator:
			return hex(0xd8, 0xe5, 0xf0)
		case theme.ColorNamePrimary:
			return brandBlue
		case theme.ColorNameFocus:
			return brandMid
		case theme.ColorNameHyperlink:
			return brandBlue
		case theme.ColorNameSelection:
			return alpha(brandSky, 0x80)
		case theme.ColorNameHover:
			return alpha(brandBlue, 0x0d)
		case theme.ColorNamePressed:
			return alpha(brandBlue, 0x20)
		case theme.ColorNameScrollBar:
			return alpha(brandNavy, 0x66)
		case theme.ColorNameSuccess:
			return hex(0x0f, 0x8a, 0x5f)
		case theme.ColorNameWarning:
			return hex(0xc7, 0x6a, 0x04)
		case theme.ColorNameError:
			return hex(0xd0, 0x2f, 0x2f)
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
