package ui

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mrf/envui/env"
)

var compareColumns = []struct {
	key   string
	width float32
}{{"compare.colKey", 210}, {"compare.colLeft", 290}, {"compare.colRight", 290}, {"compare.colStatus", 130}}

// showCompare opens a side-by-side comparison of two env files.
func (a *App) showCompare() {
	if a.project == nil || len(a.project.Files) < 2 {
		a.showInfo(T("compare.title"), T("compare.needTwo"))
		return
	}
	files := a.project.Files
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = filepath.Base(f.Path)
	}
	byName := func(n string) *env.EnvFile {
		for i, name := range names {
			if name == n {
				return files[i]
			}
		}
		return nil
	}

	left := a.activeFile()
	if left == nil {
		left = files[0]
	}
	var right *env.EnvFile
	for _, f := range files { // prefer another real env file over the template
		if f != left && !isExampleFile(f) {
			right = f
			break
		}
	}
	if right == nil {
		for _, f := range files {
			if f != left {
				right = f
				break
			}
		}
	}

	var rows []env.DiffRow
	onlyDiff := true
	summary := widget.NewLabel("")
	summary.Importance = widget.LowImportance

	table := widget.NewTable(
		func() (int, int) { return len(rows), len(compareColumns) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		nil,
	)

	recompute := func() {
		all := env.Compare(left, right)
		rows = rows[:0]
		counts := map[env.DiffKind]int{}
		for _, r := range all {
			counts[r.Kind]++
			if !onlyDiff || r.Kind != env.DiffSame {
				rows = append(rows, r)
			}
		}
		summary.SetText(T("compare.summary", counts[env.DiffChanged], counts[env.DiffOnlyLeft], counts[env.DiffOnlyRight], counts[env.DiffSame]))
		table.Refresh()
	}

	value := func(s string, present, secret bool) (string, widget.Importance) {
		switch {
		case !present:
			return "—", widget.LowImportance
		case s == "":
			return T("cell.empty"), widget.LowImportance
		case secret && !a.showSecrets:
			return secretMask, widget.LowImportance
		}
		return displayValue(s), widget.MediumImportance
	}

	table.UpdateCell = func(id widget.TableCellID, o fyne.CanvasObject) {
		l := o.(*widget.Label)
		l.TextStyle = fyne.TextStyle{}
		l.Importance = widget.MediumImportance
		if id.Row >= len(rows) {
			l.SetText("")
			return
		}
		r := rows[id.Row]
		text := ""
		switch id.Col {
		case 0:
			l.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
			text = r.Key
		case 1:
			l.TextStyle = fyne.TextStyle{Monospace: true}
			text, l.Importance = value(r.Left, r.Kind != env.DiffOnlyRight, r.IsSecret)
		case 2:
			l.TextStyle = fyne.TextStyle{Monospace: true}
			text, l.Importance = value(r.Right, r.Kind != env.DiffOnlyLeft, r.IsSecret)
		case 3:
			switch r.Kind {
			case env.DiffSame:
				text, l.Importance = T("compare.same"), widget.LowImportance
			case env.DiffChanged:
				text, l.Importance = T("compare.changed"), widget.WarningImportance
			case env.DiffOnlyLeft:
				text, l.Importance = T("compare.onlyLeft"), widget.HighImportance
			case env.DiffOnlyRight:
				text, l.Importance = T("compare.onlyRight"), widget.HighImportance
			}
		}
		if l.TextStyle.Monospace {
			text = fitText(text, compareColumns[id.Col].width-2*theme.InnerPadding()-2, l.TextStyle)
		}
		l.Text = text
		l.Refresh()
	}
	table.ShowHeaderRow = true
	table.CreateHeader = func() fyne.CanvasObject {
		l := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		l.Importance = widget.LowImportance
		return l
	}
	table.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		if id.Col < 0 || id.Col >= len(compareColumns) {
			return
		}
		switch id.Col {
		case 1:
			o.(*widget.Label).SetText(filepath.Base(left.Path))
		case 2:
			o.(*widget.Label).SetText(filepath.Base(right.Path))
		default:
			o.(*widget.Label).SetText(T(compareColumns[id.Col].key))
		}
	}
	for i, c := range compareColumns {
		table.SetColumnWidth(i, c.width)
	}
	table.OnSelected = func(id widget.TableCellID) {
		table.UnselectAll()
		if id.Row >= 0 && id.Row < len(rows) {
			a.showCopyBetween(rows[id.Row], left, right, recompute)
		}
	}

	leftSel := widget.NewSelect(names, nil)
	rightSel := widget.NewSelect(names, nil)
	leftSel.SetSelected(filepath.Base(left.Path))
	rightSel.SetSelected(filepath.Base(right.Path))
	leftSel.OnChanged = func(n string) {
		if f := byName(n); f != nil {
			left = f
			recompute()
		}
	}
	rightSel.OnChanged = func(n string) {
		if f := byName(n); f != nil {
			right = f
			recompute()
		}
	}
	swap := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		l, r := leftSel.Selected, rightSel.Selected
		leftSel.SetSelected(r)
		rightSel.SetSelected(l)
	})
	swap.Importance = widget.LowImportance

	diffCheck := widget.NewCheck(T("compare.onlyDiff"), func(on bool) {
		onlyDiff = on
		recompute()
	})
	diffCheck.SetChecked(true)

	hint := widget.NewLabel(T("compare.hint"))
	hint.Importance = widget.LowImportance

	top := container.NewVBox(
		container.NewHBox(leftSel, swap, rightSel, diffCheck),
		summary,
	)
	size := canvas.NewRectangle(nil)
	size.SetMinSize(fyne.NewSize(940, 380))

	recompute()
	d := dialog.NewCustom(T("compare.title"), T("common.close"),
		container.NewBorder(top, hint, nil, nil, container.NewStack(size, table)), a.window)
	d.Show()
}

