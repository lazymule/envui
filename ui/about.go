package ui

import (
	"image/color"
	"net/url"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mrf/envui/assets"
)

// Version is the application version shown in the About dialog. It can be
// overridden at build time with -ldflags "-X github.com/mrf/envui/ui.Version=1.2.3".
var Version = "0.3.0"

const (
	repoURL     = "https://github.com/lazymule/envui"
	upstreamURL = "https://github.com/mrf/envui"
	fyneURL     = "https://fyne.io"
)

func link(text, raw string) *widget.Hyperlink {
	u, _ := url.Parse(raw)
	return widget.NewHyperlink(text, u)
}

// rendererName describes how the window is being drawn, which helps when
// diagnosing problems on virtual machines and Remote Desktop.
func rendererName() string {
	if d := os.Getenv("GALLIUM_DRIVER"); d != "" {
		return T("about.rendererSoftware", d)
	}
	return T("about.rendererHardware")
}

func (a *App) systemText() string {
	s := runtime.GOOS + " / " + runtime.GOARCH
	if a.portable {
		s += "  ·  " + T("about.portable")
	}
	return s
}

// showAbout opens the About dialog.
func (a *App) showAbout() {
	logo := widget.NewIcon(assets.Icon)

	name := widget.NewRichText(&widget.TextSegment{
		Text: "envui",
		Style: widget.RichTextStyle{
			Alignment: fyne.TextAlignLeading,
			SizeName:  "headingText",
			TextStyle: fyne.TextStyle{Bold: true},
			ColorName: "primary",
		},
	})
	version := widget.NewLabel(T("about.version", Version))
	version.Importance = widget.LowImportance

	desc := widget.NewLabel(T("about.description"))
	desc.Wrapping = fyne.TextWrapWord

	label := func(s string) *widget.Label {
		l := widget.NewLabelWithStyle(s, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
		l.Importance = widget.LowImportance
		return l
	}
	value := func(s string) *widget.Label { return widget.NewLabel(s) }

	info := container.New(&formLayout{},
		label(T("about.developer")), value(T("about.developerName")),
		label(T("about.original")), container.NewHBox(value("Mark Ferree"), link("mrf/envui", upstreamURL)),
		label(T("about.source")), link("github.com/lazymule/envui", repoURL),
		label(T("about.license")), value(T("about.licenseText")),
		label(T("about.builtWith")), container.NewHBox(link("Fyne", fyneURL), value("·  Go "+runtime.Version()[2:])),
		label(T("about.renderer")), value(rendererName()),
		label(T("about.system")), value(a.systemText()),
	)

	thanks := widget.NewLabelWithStyle(T("about.thanks"), fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	thanks.Wrapping = fyne.TextWrapWord
	thanks.Importance = widget.LowImportance

	width := canvas.NewRectangle(color.Transparent)
	width.SetMinSize(fyne.NewSize(520, 0))

	// logo on the left, name / version / description on the right
	header := container.NewBorder(nil, nil,
		container.NewVBox(container.NewGridWrap(fyne.NewSize(88, 88), logo)),
		nil,
		container.NewVBox(name, version, desc),
	)

	content := container.NewVBox(
		width,
		header,
		widget.NewSeparator(),
		info,
		widget.NewSeparator(),
		thanks,
	)

	var d *dialog.CustomDialog
	updateBtn := widget.NewButtonWithIcon(T("update.check"), theme.DownloadIcon(), func() {
		d.Hide()
		a.checkForUpdatesNow()
	})
	closeBtn := widget.NewButton(T("common.close"), func() { d.Hide() })
	closeBtn.Importance = widget.HighImportance
	d = dialog.NewCustomWithoutButtons(T("about.title"), content, a.window)
	d.SetButtons([]fyne.CanvasObject{updateBtn, closeBtn})
	d.Show()
}

// formLayout is a two-column layout like widget.Form, but with the labels
// right-aligned to the width of the widest label.
type formLayout struct{}

func (formLayout) widths(objs []fyne.CanvasObject) (float32, float32) {
	var left, right float32
	for i := 0; i+1 < len(objs); i += 2 {
		if w := objs[i].MinSize().Width; w > left {
			left = w
		}
		if w := objs[i+1].MinSize().Width; w > right {
			right = w
		}
	}
	return left, right
}

func (l formLayout) rowHeight(objs []fyne.CanvasObject, i int) float32 {
	h := objs[i].MinSize().Height
	if h2 := objs[i+1].MinSize().Height; h2 > h {
		h = h2
	}
	return h
}

func (l formLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	left, _ := l.widths(objs)
	var y float32
	for i := 0; i+1 < len(objs); i += 2 {
		h := l.rowHeight(objs, i)
		objs[i].Move(fyne.NewPos(0, y))
		objs[i].Resize(fyne.NewSize(left, h))
		objs[i+1].Move(fyne.NewPos(left, y))
		objs[i+1].Resize(fyne.NewSize(size.Width-left, h))
		y += h
	}
}

func (l formLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	left, right := l.widths(objs)
	var h float32
	for i := 0; i+1 < len(objs); i += 2 {
		h += l.rowHeight(objs, i)
	}
	return fyne.NewSize(left+right, h)
}
