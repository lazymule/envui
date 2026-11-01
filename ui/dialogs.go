package ui

import (
	"errors"
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
		return errors.New(T("key.required"))
	}
	if !validKey.MatchString(key) {
		return errors.New(T("key.invalid"))
	}
	return nil
}

// onAddVariable handles adding a new variable
func (a *App) onAddVariable() {
	f := a.activeFile()
	if f == nil {
		a.showInfo(T("var.noFileTitle"), T("var.noFileText"))
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
	valueEntry.SetPlaceHolder(T("var.valueHint"))
	valueEntry.TextStyle = fyne.TextStyle{Monospace: true}
	valueEntry.SetText(value)

	secretHint := widget.NewLabel(T("var.secretHint"))
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
		a.flashStatus(T("status.copied", strings.TrimSpace(keyEntry.Text)))
	})
	copyBtn.Importance = widget.LowImportance

	form := container.NewVBox(
		width,
		caption(T("var.name")),
		keyEntry,
		caption(T("var.value")),
		container.NewBorder(nil, nil, nil, copyBtn, valueEntry),
		secretHint,
		errLabel,
	)
	if editing && a.project != nil && len(a.project.Files) > 1 {
		copyTo := widget.NewButtonWithIcon(T("copyTo.button"), theme.MailForwardIcon(), func() {
			a.showCopyToFiles(f, strings.TrimSpace(keyEntry.Text), valueEntry.Text)
		})
		copyTo.Importance = widget.LowImportance
		copyTo.Alignment = widget.ButtonAlignLeading
		form.Add(copyTo)
	}

	var d *dialog.CustomDialog

	save := func() {
		k := strings.TrimSpace(keyEntry.Text)
		if err := validateKey(k); err != nil {
			errLabel.SetText(err.Error())
			errLabel.Show()
			a.window.Canvas().Focus(keyEntry)
			return
		}
		// the name must be unique, unless it is unchanged (so existing
		// duplicates can still be edited)
		for i, v := range f.Variables {
			if v.Key == k && i != idx && k != key {
				errLabel.SetText(T("var.exists", k))
				errLabel.Show()
				return
			}
		}

		d.Hide()
		a.mutate([]*env.EnvFile{f}, func() {
			if editing {
				v := &f.Variables[idx]
				if v.TrailingSpace {
					v.MarkEdited() // saving cleans up the stray spaces
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
		})
	}
	keyEntry.OnSubmitted = func(string) { a.window.Canvas().Focus(valueEntry) }
	valueEntry.OnSubmitted = func(string) { save() }

	saveBtn := widget.NewButtonWithIcon(T("common.save"), theme.ConfirmIcon(), save)
	if !editing {
		saveBtn.SetText(T("common.add"))
		saveBtn.SetIcon(theme.ContentAddIcon())
	}
	saveBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButton(T("common.cancel"), func() { d.Hide() })

	buttons := []fyne.CanvasObject{cancelBtn, saveBtn}
	if editing {
		deleteBtn := widget.NewButtonWithIcon(T("common.delete"), theme.DeleteIcon(), func() {
			a.showConfirm(T("var.deleteTitle"),
				T("var.deleteAsk", f.Variables[idx].Key, filepath.Base(f.Path)),
				T("common.delete"),
				func() {
					d.Hide()
					a.mutate([]*env.EnvFile{f}, func() {
						vars := copyVars(f.Variables)
						f.Variables = append(vars[:idx], vars[idx+1:]...)
					})
				})
		})
		deleteBtn.Importance = widget.DangerImportance
		buttons = []fyne.CanvasObject{deleteBtn, layout.NewSpacer(), cancelBtn, saveBtn}
	}

	title := T("var.addTitle")
	if editing {
		title = T("var.editTitle")
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

// showError shows an error dialog with translated title and button.
func (a *App) showError(err error) {
	msg := widget.NewLabel(err.Error())
	msg.Wrapping = fyne.TextWrapWord
	width := canvas.NewRectangle(color.Transparent)
	width.SetMinSize(fyne.NewSize(380, 0))
	d := dialog.NewCustom(T("common.error"), T("common.ok"),
		container.NewBorder(nil, nil, widget.NewIcon(theme.ErrorIcon()), nil,
			container.NewVBox(width, msg)), a.window)
	d.Show()
}

// showInfo shows an information dialog with a translated button.
func (a *App) showInfo(title, message string) {
	dialog.NewCustom(title, T("common.ok"), widget.NewLabel(message), a.window).Show()
}

// showConfirm asks for confirmation; onConfirm runs only if the user agrees.
func (a *App) showConfirm(title, message, confirmText string, onConfirm func()) {
	d := dialog.NewConfirm(title, message, func(ok bool) {
		if ok {
			onConfirm()
		}
	}, a.window)
	d.SetConfirmText(confirmText)
	d.SetDismissText(T("common.cancel"))
	d.Show()
}
