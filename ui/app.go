package ui

import (
	"errors"
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

const (
	prefDarkMode    = "darkMode"
	prefLanguage    = "language"
	prefShowShell   = "showShellOnStart"
	prefShowSecrets = "showSecretsOnStart"
)

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

	prefs := fyneApp.Preferences()
	SetLanguage(Lang(prefs.StringWithFallback(prefLanguage, string(detectLanguage()))))

	a := &App{
		fyneApp:     fyneApp,
		window:      window,
		dark:        dark,
		showShell:   prefs.Bool(prefShowShell),
		showSecrets: prefs.Bool(prefShowSecrets),
	}
	window.SetTitle(T("app.title"))
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
		return errors.New(T("err.notFolder", abs))
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
	a.themedRects = nil
	a.onThemeChange = nil

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

	a.pathLabel = widget.NewLabel(T("header.noProject"))
	a.pathLabel.Importance = widget.LowImportance

	pathBox := container.NewBorder(nil, nil, widget.NewIcon(theme.FolderIcon()), nil, a.pathLabel)

	a.themeBtn = widget.NewButtonWithIcon("", a.themeIcon(), a.toggleTheme)
	a.themeBtn.Importance = widget.LowImportance

	reloadBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), a.onReload)
	reloadBtn.Importance = widget.LowImportance

	openBtn := widget.NewButtonWithIcon(T("header.open"), theme.FolderOpenIcon(), a.onOpen)

	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), a.showSettings)
	settingsBtn.Importance = widget.LowImportance

	right := container.NewHBox(reloadBtn, a.themeBtn, settingsBtn, openBtn)
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

	title := widget.NewLabelWithStyle(T("sidebar.files"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.LowImportance

	newBtn := widget.NewButtonWithIcon(T("sidebar.newFile"), theme.DocumentCreateIcon(), a.onNewFile)
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

	a.secretsBtn = widget.NewButtonWithIcon("", theme.VisibilityIcon(), func() {
		a.showSecrets = !a.showSecrets
		a.updateSecretsButton()
		a.refresh()
	})
	a.secretsBtn.Importance = widget.LowImportance
	a.updateSecretsButton()

	shellCheck := widget.NewCheck(T("toolbar.shell"), nil)
	shellCheck.Checked = a.showShell
	shellCheck.OnChanged = func(on bool) {
		a.showShell = on
		a.refresh()
	}

	addBtn := widget.NewButtonWithIcon(T("toolbar.add"), theme.ContentAddIcon(), a.onAddVariable)
	addBtn.Importance = widget.HighImportance

	toolbar := container.NewBorder(nil, nil, a.fileTitle,
		container.NewHBox(shellCheck, a.secretsBtn, addBtn))

	a.searchEntry = widget.NewEntry()
	a.searchEntry.SetPlaceHolder(T("toolbar.search"))
	a.searchEntry.ActionItem = widget.NewIcon(theme.SearchIcon())
	a.searchEntry.Text = a.query
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

	a.unsavedLbl = widget.NewLabel(T("status.unsaved"))
	a.unsavedLbl.Importance = widget.WarningImportance
	a.unsavedLbl.Hide()

	a.saveBtn = widget.NewButtonWithIcon(T("status.save"), theme.DocumentSaveIcon(), a.onSave)
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
		view = a.emptyView(T("empty.open.title"), T("empty.open.message"),
			T("header.open"), theme.FolderOpenIcon(), a.onOpen)
	case len(a.project.Files) == 0:
		view = a.emptyView(T("empty.none.title"), T("empty.none.message", a.project.Path),
			T("empty.none.button"), theme.DocumentCreateIcon(), a.onNewFile)
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
		a.pathLabel.SetText(T("header.noProject"))
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
	parts := []string{Tn("stats.vars", len(f.Variables)), Tn("stats.secrets", secrets)}
	if ex := a.exampleFile(); ex != nil && ex != f {
		if a.missingCount > 0 {
			parts = append(parts, T("status.missing", a.missingCount, filepath.Base(ex.Path)))
		} else {
			parts = append(parts, T("status.inSync", filepath.Base(ex.Path)))
		}
	}
	if strings.TrimSpace(a.query) != "" {
		parts = append(parts, Tn("stats.matches", len(a.rows)))
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
		a.secretsBtn.SetText(T("toolbar.hide"))
		a.secretsBtn.SetIcon(theme.VisibilityOffIcon())
	} else {
		a.secretsBtn.SetText(T("toolbar.show"))
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

func (a *App) toggleTheme() { a.setDark(!a.dark) }

func (a *App) setDark(dark bool) {
	if dark == a.dark {
		return
	}
	a.dark = dark
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
	a.showConfirm(T("unsaved.title"), T("unsaved.discardAsk"), T("unsaved.discard"), next)
}

// onOpen handles the open folder button
func (a *App) onOpen() {
	a.confirmDiscard(func() {
		d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			if err := a.LoadProject(uri.Path()); err != nil {
				a.showError(err)
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
			a.showError(err)
			return
		}
		a.selectFile(active)
		a.flashStatus(T("status.reloaded"))
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
		a.showInfo(T("newFile.noProject"), T("newFile.noProjectText"))
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
	fromExample := widget.NewCheck("", nil)
	if ex != nil {
		fromExample.SetText(T("newFile.copy", filepath.Base(ex.Path)))
		fromExample.SetChecked(true)
	} else {
		fromExample.Hide()
	}

	items := []*widget.FormItem{widget.NewFormItem(T("newFile.name"), nameEntry)}
	if ex != nil {
		items = append(items, widget.NewFormItem("", fromExample))
	}

	d := dialog.NewForm(T("newFile.title"), T("common.create"), T("common.cancel"), items, func(ok bool) {
		if !ok {
			return
		}
		name := strings.TrimSpace(nameEntry.Text)
		if name == "" || strings.ContainsAny(name, `/\`) {
			a.showError(errors.New(T("newFile.invalid", name)))
			return
		}
		path := filepath.Join(a.project.Path, name)
		if _, err := os.Stat(path); err == nil {
			a.showError(errors.New(T("newFile.exists", name)))
			return
		}
		var content []byte
		if ex != nil && fromExample.Checked {
			data, err := os.ReadFile(ex.Path)
			if err != nil {
				a.showError(err)
				return
			}
			content = data
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			a.showError(err)
			return
		}
		f, err := env.ParseEnvFile(path)
		if err != nil {
			a.showError(err)
			return
		}
		a.project.Files = append(a.project.Files, f)
		a.showView()
		a.refresh()
		a.selectFile(path)
		a.flashStatus(T("status.created", name))
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
			a.showError(errors.New(T("err.saving", filepath.Base(f.Path), err)))
			a.refresh()
			return
		}
		f.Modified = false
		saved = append(saved, filepath.Base(f.Path))
	}
	a.flashStatus(T("status.saved", strings.Join(saved, ", ")))
	a.refresh()
}

func (a *App) onClose() {
	if len(a.modifiedFiles()) == 0 {
		a.window.Close()
		return
	}
	var d *dialog.CustomDialog
	saveBtn := widget.NewButtonWithIcon(T("unsaved.saveAndQuit"), theme.DocumentSaveIcon(), func() {
		d.Hide()
		a.onSave()
		if len(a.modifiedFiles()) == 0 {
			a.window.Close()
		}
	})
	saveBtn.Importance = widget.HighImportance
	discardBtn := widget.NewButton(T("unsaved.discard"), func() {
		d.Hide()
		a.window.Close()
	})
	discardBtn.Importance = widget.DangerImportance
	cancelBtn := widget.NewButton(T("common.cancel"), func() { d.Hide() })

	d = dialog.NewCustomWithoutButtons(T("unsaved.title"),
		widget.NewLabel(T("unsaved.quitAsk")), a.window)
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

// truncate shortens a string to the given number of characters (runes).
func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-3]) + "..."
}
