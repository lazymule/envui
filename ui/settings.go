package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// showSettings opens the settings dialog (language, theme, startup options).
func (a *App) showSettings() {
	prefs := a.prefs

	names := make([]string, len(languages))
	selected := ""
	for i, l := range languages {
		names[i] = l.Name
		if l.Code == CurrentLanguage() {
			selected = l.Name
		}
	}
	// a radio group rather than a Select: Fyne 2.4's Select truncates
	// non-ASCII labels such as "Türkçe" incorrectly
	langSelect := widget.NewRadioGroup(names, nil)
	langSelect.Horizontal = true
	langSelect.Required = true
	langSelect.SetSelected(selected)

	light, dark := T("settings.light"), T("settings.dark")
	themeRadio := widget.NewRadioGroup([]string{light, dark}, nil)
	themeRadio.Horizontal = true
	themeRadio.Required = true
	if a.dark {
		themeRadio.SetSelected(dark)
	} else {
		themeRadio.SetSelected(light)
	}

	shellCheck := widget.NewCheck(T("settings.shell"), nil)
	shellCheck.SetChecked(prefs.Bool(prefShowShell))
	secretsCheck := widget.NewCheck(T("settings.secrets"), nil)
	secretsCheck.SetChecked(prefs.Bool(prefShowSecrets))

	updatesCheck := widget.NewCheck(T("settings.updates"), nil)
	updatesCheck.SetChecked(prefs.BoolWithFallback(prefCheckUpdates, true))

	note := widget.NewLabel(T("settings.note"))
	note.Importance = widget.LowImportance
	note.Wrapping = fyne.TextWrapWord

	width := canvas.NewRectangle(color.Transparent)
	width.SetMinSize(fyne.NewSize(420, 0))

	form := widget.NewForm(
		widget.NewFormItem(T("settings.language"), langSelect),
		widget.NewFormItem(T("settings.theme"), themeRadio),
		widget.NewFormItem(T("settings.startup"), container.NewVBox(shellCheck, secretsCheck, updatesCheck)),
	)

	var d *dialog.CustomDialog
	save := func() {
		d.Hide()

		prefs.SetBool(prefShowShell, shellCheck.Checked)
		prefs.SetBool(prefShowSecrets, secretsCheck.Checked)
		prefs.SetBool(prefCheckUpdates, updatesCheck.Checked)
		a.setDark(themeRadio.Selected == dark)

		newLang := CurrentLanguage()
		for _, l := range languages {
			if l.Name == langSelect.Selected {
				newLang = l.Code
			}
		}
		if newLang != CurrentLanguage() {
			SetLanguage(newLang)
			prefs.SetString(prefLanguage, string(newLang))
			a.rebuildUI()
		}
	}

	saveBtn := widget.NewButtonWithIcon(T("common.save"), theme.ConfirmIcon(), save)
	saveBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButton(T("common.cancel"), func() { d.Hide() })

	d = dialog.NewCustomWithoutButtons(T("settings.title"),
		container.NewVBox(width, form, widget.NewSeparator(), note), a.window)
	d.SetButtons([]fyne.CanvasObject{cancelBtn, saveBtn})
	d.Show()
}

// rebuildUI recreates the whole interface (e.g. after a language change)
// while keeping the open project, the selected file and unsaved edits.
func (a *App) rebuildUI() {
	active := ""
	if f := a.activeFile(); f != nil {
		active = f.Path
	}
	a.window.SetTitle(T("app.title"))
	a.window.SetContent(a.BuildUI())
	if a.project == nil {
		return
	}
	a.showView()
	a.refresh()
	if active != "" {
		a.selectFile(active)
	}
}
