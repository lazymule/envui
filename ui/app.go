package ui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mrf/envui/env"
)

const prefDarkMode = "darkMode"

// themedRect is a background rectangle whose colour follows the theme.
type themedRect struct {
	rect *canvas.Rectangle
	name fyne.ThemeColorName
}

// App represents the main application
type App struct {
	fyneApp fyne.App
	window  fyne.Window
	project *env.Project

	showSecrets bool
	showShell   bool
	dark        bool
	query       string

	rows         []row
	colWidths    [3]float32
	missingCount int
	statusTimer  *time.Timer

	// UI components
	body          *fyne.Container
	mainView      fyne.CanvasObject
	fileList      *widget.List
	varTable      *widget.Table
	pathLabel     *widget.Label
	fileTitle     *widget.RichText
	searchEntry   *widget.Entry
	secretsBtn    *widget.Button
	themeBtn      *widget.Button
	statsLabel    *widget.Label
	statusLabel   *widget.Label
	unsavedLbl    *widget.Label
	saveBtn       *widget.Button
	themedRects   []themedRect
	onThemeChange []func()
}

// NewApp creates a new application instance
func NewApp(fyneApp fyne.App, window fyne.Window) *App {
	dark := fyneApp.Settings().ThemeVariant() == theme.VariantDark
	dark = fyneApp.Preferences().BoolWithFallback(prefDarkMode, dark)

	a := &App{
		fyneApp: fyneApp,
		window:  window,
		dark:    dark,
	}
	fyneApp.Settings().SetTheme(newEnvTheme(dark))
	window.SetCloseIntercept(a.onClose)
	return a
}

// LoadProject loads a project directory
func (a *App) LoadProject(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a folder", abs)
	}

	project, err := env.LoadProject(abs)
	if err != nil {
		return err
	}
	a.project = project
	a.showView()
	if len(project.Files) > 0 {
		a.fileList.UnselectAll()
		a.fileList.Select(0)
	}
	a.refresh()
	return nil
}

// BuildUI constructs the main UI
func (a *App) BuildUI() fyne.CanvasObject {
	header := a.buildHeader()
	a.mainView = container.NewBorder(nil, a.buildStatusBar(), nil, nil, a.buildWorkspace())
	a.body = container.NewStack()
	a.showView()
	a.registerShortcuts()

	return container.NewBorder(header, nil, nil, nil, a.body)
}

// ---------------------------------------------------------------------------
// Layout pieces
// ---------------------------------------------------------------------------

func (a *App) bg(name fyne.ThemeColorName) *canvas.Rectangle {
	r := canvas.NewRectangle(a.themeColor(name))
	a.themedRects = append(a.themedRects, themedRect{rect: r, name: name})
	return r
}

func (a *App) buildHeader() fyne.CanvasObject {
	logo := canvas.NewText("envui", theme.PrimaryColor())
	logo.TextStyle = fyne.TextStyle{Bold: true}
	logo.TextSize = 18

	a.pathLabel = widget.NewLabel("No project open")
	a.pathLabel.Importance = widget.LowImportance

	pathBox := container.NewBorder(nil, nil, widget.NewIcon(theme.FolderIcon()), nil, a.pathLabel)

	a.themeBtn = widget.NewButtonWithIcon("", a.themeIcon(), a.toggleTheme)
	a.themeBtn.Importance = widget.LowImportance

	reloadBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), a.onReload)
	reloadBtn.Importance = widget.LowImportance

	openBtn := widget.NewButtonWithIcon("Open folder", theme.FolderOpenIcon(), a.onOpen)

	right := container.NewHBox(reloadBtn, a.themeBtn, openBtn)
	left := container.NewHBox(container.NewPadded(logo), widget.NewSeparator())
	bar := container.NewBorder(nil, nil, left, right, pathBox)

	// keep the logo colour in sync with theme changes
	a.onThemeChange = append(a.onThemeChange, func() {
		logo.Color = theme.PrimaryColor()
		logo.Refresh()
	})

	return container.NewStack(
		a.bg(theme.ColorNameHeaderBackground),
		container.NewBorder(nil, widget.NewSeparator(), nil, nil, container.NewPadded(bar)),
	)
}

