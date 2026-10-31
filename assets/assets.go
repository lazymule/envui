// Package assets holds files embedded into the envui binary.
package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed icon.png
var iconPNG []byte

// Icon is the application icon (512x512 PNG). The Windows .exe icon is
// generated separately from winres/envui.ico, see winres/envui.rc.
var Icon = fyne.NewStaticResource("icon.png", iconPNG)
