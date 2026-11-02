package main

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"

	"github.com/mrf/envui/assets"
	"github.com/mrf/envui/ui"
)

func main() {
	if relaunchWithSoftwareGL() {
		return
	}

	a := app.NewWithID("io.github.mrf.envui")
	a.SetIcon(assets.Icon)
	w := a.NewWindow("envui")
	w.SetIcon(assets.Icon)
	w.Resize(fyne.NewSize(1100, 680))
	w.CenterOnScreen()

	appUI := ui.NewApp(a, w)
	w.SetContent(appUI.BuildUI())

	// A folder or an env file given on the command line (also used by the
	// "Open with envui" entry of the Windows context menu)
	if len(os.Args) > 1 {
		if err := appUI.OpenPath(os.Args[1]); err != nil {
			dialog.ShowError(err, w)
		}
	}

	appUI.AfterStart()
	w.ShowAndRun()
}
