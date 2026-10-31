package ui

import (
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mrf/envui/env"
)

const secretMask = "••••••••••••"

// blankIcon keeps the icon slot's size without drawing anything
// (Fyne 2.4's Icon keeps showing the previous image when set to nil).
var blankIcon = fyne.NewStaticResource("blank.svg",
	[]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"></svg>`))

// rowKind describes where a table row comes from.
type rowKind int

const (
	rowFile    rowKind = iota // defined in the active file
	rowMissing                // in the template (.env.example) but not in the active file
	rowShell                  // inherited from the shell environment only
)

// row is one line of the variable table.
type row struct {
	v    env.Variable
	idx  int // index into ActiveFile.Variables, -1 if not in the file
	kind rowKind
	hint string // template value for missing rows
}

var columnKeys = []string{"col.key", "col.value", "col.source"}

// rebuildRows recomputes the visible rows from the active file, the
// template file, the shell and the search query.
func (a *App) rebuildRows() {
	a.rows = a.rows[:0]
	a.missingCount = 0

	f := a.activeFile()
	if f == nil {
		return
	}

	q := strings.ToLower(strings.TrimSpace(a.query))
	match := func(key, value string, secret bool) bool {
		if q == "" {
			return true
		}
		if strings.Contains(strings.ToLower(key), q) {
			return true
		}
		return (!secret || a.showSecrets) && strings.Contains(strings.ToLower(value), q)
	}

	present := make(map[string]bool, len(f.Variables))
	for i, v := range f.Variables {
		present[v.Key] = true
		if match(v.Key, v.Value, v.IsSecret) {
			a.rows = append(a.rows, row{v: v, idx: i, kind: rowFile})
		}
	}

	if ex := a.exampleFile(); ex != nil && ex != f {
		for _, v := range ex.Variables {
			if present[v.Key] {
				continue
			}
			present[v.Key] = true
			a.missingCount++
			if match(v.Key, "", false) {
				a.rows = append(a.rows, row{
					v:    env.Variable{Key: v.Key, Source: env.SourceMissing, IsSecret: v.IsSecret, Line: -1},
					idx:  -1,
					kind: rowMissing,
					hint: v.Value,
				})
			}
		}
	}

	if a.showShell && a.project != nil {
		shell := make([]env.Variable, 0, len(a.project.ShellVars))
		for _, v := range a.project.ShellVars {
			if !present[v.Key] && match(v.Key, v.Value, v.IsSecret) {
				shell = append(shell, v)
			}
		}
		sort.Slice(shell, func(i, j int) bool { return shell[i].Key < shell[j].Key })
		for _, v := range shell {
			a.rows = append(a.rows, row{v: v, idx: -1, kind: rowShell})
		}
	}
}

// newCell builds the template used for every table cell:
// [icon] label
func newCell() fyne.CanvasObject {
	label := widget.NewLabel("")
	icon := widget.NewIcon(blankIcon)
	return container.NewBorder(nil, nil, icon, nil, label)
}

func (a *App) buildTable() *widget.Table {
	a.varTable = widget.NewTable(
		func() (int, int) { return len(a.rows), len(columnKeys) },
		newCell,
		a.updateCell,
	)
	a.varTable.ShowHeaderRow = true
	a.varTable.CreateHeader = func() fyne.CanvasObject {
		l := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		l.Importance = widget.LowImportance
		return l
	}
	a.varTable.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {
		if id.Col >= 0 && id.Col < len(columnKeys) {
			obj.(*widget.Label).SetText(T(columnKeys[id.Col]))
		}
	}
	a.varTable.OnSelected = func(id widget.TableCellID) {
		a.varTable.UnselectAll()
		if id.Row >= 0 && id.Row < len(a.rows) {
			a.onRowActivated(a.rows[id.Row])
		}
	}
	return a.varTable
}

func (a *App) updateCell(id widget.TableCellID, obj fyne.CanvasObject) {
	c := obj.(*fyne.Container)
	label := c.Objects[0].(*widget.Label)
	icon := c.Objects[1].(*widget.Icon)

	if id.Row >= len(a.rows) {
		label.SetText("")
		icon.Hide()
		return
	}
	r := a.rows[id.Row]

	label.TextStyle = fyne.TextStyle{}
	label.Importance = widget.MediumImportance
	icon.SetResource(blankIcon)
	// keep an (empty) icon slot in the key column so keys stay aligned
	icon.Hidden = id.Col != 0

	switch id.Col {
	case 0: // key
		label.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
		switch {
		case r.kind == rowMissing:
			icon.SetResource(theme.NewWarningThemedResource(theme.WarningIcon()))
		case r.v.IsSecret:
			icon.SetResource(theme.NewDisabledResource(theme.VisibilityOffIcon()))
		}
		if r.kind == rowShell {
			label.Importance = widget.LowImportance
		}
		label.Text = r.v.Key

	case 1: // value
		label.TextStyle = fyne.TextStyle{Monospace: true}
		switch {
		case r.kind == rowMissing:
			// note: plain style + ASCII only, the italic/monospace fonts lack
			// some glyphs which breaks truncation measurement
			label.TextStyle = fyne.TextStyle{}
			label.Importance = widget.LowImportance
			if r.hint != "" && !r.v.IsSecret {
				label.Text = T("cell.notSetExample", displayValue(r.hint))
			} else {
				label.Text = T("cell.notSet")
			}
		case r.v.IsSecret && !a.showSecrets:
			label.TextStyle = fyne.TextStyle{}
			label.Importance = widget.LowImportance
			label.Text = secretMask
		case r.v.Value == "":
			label.TextStyle = fyne.TextStyle{}
			label.Importance = widget.LowImportance
			label.Text = T("cell.empty")
		default:
			if r.kind == rowShell {
				label.Importance = widget.LowImportance
			}
			label.Text = displayValue(r.v.Value)
		}

	case 2: // source
		switch r.kind {
		case rowFile:
			label.Text = T("source.file")
			label.Importance = widget.SuccessImportance
		case rowMissing:
			label.Text = T("source.missing")
			label.Importance = widget.WarningImportance
		case rowShell:
			label.Text = T("source.shell")
			label.Importance = widget.LowImportance
		}
	}

	// fit the text to the column (see fitText for why we don't use Fyne's truncation)
	avail := a.colWidths[id.Col] - 2*theme.InnerPadding() - 2
	if id.Col == 0 {
		avail -= theme.IconInlineSize() + theme.Padding()
	}
	label.Text = fitText(label.Text, avail, label.TextStyle)
	label.Refresh()
}

// displayValue makes a value safe for single-line display.
func displayValue(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return truncate(s, 200)
}

func (a *App) onRowActivated(r row) {
	f := a.activeFile()
	if f == nil {
		return
	}
	switch r.kind {
	case rowFile:
		a.showVariableDialog(f, r.idx, r.v.Key, r.v.Value)
	case rowMissing:
		value := r.hint
		if r.v.IsSecret {
			value = ""
		}
		a.showVariableDialog(f, -1, r.v.Key, value)
	case rowShell:
		a.showVariableDialog(f, -1, r.v.Key, r.v.Value)
	}
}

// tableLayout resizes the table and spreads the columns across the
// available width so the table always fills its card.
type tableLayout struct {
	app    *App
	widths [3]float32
}

func (l *tableLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objs {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(size)
	}
	t := l.app.varTable
	if t == nil {
		return
	}
	pad := theme.Padding()
	source := float32(100)
	avail := size.Width - source - 4*pad
	key := avail * 0.36
	if key < 160 {
		key = 160
	}
	value := avail - key
	if value < 160 {
		value = 160
	}
	widths := [3]float32{key, value, source}
	if widths == l.widths {
		return
	}
	l.widths = widths
	l.app.colWidths = widths
	for i, w := range widths {
		t.SetColumnWidth(i, w)
	}
}

func (l *tableLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(480, 200)
}