func (a *App) buildSidebar() fyne.CanvasObject {
	a.fileList = widget.NewList(
		func() int {
			if a.project == nil {
				return 0
			}
			return len(a.project.Files)
		},
		func() fyne.CanvasObject {
			name := widget.NewLabel("")
			name.Truncation = fyne.TextTruncateEllipsis
			dot := widget.NewLabelWithStyle("•", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
			dot.Importance = widget.WarningImportance
			return container.NewBorder(nil, nil, widget.NewIcon(theme.FileTextIcon()), dot, name)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if a.project == nil || id >= len(a.project.Files) {
				return
			}
			file := a.project.Files[id]
			c := obj.(*fyne.Container)
			name := c.Objects[0].(*widget.Label)
			icon := c.Objects[1].(*widget.Icon)
			dot := c.Objects[2].(*widget.Label)

			name.SetText(filepath.Base(file.Path))
			if isExampleFile(file) {
				icon.SetResource(theme.DocumentIcon())
				name.Importance = widget.LowImportance
			} else {
				icon.SetResource(theme.FileTextIcon())
				name.Importance = widget.MediumImportance
			}
			name.Refresh()
			if file.Modified {
				dot.Show()
			} else {
				dot.Hide()
			}
		},
	)
	a.fileList.OnSelected = func(id widget.ListItemID) {
		if a.project != nil && id < len(a.project.Files) {
			a.project.ActiveFile = a.project.Files[id]
			a.refresh()
		}
	}

	title := widget.NewLabelWithStyle("FILES", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.LowImportance

	newBtn := widget.NewButtonWithIcon("New .env file", theme.DocumentCreateIcon(), a.onNewFile)
	newBtn.Importance = widget.LowImportance

	return container.NewStack(
		a.bg(theme.ColorNameHeaderBackground),
		container.NewBorder(title, newBtn, nil, nil, a.fileList),
	)
}

func (a *App) buildWorkspace() fyne.CanvasObject {
	a.fileTitle = widget.NewRichText(&widget.TextSegment{
		Style: widget.RichTextStyleSubHeading,
		Text:  "",
	})

	a.secretsBtn = widget.NewButtonWithIcon("Show secrets", theme.VisibilityIcon(), func() {
		a.showSecrets = !a.showSecrets
		a.updateSecretsButton()
		a.refresh()
	})
	a.secretsBtn.Importance = widget.LowImportance

	shellCheck := widget.NewCheck("Shell variables", func(on bool) {
		a.showShell = on
		a.refresh()
	})

	addBtn := widget.NewButtonWithIcon("Add variable", theme.ContentAddIcon(), a.onAddVariable)
	addBtn.Importance = widget.HighImportance

	toolbar := container.NewBorder(nil, nil, a.fileTitle,
		container.NewHBox(shellCheck, a.secretsBtn, addBtn))

	a.searchEntry = widget.NewEntry()
	a.searchEntry.SetPlaceHolder("Search variables…  (Ctrl+F)")
	a.searchEntry.ActionItem = widget.NewIcon(theme.SearchIcon())
	a.searchEntry.OnChanged = func(s string) {
		a.query = s
		a.refresh()
	}

	table := a.buildTable()
	tableCard := container.NewStack(
		a.bg(theme.ColorNameInputBackground),
		container.New(&tableLayout{app: a}, table),
	)

	main := container.NewBorder(
		container.NewVBox(toolbar, a.searchEntry),
		nil, nil, nil,
		tableCard,
	)

	split := container.NewHSplit(a.buildSidebar(), container.NewPadded(main))
	split.SetOffset(0.2)
	return split
}

func (a *App) buildStatusBar() fyne.CanvasObject {
	a.statsLabel = widget.NewLabel("")
	a.statsLabel.Importance = widget.LowImportance

	a.statusLabel = widget.NewLabel("")
	a.statusLabel.Importance = widget.SuccessImportance

	a.unsavedLbl = widget.NewLabel("Unsaved changes")
	a.unsavedLbl.Importance = widget.WarningImportance
	a.unsavedLbl.Hide()

	a.saveBtn = widget.NewButtonWithIcon("Save  (Ctrl+S)", theme.DocumentSaveIcon(), a.onSave)
	a.saveBtn.Importance = widget.HighImportance
	a.saveBtn.Disable()

	bar := container.NewHBox(a.statsLabel, layout.NewSpacer(), a.statusLabel, a.unsavedLbl, a.saveBtn)
	return container.NewStack(
		a.bg(theme.ColorNameHeaderBackground),
		container.NewBorder(widget.NewSeparator(), nil, nil, nil, container.NewPadded(bar)),
	)
}

func (a *App) emptyView(title, message, buttonText string, icon fyne.Resource, action func()) fyne.CanvasObject {
	img := widget.NewIcon(icon)
	heading := widget.NewRichText(&widget.TextSegment{
		Style: widget.RichTextStyle{
			Alignment: fyne.TextAlignCenter,
			SizeName:  theme.SizeNameHeadingText,
			TextStyle: fyne.TextStyle{Bold: true},
		},
		Text: title,
	})
	msg := widget.NewLabelWithStyle(message, fyne.TextAlignCenter, fyne.TextStyle{})
	msg.Importance = widget.LowImportance
	msg.Wrapping = fyne.TextWrapWord

	btn := widget.NewButtonWithIcon(buttonText, icon, action)
	btn.Importance = widget.HighImportance
	if action == nil {
		btn.Hide()
	}

	box := container.NewVBox(
		container.NewCenter(container.NewGridWrap(fyne.NewSize(72, 72), img)),
		heading,
		container.NewGridWrap(fyne.NewSize(420, msg.MinSize().Height*2), msg),
		container.NewCenter(btn),
	)
	return container.NewCenter(box)
}

// showView switches between the empty states and the main editor.
func (a *App) showView() {
	if a.body == nil {
		return
	}
	var view fyne.CanvasObject
	switch {
	case a.project == nil:
		view = a.emptyView("Open a project",
			"Choose a folder that contains .env files to view and edit its environment variables.",
			"Open folder", theme.FolderOpenIcon(), a.onOpen)
	case len(a.project.Files) == 0:
		view = a.emptyView("No .env files found",
			fmt.Sprintf("%s does not contain any .env files yet.", a.project.Path),
			"Create .env", theme.DocumentCreateIcon(), a.onNewFile)
	default:
		view = a.mainView
	}
	a.body.Objects = []fyne.CanvasObject{view}
	a.body.Refresh()
}

// ---------------------------------------------------------------------------
// State updates
// ---------------------------------------------------------------------------

// refresh updates all UI components from the current state
func (a *App) refresh() {
	a.rebuildRows()

	if a.project != nil {
		a.pathLabel.SetText(fitText(a.project.Path, 560, fyne.TextStyle{}))
	} else {
		a.pathLabel.SetText("No project open")
	}

	title := ""
	if f := a.activeFile(); f != nil {
		title = filepath.Base(f.Path)
	}
	if seg, ok := a.fileTitle.Segments[0].(*widget.TextSegment); ok && seg.Text != title {
		seg.Text = title
		a.fileTitle.Refresh()
	}

	a.updateStats()
	a.updateSaveState()

	if a.fileList != nil {
		a.fileList.Refresh()
	}
	if a.varTable != nil {
		a.varTable.Refresh()
	}
}

func (a *App) updateStats() {
	f := a.activeFile()
	if f == nil {
		a.statsLabel.SetText("")
		return
	}
	secrets := 0
	for _, v := range f.Variables {
		if v.IsSecret {
			secrets++
		}
	}
	parts := []string{plural(len(f.Variables), "variable"), plural(secrets, "secret")}
	if ex := a.exampleFile(); ex != nil && ex != f {
		if a.missingCount > 0 {
			parts = append(parts, fmt.Sprintf("%d missing from %s", a.missingCount, filepath.Base(ex.Path)))
		} else {
			parts = append(parts, "in sync with "+filepath.Base(ex.Path))
		}
	}
	if strings.TrimSpace(a.query) != "" {
		parts = append(parts, plural(len(a.rows), "match"))
	}
	a.statsLabel.SetText(strings.Join(parts, "  ·  "))
}

func (a *App) updateSaveState() {
	if len(a.modifiedFiles()) > 0 {
		a.unsavedLbl.Show()
		a.saveBtn.Enable()
	} else {
		a.unsavedLbl.Hide()
		a.saveBtn.Disable()
	}
}

func (a *App) updateSecretsButton() {
	if a.showSecrets {
		a.secretsBtn.SetText("Hide secrets")
		a.secretsBtn.SetIcon(theme.VisibilityOffIcon())
	} else {
		a.secretsBtn.SetText("Show secrets")
		a.secretsBtn.SetIcon(theme.VisibilityIcon())
	}
}

func (a *App) flashStatus(msg string) {
	a.statusLabel.SetText(msg)
	if a.statusTimer != nil {
		a.statusTimer.Stop()
	}
	a.statusTimer = time.AfterFunc(3*time.Second, func() {
		a.statusLabel.SetText("")
	})
}

func (a *App) activeFile() *env.EnvFile {
	if a.project == nil {
		return nil
	}
	return a.project.ActiveFile
}

func (a *App) exampleFile() *env.EnvFile {
	if a.project == nil {
		return nil
	}
	for _, f := range a.project.Files {
		if isExampleFile(f) {
			return f
		}
	}
	return nil
}

func (a *App) modifiedFiles() []*env.EnvFile {
	if a.project == nil {
		return nil
	}
	var out []*env.EnvFile
	for _, f := range a.project.Files {
		if f.Modified {
			out = append(out, f)
		}
	}
	return out
}

func (a *App) markModified(f *env.EnvFile) {
	f.Modified = true
	a.refresh()
}

// ---------------------------------------------------------------------------
// Theme
// ---------------------------------------------------------------------------

func (a *App) themeColor(name fyne.ThemeColorName) color.Color {
	return a.fyneApp.Settings().Theme().Color(name, a.fyneApp.Settings().ThemeVariant())
}

func (a *App) themeIcon() fyne.Resource {
	if a.dark {
		return theme.ColorChromaticIcon()
	}
	return theme.ColorAchromaticIcon()
}

func (a *App) toggleTheme() {
	a.dark = !a.dark
	a.fyneApp.Preferences().SetBool(prefDarkMode, a.dark)
	a.fyneApp.Settings().SetTheme(newEnvTheme(a.dark))
	for _, tr := range a.themedRects {
		if tr.rect == nil {
			continue
		}
		tr.rect.FillColor = a.themeColor(tr.name)
		tr.rect.Refresh()
	}
	for _, fn := range a.onThemeChange {
		fn()
	}
	a.themeBtn.SetIcon(a.themeIcon())
}

// ---------------------------------------------------------------------------
// Actions
// ---------------------------------------------------------------------------

func (a *App) registerShortcuts() {
	c := a.window.Canvas()
	mod := fyne.KeyModifierShortcutDefault
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: mod}, func(fyne.Shortcut) { a.onSave() })
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: mod}, func(fyne.Shortcut) { a.onAddVariable() })
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: mod}, func(fyne.Shortcut) { a.onOpen() })
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: mod}, func(fyne.Shortcut) {
		if a.searchEntry != nil && a.activeFile() != nil {
			c.Focus(a.searchEntry)
		}
	})
}

