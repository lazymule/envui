package ui

import (
	"reflect"

	"github.com/mrf/envui/env"
)

// fileState is a copy of one file's variables at some point in time.
type fileState struct {
	file *env.EnvFile
	vars []env.Variable
}

// historyEntry is one undoable action; it may touch several files (e.g.
// copying a variable into other files).
type historyEntry struct {
	states []fileState
}

const maxHistory = 200

func copyVars(vs []env.Variable) []env.Variable {
	out := make([]env.Variable, len(vs))
	copy(out, vs)
	return out
}

func sameVars(a, b []env.Variable) bool {
	if len(a) != len(b) {
		return false
	}
	return reflect.DeepEqual(a, b)
}

func capture(files []*env.EnvFile) historyEntry {
	var e historyEntry
	for _, f := range files {
		e.states = append(e.states, fileState{file: f, vars: copyVars(f.Variables)})
	}
	return e
}

// resetHistory forgets undo/redo and records the current contents of every
// file as the saved state (after loading or reloading a project).
func (a *App) resetHistory() {
	a.undoStack, a.redoStack = nil, nil
	a.savedVars = make(map[*env.EnvFile][]env.Variable)
	if a.project != nil {
		for _, f := range a.project.Files {
			a.savedVars[f] = copyVars(f.Variables)
		}
	}
}

// markSaved records f's current contents as what is on disk.
func (a *App) markSaved(f *env.EnvFile) {
	if a.savedVars == nil {
		a.savedVars = make(map[*env.EnvFile][]env.Variable)
	}
	a.savedVars[f] = copyVars(f.Variables)
	f.Modified = false
}

func (a *App) updateModified(f *env.EnvFile) {
	saved, ok := a.savedVars[f]
	f.Modified = !ok || !sameVars(saved, f.Variables)
}

// mutate runs change, which may modify the variables of the given files,
// and records it as one undoable step.
func (a *App) mutate(files []*env.EnvFile, change func()) {
	before := capture(files)
	change()

	changed := false
	for _, s := range before.states {
		if !sameVars(s.vars, s.file.Variables) {
			changed = true
		}
		a.updateModified(s.file)
	}
	if changed {
		a.undoStack = append(a.undoStack, before)
		if len(a.undoStack) > maxHistory {
			a.undoStack = a.undoStack[1:]
		}
		a.redoStack = nil
	}
	a.refresh()
}

func (a *App) onUndo() { a.stepHistory(&a.undoStack, &a.redoStack, "status.undone") }
func (a *App) onRedo() { a.stepHistory(&a.redoStack, &a.undoStack, "status.redone") }

func (a *App) stepHistory(from, to *[]historyEntry, msgKey string) {
	if len(*from) == 0 {
		return
	}
	e := (*from)[len(*from)-1]
	*from = (*from)[:len(*from)-1]

	var files []*env.EnvFile
	for _, s := range e.states {
		files = append(files, s.file)
	}
	*to = append(*to, capture(files))

	for _, s := range e.states {
		s.file.Variables = copyVars(s.vars)
		a.updateModified(s.file)
	}
	// show the file the step happened in
	if len(files) > 0 && a.project != nil && a.project.ActiveFile != files[0] {
		a.selectFile(files[0].Path)
	}
	a.refresh()
	a.flashStatus(T(msgKey))
}
