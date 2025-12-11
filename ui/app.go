package ui

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/mrf/envui/env"
)

// App represents the main application
type App struct {
	window      fyne.Window
	project     *env.Project
	showSecrets bool

	// UI components
	fileList    *widget.List
	varTable    *widget.Table
	pathLabel   *widget.Label
	statusLabel *widget.Label
}

// NewApp creates a new application instance
func NewApp(window fyne.Window) *App {
	return &App{
		window:      window,
		showSecrets: false,
	}
}

// LoadProject loads a project directory
func (a *App) LoadProject(path string) error {
	project, err := env.LoadProject(path)
	if err != nil {
		return err
	}
	a.project = project
	a.refresh()
	return nil
}

// BuildUI constructs the main UI
func (a *App) BuildUI() fyne.CanvasObject {
	// Top: Path display and open button
	a.pathLabel = widget.NewLabel("No project loaded")
	openBtn := widget.NewButton("Open...", a.onOpen)
	topBar := container.NewBorder(nil, nil, nil, openBtn, a.pathLabel)

	// Left: File list
	a.fileList = widget.NewList(
		func() int {
			if a.project == nil {
				return 0
			}
			return len(a.project.Files)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if a.project == nil || id >= len(a.project.Files) {
				return
			}
			label := obj.(*widget.Label)
			file := a.project.Files[id]
			name := filepath.Base(file.Path)
			if file.Modified {
				name = "● " + name
			}
			label.SetText(name)
		},
	)
	a.fileList.OnSelected = func(id widget.ListItemID) {
		if a.project != nil && id < len(a.project.Files) {
			a.project.ActiveFile = a.project.Files[id]
			a.varTable.Refresh()
		}
	}

	// Right: Variable table
	a.varTable = widget.NewTable(
		func() (int, int) {
			return len(a.activeVars()), 3
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			vars := a.activeVars()
			if id.Row >= len(vars) {
				return
			}
			v := vars[id.Row]
			label := obj.(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(v.Key)
			case 1:
				if v.IsSecret && !a.showSecrets {
					label.SetText("••••••••••••")
				} else {
					label.SetText(truncate(v.Value, 40))
				}
			case 2:
				label.SetText(v.Source.String())
			}
		},
	)
	a.varTable.SetColumnWidth(0, 200)
	a.varTable.SetColumnWidth(1, 300)
	a.varTable.SetColumnWidth(2, 80)
	a.varTable.OnSelected = func(id widget.TableCellID) {
		a.onEditVariable(id.Row)
	}

	// Table header
	header := container.NewHBox(
		widget.NewLabelWithStyle("Variable", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewLabelWithStyle("Value", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewLabelWithStyle("Source", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	tableWithHeader := container.NewBorder(header, nil, nil, nil, a.varTable)

	// Bottom: Action buttons
	a.statusLabel = widget.NewLabel("")
	secretsCheck := widget.NewCheck("Show Secrets", func(show bool) {
		a.showSecrets = show
		a.varTable.Refresh()
	})

	actions := container.NewHBox(
		widget.NewButton("+ Add Variable", a.onAddVariable),
		secretsCheck,
		layout.NewSpacer(),
		a.statusLabel,
		widget.NewButton("Save", a.onSave),
	)

	// Layout
	fileListWithLabel := container.NewBorder(
		widget.NewLabelWithStyle("Files", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		a.fileList,
	)

	split := container.NewHSplit(fileListWithLabel, tableWithHeader)
	split.SetOffset(0.2)

	return container.NewBorder(
		topBar,  // top
		actions, // bottom
		nil,     // left
		nil,     // right
		split,   // center
	)
}

// activeVars returns the variables to display in the table
func (a *App) activeVars() []env.Variable {
	if a.project == nil || a.project.ActiveFile == nil {
		return nil
	}

	// Start with file variables
	vars := make([]env.Variable, len(a.project.ActiveFile.Variables))
	copy(vars, a.project.ActiveFile.Variables)

	return vars
}

// refresh updates all UI components
func (a *App) refresh() {
	if a.project != nil {
		a.pathLabel.SetText(a.project.Path)
	}
	if a.fileList != nil {
		a.fileList.Refresh()
	}
	if a.varTable != nil {
		a.varTable.Refresh()
	}
}

// onOpen handles the open folder button
func (a *App) onOpen() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		if err := a.LoadProject(uri.Path()); err != nil {
			dialog.ShowError(err, a.window)
		}
	}, a.window)
}

// onAddVariable handles adding a new variable
func (a *App) onAddVariable() {
	if a.project == nil || a.project.ActiveFile == nil {
		dialog.ShowInformation("No File", "Please open a project and select a file first.", a.window)
		return
	}

	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("VARIABLE_NAME")
	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("value")

	form := dialog.NewForm("Add Variable", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Key", keyEntry),
			widget.NewFormItem("Value", valueEntry),
		},
		func(ok bool) {
			if !ok || keyEntry.Text == "" {
				return
			}
			a.project.ActiveFile.Variables = append(a.project.ActiveFile.Variables, env.Variable{
				Key:      keyEntry.Text,
				Value:    valueEntry.Text,
				Source:   env.SourceFile,
				IsSecret: env.IsSecret(keyEntry.Text),
				Line:     -1, // New variable, will be appended
			})
			a.project.ActiveFile.Modified = true
			a.refresh()
		},
		a.window,
	)
	form.Resize(fyne.NewSize(400, 200))
	form.Show()
}

// onEditVariable handles editing a variable
func (a *App) onEditVariable(row int) {
	vars := a.activeVars()
	if row >= len(vars) {
		return
	}
	if a.project == nil || a.project.ActiveFile == nil {
		return
	}

	// Find the variable in the active file
	if row >= len(a.project.ActiveFile.Variables) {
		return
	}
	v := &a.project.ActiveFile.Variables[row]

	keyEntry := widget.NewEntry()
	keyEntry.SetText(v.Key)
	valueEntry := widget.NewEntry()
	valueEntry.SetText(v.Value)

	deleteBtn := widget.NewButton("Delete", func() {
		// Remove variable
		a.project.ActiveFile.Variables = append(
			a.project.ActiveFile.Variables[:row],
			a.project.ActiveFile.Variables[row+1:]...,
		)
		a.project.ActiveFile.Modified = true
		a.refresh()
	})

	form := dialog.NewForm("Edit Variable", "Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Key", keyEntry),
			widget.NewFormItem("Value", valueEntry),
			widget.NewFormItem("", deleteBtn),
		},
		func(ok bool) {
			if !ok {
				return
			}
			v.Key = keyEntry.Text
			v.Value = valueEntry.Text
			v.IsSecret = env.IsSecret(keyEntry.Text)
			a.project.ActiveFile.Modified = true
			a.refresh()
		},
		a.window,
	)
	form.Resize(fyne.NewSize(400, 250))
	form.Show()
}

// onSave handles saving the current file
func (a *App) onSave() {
	if a.project == nil || a.project.ActiveFile == nil {
		return
	}

	if err := env.WriteEnvFile(a.project.ActiveFile); err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	a.project.ActiveFile.Modified = false
	a.statusLabel.SetText("Saved!")
	a.refresh()
}

// truncate shortens a string to the given length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
