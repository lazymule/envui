package ui

import (
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mrf/envui/env"
)

// updateIssues re-validates the active file.
func (a *App) updateIssues() {
	a.issues = env.Validate(a.activeFile(), a.exampleFile())
	a.issueIdx = make(map[int]bool, len(a.issues))
	for _, is := range a.issues {
		a.issueIdx[is.Index] = true
	}
	if a.issuesBtn == nil {
		return
	}
	if len(a.issues) == 0 {
		a.issuesBtn.Hide()
		return
	}
	a.issuesBtn.SetText(Tn("issues.count", len(a.issues)))
	a.issuesBtn.Show()
}

func issueText(is env.Issue, template string) string {
	switch is.Kind {
	case env.IssueDuplicate:
		return T("issue.duplicate")
	case env.IssueSpaces:
		return T("issue.spaces")
	case env.IssueEmptyRequired:
		return T("issue.emptyRequired", template)
	}
	return ""
}

// showIssues lists the problems of the active file; clicking one opens the
// variable for editing.
func (a *App) showIssues() {
	f := a.activeFile()
	if f == nil || len(a.issues) == 0 {
		return
	}
	issues := append([]env.Issue(nil), a.issues...)
	tpl := ""
	if ex := a.exampleFile(); ex != nil {
		tpl = filepath.Base(ex.Path)
	}

	var d dialog.Dialog
	list := widget.NewList(
		func() int { return len(issues) },
		func() fyne.CanvasObject {
			key := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true})
			msg := widget.NewLabel("")
			msg.Importance = widget.LowImportance
			return container.NewBorder(nil, nil,
				container.NewHBox(widget.NewIcon(theme.NewWarningThemedResource(theme.WarningIcon())), key), nil, msg)
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			c := o.(*fyne.Container)
			msg := c.Objects[0].(*widget.Label)
			key := c.Objects[1].(*fyne.Container).Objects[1].(*widget.Label)
			key.SetText(issues[id].Key)
			msg.SetText(issueText(issues[id], tpl))
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		is := issues[id]
		if is.Index < len(f.Variables) {
			d.Hide()
			v := f.Variables[is.Index]
			a.showVariableDialog(f, is.Index, v.Key, v.Value)
		}
	}

	intro := widget.NewLabel(T("issues.intro", filepath.Base(f.Path)))
	intro.Wrapping = fyne.TextWrapWord
	size := canvas.NewRectangle(nil)
	size.SetMinSize(fyne.NewSize(560, 280))

	d = dialog.NewCustom(T("issues.title"), T("common.close"),
		container.NewBorder(intro, nil, nil, nil, container.NewStack(size, list)), a.window)
	d.Show()
}

// ---------------------------------------------------------------------------
// git safety banner
// ---------------------------------------------------------------------------

func (a *App) buildGitBanner() fyne.CanvasObject {
	a.gitLabel = widget.NewLabel("")
	a.gitLabel.Wrapping = fyne.TextWrapWord
	a.gitLabel.Importance = widget.WarningImportance

	fix := widget.NewButtonWithIcon(T("git.add"), theme.ContentAddIcon(), a.onAddToGitignore)
	fix.Importance = widget.WarningImportance
	dismiss := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		a.gitDismissed = true
		a.gitBanner.Hide()
	})
	dismiss.Importance = widget.LowImportance

	bg := canvas.NewRectangle(nil)
	a.onThemeChange = append(a.onThemeChange, func() { bg.FillColor = alpha(a.themeColor(theme.ColorNameWarning), 0x22); bg.Refresh() })
	bg.FillColor = alpha(a.themeColor(theme.ColorNameWarning), 0x22)
	bg.CornerRadius = 6

	a.gitBanner = container.NewStack(bg, container.NewPadded(container.NewBorder(nil, nil,
		widget.NewIcon(theme.NewWarningThemedResource(theme.WarningIcon())),
		container.NewHBox(fix, dismiss),
		a.gitLabel)))
	a.gitBanner.Hide()
	return a.gitBanner
}

// checkGit looks for env files that git would commit.
func (a *App) checkGit() {
	a.unignored = env.UnignoredEnvFiles(a.project)
	a.gitDismissed = false
	a.updateGitBanner()
}

func (a *App) updateGitBanner() {
	if a.gitBanner == nil {
		return
	}
	if len(a.unignored) == 0 || a.gitDismissed {
		a.gitBanner.Hide()
		return
	}
	var names []string
	for _, f := range a.unignored {
		names = append(names, filepath.Base(f.Path))
	}
	a.gitLabel.SetText(T("git.warning", strings.Join(names, ", ")))
	a.gitBanner.Show()
}

func (a *App) onAddToGitignore() {
	if a.project == nil || len(a.unignored) == 0 {
		return
	}
	var names []string
	for _, f := range a.unignored {
		rel, err := filepath.Rel(a.project.Path, f.Path)
		if err != nil {
			rel = filepath.Base(f.Path)
		}
		names = append(names, filepath.ToSlash(rel))
	}
	if err := env.AddToGitignore(a.project.Path, names); err != nil {
		a.showError(err)
		return
	}
	a.checkGit()
	if len(a.unignored) > 0 {
		// still not ignored (e.g. a "!" rule elsewhere): tell the user
		a.showInfo(T("git.title"), T("git.stillTracked"))
		return
	}
	a.flashStatus(T("git.added"))
}
