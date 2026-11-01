package ui

import (
	"path/filepath"
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
	rowSection                // a section heading ("# ==== / # Title / # ====")
)

// row is one line of the variable table.
type row struct {
	v     env.Variable
	idx   int // index into ActiveFile.Variables, -1 if not in the file
	kind  rowKind
	hint  string // template value for missing rows
	title string // section rows
}

// table columns
const (
	colKey = iota
	colValue
	colDesc
	colSource
	numCols
)

var columnKeys = []string{"col.key", "col.value", "col.desc", "col.source"}

// description returns the text shown in the description column and whether
// it is inherited from the variable above (shown dimmer).
func description(v env.Variable) (string, bool) {
	switch {
	case v.Doc != "":
		return v.Doc, false
	case v.InlineComment != "":
		return strings.TrimSpace(strings.TrimPrefix(v.InlineComment, "#")), false
	case v.GroupDoc != "":
		return v.GroupDoc, true
	}
	return "", false
}

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
	match := func(v env.Variable, value string) bool {
		if q == "" {
			return true
		}
		desc, _ := description(v)
		for _, s := range []string{v.Key, desc, v.Section} {
			if strings.Contains(strings.ToLower(s), q) {
				return true
			}
		}
		return (!v.IsSecret || a.showSecrets) && strings.Contains(strings.ToLower(value), q)
	}

	section := ""
	addSection := func(title string) {
		if title != "" && title != section {
			a.rows = append(a.rows, row{kind: rowSection, idx: -1, title: title})
		}
		section = title
	}

	present := make(map[string]bool, len(f.Variables))
	for i, v := range f.Variables {
		present[v.Key] = true
		if match(v, v.Value) {
			if v.Section != "" || v.Line >= 0 { // new variables stay in the last section
				addSection(v.Section)
			}
			a.rows = append(a.rows, row{v: v, idx: i, kind: rowFile})
		}
	}

	if ex := a.exampleFile(); ex != nil && ex != f {
		first := true
		for _, v := range ex.Variables {
			if present[v.Key] {
				continue
			}
			present[v.Key] = true
			a.missingCount++
			if match(v, "") {
				if first {
					addSection(T("section.missing", filepath.Base(ex.Path)))
					first = false
				}
				a.rows = append(a.rows, row{
					v: env.Variable{Key: v.Key, Source: env.SourceMissing, IsSecret: v.IsSecret, Line: -1,
						Doc: v.Doc, GroupDoc: v.GroupDoc, InlineComment: v.InlineComment, Section: v.Section},
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
			if !present[v.Key] && match(v, v.Value) {
				shell = append(shell, v)
			}
		}
		sort.Slice(shell, func(i, j int) bool { return shell[i].Key < shell[j].Key })
		if len(shell) > 0 {
			addSection(T("section.shell"))
		}
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
		func() (int, int) { return len(a.rows), numCols },
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

// splitTitle splits "Database (SQL Server login; ...)" into the name shown in
// the key column and the explanation shown in the description column.
func splitTitle(t string) (string, string) {
	if i := strings.Index(t, " ("); i > 0 {
		return t[:i], strings.TrimSuffix(strings.TrimSpace(t[i+1:]), "")
	}
	if i := strings.Index(t, " - "); i > 0 {
		return t[:i], t[i+3:]
	}
	return t, ""
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
	icon.Hidden = id.Col != colKey
	label.Text = ""

	if r.kind == rowSection {
		name, rest := splitTitle(r.title)
		switch id.Col {
		case colKey:
			icon.SetResource(theme.NewPrimaryThemedResource(theme.ListIcon()))
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Importance = widget.HighImportance
			label.Text = name
		case colValue:
			label.Importance = widget.LowImportance
			label.Text = rest
		}
		a.fitCell(id.Col, label)
		return
	}

	switch id.Col {
	case colKey:
		label.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
		switch {
		case r.kind == rowMissing:
			icon.SetResource(theme.NewWarningThemedResource(theme.WarningIcon()))
		case r.kind == rowFile && a.issueIdx[r.idx]:
			icon.SetResource(theme.NewErrorThemedResource(theme.ErrorIcon()))
		case r.v.IsSecret:
			icon.SetResource(theme.NewDisabledResource(theme.VisibilityOffIcon()))
		}
		if r.kind == rowShell {
			label.Importance = widget.LowImportance
		}
		label.Text = r.v.Key

	case colValue:
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
		case r.v.Value == "":
			label.TextStyle = fyne.TextStyle{}
			label.Importance = widget.LowImportance
			label.Text = T("cell.empty")
		case r.v.IsSecret && !a.showSecrets:
			label.TextStyle = fyne.TextStyle{}
			label.Importance = widget.LowImportance
			label.Text = secretMask
		default:
			if r.kind == rowShell {
				label.Importance = widget.LowImportance
			}
			label.Text = displayValue(r.v.Value)
		}

	case colDesc:
		desc, inherited := description(r.v)
		label.Text = strings.Join(strings.Fields(desc), " ")
		if inherited { // shared with the variable above: shown dimmer
			label.Importance = widget.LowImportance
		}

	case colSource:
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
	a.fitCell(id.Col, label)
}

// fitCell shortens the label text to its column (see fitText for why we
// don't use Fyne's truncation) and redraws it.
func (a *App) fitCell(col int, label *widget.Label) {
	avail := a.colWidths[col] - 2*theme.InnerPadding() - 2
	if col == colKey {
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
		a.showVariableDialog(f, r.idx, r.v.Key, r.v.Value, "")
	case rowMissing:
		value := r.hint
		if r.v.IsSecret {
			value = ""
		}
		a.showVariableDialog(f, -1, r.v.Key, value, r.v.Doc)
	case rowShell:
		a.showVariableDialog(f, -1, r.v.Key, r.v.Value, "")
	}
}

// tableLayout resizes the table and spreads the columns across the
// available width so the table always fills its card.
type tableLayout struct {
	app    *App
	widths [numCols]float32
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
	source := float32(90)
	avail := size.Width - source - 5*pad
	key := max32(avail*0.28, 150)
	value := max32(avail*0.30, 140)
	desc := max32(avail-key-value, 140)
	widths := [numCols]float32{key, value, desc, source}
	if widths == l.widths {
		return
	}
	l.widths = widths
	l.app.colWidths = widths
	for i, w := range widths {
		t.SetColumnWidth(i, w)
	}
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func (l *tableLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(560, 200)
}
