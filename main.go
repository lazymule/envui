package main

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"

	"github.com/mrf/envui/ui"
)

func main() {
	a := app.NewWithID("io.github.mrf.envui")
	w := a.NewWindow("envui")
	w.Resize(fyne.NewSize(1100, 680))
	w.CenterOnScreen()

	appUI := ui.NewApp(a, w)
	w.SetContent(appUI.BuildUI())

	// If a path was provided as argument, load it
	if len(os.Args) > 1 {
		if err := appUI.LoadProject(os.Args[1]); err != nil {
			dialog.ShowError(err, w)
		}
	}

	w.ShowAndRun()
}
