package main

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/mrf/envui/ui"
)

func main() {
	a := app.New()
	w := a.NewWindow("envui - Environment Variable Manager")
	w.Resize(fyne.NewSize(900, 600))

	appUI := ui.NewApp(w)
	w.SetContent(appUI.BuildUI())

	// If a path was provided as argument, load it
	if len(os.Args) > 1 {
		if err := appUI.LoadProject(os.Args[1]); err != nil {
			// Silently fail, user can open manually
		}
	}

	w.ShowAndRun()
}
