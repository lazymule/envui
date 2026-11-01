package ui

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	prefRecent = "recentProjects"
	maxRecent  = 8
)

// recentProjects returns the recently opened project folders that still
// exist, most recent first.
func (a *App) recentProjects() []string {
	var out []string
	for _, p := range strings.Split(a.prefs.String(prefRecent), "\n") {
		if p == "" {
			continue
		}
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			out = append(out, p)
		}
	}
	return out
}

func (a *App) addRecent(path string) {
	list := []string{path}
	for _, p := range a.recentProjects() {
		if p != path && len(list) < maxRecent {
			list = append(list, p)
		}
	}
	a.prefs.SetString(prefRecent, strings.Join(list, "\n"))
}

func (a *App) clearRecent() {
	a.prefs.SetString(prefRecent, "")
	a.showView()
}

func (a *App) openRecent(path string) {
	a.confirmDiscard(func() {
		if err := a.LoadProject(path); err != nil {
			a.showError(err)
		}
	})
}

// showRecentMenu pops up the list of recent projects below the given button.
func (a *App) showRecentMenu(anchor fyne.CanvasObject) {
	var items []*fyne.MenuItem
	for _, p := range a.recentProjects() {
		p := p
		items = append(items, fyne.NewMenuItem(fitText(p, 420, fyne.TextStyle{}), func() { a.openRecent(p) }))
	}
	if len(items) == 0 {
		none := fyne.NewMenuItem(T("recent.none"), nil)
		none.Disabled = true
		items = append(items, none)
	} else {
		items = append(items, fyne.NewMenuItemSeparator(), fyne.NewMenuItem(T("recent.clear"), a.clearRecent))
	}
	menu := fyne.NewMenu("", items...)
	c := a.window.Canvas()
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(anchor)
	widget.ShowPopUpMenuAtPosition(menu, c, pos.Add(fyne.NewPos(0, anchor.Size().Height)))
}

// recentList builds the "recent projects" block of the start screen.
func (a *App) recentList() fyne.CanvasObject {
	recent := a.recentProjects()
	if len(recent) == 0 {
		return nil
	}
	title := widget.NewLabelWithStyle(T("recent.title"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Importance = widget.LowImportance
	box := container.NewVBox(widget.NewSeparator(), title)
	for i, p := range recent {
		if i >= 5 {
			break
		}
		p := p
		btn := widget.NewButtonWithIcon(fitText(filepath.Base(p)+"   —   "+p, 520, fyne.TextStyle{}), theme.FolderIcon(), func() { a.openRecent(p) })
		btn.Importance = widget.LowImportance
		btn.Alignment = widget.ButtonAlignLeading
		box.Add(btn)
	}
	return box
}

// onDropped opens a folder (or the folder of a file) dropped on the window.
func (a *App) onDropped(_ fyne.Position, uris []fyne.URI) {
	if len(uris) == 0 {
		return
	}
	path := uris[0].Path()
	info, err := os.Stat(path)
	if err != nil {
		a.showError(err)
		return
	}
	dir, file := path, ""
	if !info.IsDir() {
		dir, file = filepath.Dir(path), path
	}
	a.confirmDiscard(func() {
		if err := a.LoadProject(dir); err != nil {
			a.showError(err)
			return
		}
		if file != "" {
			a.selectFile(file)
		}
	})
}