// confirmDiscard runs next immediately if there are no unsaved changes,
// otherwise asks the user first.
func (a *App) confirmDiscard(next func()) {
	if len(a.modifiedFiles()) == 0 {
		next()
		return
	}
	dialog.ShowConfirm("Unsaved changes",
		"You have unsaved changes. Discard them?",
		func(ok bool) {
			if ok {
				next()
			}
		}, a.window)
}

// onOpen handles the open folder button
func (a *App) onOpen() {
	a.confirmDiscard(func() {
		d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			if err := a.LoadProject(uri.Path()); err != nil {
				dialog.ShowError(err, a.window)
			}
		}, a.window)
		d.Resize(fyne.NewSize(760, 520))
		d.Show()
	})
}

func (a *App) onReload() {
	if a.project == nil {
		return
	}
	a.confirmDiscard(func() {
		active := ""
		if f := a.activeFile(); f != nil {
			active = f.Path
		}
		if err := a.LoadProject(a.project.Path); err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.selectFile(active)
		a.flashStatus("Reloaded from disk")
	})
}

func (a *App) selectFile(path string) {
	for i, f := range a.project.Files {
		if f.Path == path {
			a.fileList.Select(i)
			return
		}
	}
}

// onNewFile creates a new .env file in the project folder. If a template
// (.env.example etc.) exists, its contents are used as a starting point.
func (a *App) onNewFile() {
	if a.project == nil {
		dialog.ShowInformation("No project", "Open a project folder first.", a.window)
		return
	}
	nameEntry := widget.NewEntry()
	nameEntry.SetText(".env")
	for _, candidate := range []string{".env", ".env.local", ".env.development", ".env.staging", ".env.production", ".env.test"} {
		if _, err := os.Stat(filepath.Join(a.project.Path, candidate)); os.IsNotExist(err) {
			nameEntry.SetText(candidate)
			break
		}
	}

	ex := a.exampleFile()
	fromExample := widget.NewCheck("Copy variables from template", nil)
	if ex != nil {
		fromExample.SetText("Copy variables from " + filepath.Base(ex.Path))
		fromExample.SetChecked(true)
	} else {
		fromExample.Hide()
	}

	items := []*widget.FormItem{widget.NewFormItem("File name", nameEntry)}
	if ex != nil {
		items = append(items, widget.NewFormItem("", fromExample))
	}

	d := dialog.NewForm("New env file", "Create", "Cancel", items, func(ok bool) {
		if !ok {
			return
		}
		name := strings.TrimSpace(nameEntry.Text)
		if name == "" || strings.ContainsAny(name, `/\`) {
			dialog.ShowError(fmt.Errorf("invalid file name %q", name), a.window)
			return
		}
		path := filepath.Join(a.project.Path, name)
		if _, err := os.Stat(path); err == nil {
			dialog.ShowError(fmt.Errorf("%s already exists", name), a.window)
			return
		}
		var content []byte
		if ex != nil && fromExample.Checked {
			data, err := os.ReadFile(ex.Path)
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			content = data
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		f, err := env.ParseEnvFile(path)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.project.Files = append(a.project.Files, f)
		a.showView()
		a.refresh()
		a.selectFile(path)
		a.flashStatus("Created " + name)
	}, a.window)
	nameEntry.OnSubmitted = func(string) { d.Submit() }
	d.Resize(fyne.NewSize(420, d.MinSize().Height))
	d.Show()
	a.window.Canvas().Focus(nameEntry)
}

// onSave saves every file with unsaved changes
func (a *App) onSave() {
	files := a.modifiedFiles()
	if len(files) == 0 {
		return
	}
	var saved []string
	for _, f := range files {
		if err := env.WriteEnvFile(f); err != nil {
			dialog.ShowError(fmt.Errorf("saving %s: %w", filepath.Base(f.Path), err), a.window)
			a.refresh()
			return
		}
		f.Modified = false
		saved = append(saved, filepath.Base(f.Path))
	}
	a.flashStatus("Saved " + strings.Join(saved, ", "))
	a.refresh()
}

func (a *App) onClose() {
	if len(a.modifiedFiles()) == 0 {
		a.window.Close()
		return
	}
	var d *dialog.CustomDialog
	saveBtn := widget.NewButtonWithIcon("Save and quit", theme.DocumentSaveIcon(), func() {
		d.Hide()
		a.onSave()
		if len(a.modifiedFiles()) == 0 {
			a.window.Close()
		}
	})
	saveBtn.Importance = widget.HighImportance
	discardBtn := widget.NewButton("Discard", func() {
		d.Hide()
		a.window.Close()
	})
	discardBtn.Importance = widget.DangerImportance
	cancelBtn := widget.NewButton("Cancel", func() { d.Hide() })

	d = dialog.NewCustomWithoutButtons("Unsaved changes",
		widget.NewLabel("You have unsaved changes. Save them before quitting?"), a.window)
	d.SetButtons([]fyne.CanvasObject{cancelBtn, discardBtn, saveBtn})
	d.Show()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func isExampleFile(f *env.EnvFile) bool {
	switch filepath.Base(f.Path) {
	case ".env.example", ".env.sample", ".env.template":
		return true
	}
	return false
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", word)
	}
	if strings.HasSuffix(word, "ch") {
		return fmt.Sprintf("%d %ses", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// truncate shortens a string to the given number of characters (runes).
func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-3]) + "..."
}