// showCopyBetween offers to copy one compared value to the other file.
func (a *App) showCopyBetween(r env.DiffRow, left, right *env.EnvFile, done func()) {
	if r.Kind == env.DiffSame {
		return
	}
	ln, rn := filepath.Base(left.Path), filepath.Base(right.Path)
	var d dialog.Dialog
	var buttons []fyne.CanvasObject

	if r.Kind != env.DiffOnlyRight { // left has a value
		b := widget.NewButtonWithIcon(T("compare.copyTo", rn), theme.NavigateNextIcon(), func() {
			d.Hide()
			a.setValue(right, r.Key, r.Left)
			done()
		})
		b.Importance = widget.HighImportance
		buttons = append(buttons, b)
	}
	if r.Kind != env.DiffOnlyLeft { // right has a value
		b := widget.NewButtonWithIcon(T("compare.copyTo", ln), theme.NavigateBackIcon(), func() {
			d.Hide()
			a.setValue(left, r.Key, r.Right)
			done()
		})
		buttons = append(buttons, b)
	}
	msg := widget.NewLabel(T("compare.copyAsk", r.Key))
	d = dialog.NewCustom(r.Key, T("common.cancel"), container.NewVBox(msg, container.NewHBox(buttons...)), a.window)
	d.Show()
}

// setValue sets key in f (adding it at the end if missing) as one undoable step.
func (a *App) setValue(f *env.EnvFile, key, value string) {
	a.setValues([]*env.EnvFile{f}, key, value, true)
	a.flashStatus(T("status.copiedTo", key, filepath.Base(f.Path)))
}

// setValues sets key in every file of targets. Existing values are only
// replaced when overwrite is true.
func (a *App) setValues(targets []*env.EnvFile, key, value string, overwrite bool) {
	a.mutate(targets, func() {
		for _, f := range targets {
			found := false
			vars := copyVars(f.Variables)
			for i := range vars {
				if vars[i].Key == key {
					found = true
					if overwrite {
						vars[i].Value = value
					}
				}
			}
			if !found {
				vars = append(vars, env.Variable{
					Key: key, Value: value, Source: env.SourceFile,
					IsSecret: env.IsSecret(key), Line: -1,
				})
			}
			f.Variables = vars
		}
	})
}

// showCopyToFiles copies a variable from f into other files of the project.
func (a *App) showCopyToFiles(f *env.EnvFile, key, value string) {
	if a.project == nil {
		return
	}
	var others []*env.EnvFile
	for _, o := range a.project.Files {
		if o != f {
			others = append(others, o)
		}
	}
	if len(others) == 0 {
		a.showInfo(T("copyTo.title"), T("compare.needTwo"))
		return
	}

	checks := make([]*widget.Check, len(others))
	box := container.NewVBox()
	for i, o := range others {
		state := T("copyTo.missing")
		for _, v := range o.Variables {
			if v.Key == key {
				if v.Value == value {
					state = T("copyTo.same")
				} else {
					state = T("copyTo.different")
				}
			}
		}
		checks[i] = widget.NewCheck(filepath.Base(o.Path)+"   ("+state+")", nil)
		checks[i].SetChecked(state == T("copyTo.missing") && !isExampleFile(o))
		box.Add(checks[i])
	}
	intro := widget.NewLabel(T("copyTo.intro", key))
	intro.Wrapping = fyne.TextWrapWord
	width := canvas.NewRectangle(nil)
	width.SetMinSize(fyne.NewSize(420, 0))

	content := container.NewVBox(width, intro, box)
	d := dialog.NewCustomConfirm(T("copyTo.title"), T("copyTo.copy"), T("common.cancel"), content, func(ok bool) {
		if !ok {
			return
		}
		var targets []*env.EnvFile
		var names []string
		for i, c := range checks {
			if c.Checked {
				targets = append(targets, others[i])
				names = append(names, filepath.Base(others[i].Path))
			}
		}
		if len(targets) == 0 {
			return
		}
		// ticking a file means "put this value there", replacing any other value
		a.setValues(targets, key, value, true)
		a.flashStatus(T("status.copiedTo", key, joinNames(names)))
	}, a.window)
	d.Show()
}

func joinNames(n []string) string {
	out := ""
	for i, s := range n {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}
