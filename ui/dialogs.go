package ui

import (
	"errors"
	"fmt"
	"image/color"
	"path/filepath"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mrf/envui/env"
)

var validKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.\-]*$`)

func validateKey(key string) error {
	if key == "" {
		return errors.New("name is required")
	}
	if !validKey.MatchString(key) {
		return errors.New("use letters, digits and _ (must not start with a digit)")
	}
	return nil
}

// onAddVariable handles adding a new variable
func (a *App) onAddVariable() {
	f := a.activeFile()
	if f == nil {
		dialog.ShowInformation("No file selected", "Open a project and select a file first.", a.window)
		return
	}
	a.showVariableDialog(f, -1, "", "")
}

// showVariableDialog shows the add/edit dialog. idx is the index of the
// variable in f.Variables, or -1 to add a new one.
func (a *App) showVariableDialog(f *env.EnvFile, idx int, key, value string) {
	editing := idx >= 0

	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("VARIABLE_NAME")
	keyEntry.TextStyle = fyne.TextStyle{Monospace: true}
	keyEntry.SetText(key)
	keyEntry.Validator = validateKey

	var valueEntry *widget.Entry
	if env.IsSecret(key) {
		valueEntry = widget.NewPasswordEntry()
	} else {
		valueEntry = widget.NewEntry()
	}
	valueEntry.SetPlaceHolder("value")
	valueEntry.TextStyle = fyne.TextStyle{Monospace: true}
	valueEntry.SetText(value)

	secretHint := widget.NewLabel("Looks like a secret — the value will be masked in the table.")
	secretHint.Importance = widget.LowImportance
	secretHint.Wrapping = fyne.TextWrapWord
	secretHint.Hidden = !env.IsSecret(key)
	keyEntry.OnChanged = func(k string) {
		secretHint.Hidden = !env.IsSecret(k)
		secretHint.Refresh()
	}

	errLabel := widget.NewLabel("")
	errLabel.Importance = widget.DangerImportance
	errLabel.Wrapping = fyne.TextWrapWord
	errLabel.Hide()

	caption := func(s string) *widget.Label {
		l := widget.NewLabelWithStyle(s, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		l.Importance = widget.LowImportance
		return l
	}

	width := canvas.NewRectangle(color.Transparent)
	width.SetMinSize(fyne.NewSize(460, 0))

	copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		a.window.Clipboard().SetContent(valueEntry.Text)
		a.flashStatus("Copied " + strings.TrimSpace(keyEntry.Text) + " to clipboard")
	})
	copyBtn.Importance = widget.LowImportance

	form := container.NewVBox(
		width,
		caption("Name"),
		keyEntry,
		caption("Value"),
		container.NewBorder(nil, nil, nil, copyBtn, valueEntry),
		secretHint,
		errLabel,
	)

	var d *dialog.CustomDialog

	save := func() {
		k := strings.TrimSpace(keyEntry.Text)
		if err := validateKey(k); err != nil {
			errLabel.SetText(err.Error())
			errLabel.Show()
			a.window.Canvas().Focus(keyEntry)
			return
		}
		for i, v := range f.Variables {
			if v.Key == k && i != idx {
				errLabel.SetText(fmt.Sprintf("%s already exists in this file.", k))
				errLabel.Show()
				return
			}
		}

		if editing {
			v := &f.Variables[idx]
			if v.Key == k && v.Value == valueEntry.Text {
				d.Hide()
				return
			}
			v.Key = k
			v.Value = valueEntry.Text
			v.IsSecret = env.IsSecret(k)
		} else {
			f.Variables = append(f.Variables, env.Variable{
				Key:      k,
				Value:    valueEntry.Text,
				Source:   env.SourceFile,
				IsSecret: env.IsSecret(k),
				Line:     -1, // new variable, appended on save
			})
		}
		d.Hide()
		a.markModified(f)
	}
	keyEntry.OnSubmitted = func(string) { a.window.Canvas().Focus(valueEntry) }
	valueEntry.OnSubmitted = func(string) { save() }

	saveBtn := widget.NewButtonWithIcon("Save", theme.ConfirmIcon(), save)
	if !editing {
		saveBtn.SetText("Add")
		saveBtn.SetIcon(theme.ContentAddIcon())
	}
	saveBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButton("Cancel", func() { d.Hide() })

	buttons := []fyne.CanvasObject{cancelBtn, saveBtn}
	if editing {
		deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
			dialog.ShowConfirm("Delete variable",
				fmt.Sprintf("Delete %s from %s?", f.Variables[idx].Key, filepath.Base(f.Path)),
				func(ok bool) {
					if !ok {
						return
					}
					f.Variables = append(f.Variables[:idx], f.Variables[idx+1:]...)
					d.Hide()
					a.markModified(f)
				}, a.window)
		})
		deleteBtn.Importance = widget.DangerImportance
		buttons = []fyne.CanvasObject{deleteBtn, layout.NewSpacer(), cancelBtn, saveBtn}
	}

	title := "Add variable"
	if editing {
		title = "Edit variable"
	}
	d = dialog.NewCustomWithoutButtons(title, form, a.window)
	d.SetButtons(buttons)
	d.Show()

	if editing || key != "" {
		a.window.Canvas().Focus(valueEntry)
	} else {
		a.window.Canvas().Focus(keyEntry)
	}
}
