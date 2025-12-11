# envui - Environment Variable Manager

A visual .env file editor built with Fyne and Go.

## Language & Framework

**Go + Fyne** - Native-feeling cross-platform GUI, familiar Go ecosystem.

## Why This Exists

Managing .env files sucks:
- Which vars are missing vs .env.example?
- What's inherited from shell vs explicit?
- Did I accidentally commit a secret?
- What's different between local and staging?

## User Experience

```
┌─────────────────────────────────────────────────────────────────┐
│ envui                                              [−] [□] [×]  │
├─────────────────────────────────────────────────────────────────┤
│ 📁 ~/Projects/hayden-cloud                          [Open...] │
├────────────────┬────────────────────────────────────────────────┤
│ Files          │ Variable           │ Value           │ Source │
│ ───────────────│────────────────────────────────────────────────│
│ ● .env         │ DATABASE_URL       │ postgres://lo...│ .env   │
│   .env.example │ API_KEY            │ ••••••••••••••• │ .env   │
│   .env.staging │ REDIS_URL          │ redis://localho │ .env   │
│   .env.prod    │ DEBUG              │ true            │ .env   │
│                │ ─────────── shell ─────────────────────────── │
│                │ HOME               │ /Users/mark     │ shell  │
│                │ PATH               │ /usr/local/bin: │ shell  │
│                │ ─────────── missing ────────────────────────── │
│                │ SENTRY_DSN         │ (not set)       │ ⚠️     │
│                │ STRIPE_KEY         │ (not set)       │ ⚠️     │
├────────────────┴────────────────────────────────────────────────┤
│ [+ Add Variable]  [Compare Files...]  [Show Secrets]  [Save]   │
└─────────────────────────────────────────────────────────────────┘
```

## Core Features

### 1. Visual Editor
- Table view of all variables
- Click to edit inline
- Add/delete variables
- Drag to reorder

### 2. Secret Masking
- Auto-detect secret patterns (API_KEY, PASSWORD, TOKEN, SECRET, etc.)
- Show as `•••••••••` by default
- "Show Secrets" toggle reveals temporarily
- Copy still works when masked

### 3. Source Tracking
- **File**: Defined in .env file
- **Shell**: Inherited from current shell environment
- **Missing**: In .env.example but not defined

### 4. File Comparison
- Compare .env vs .env.example
- Compare .env.local vs .env.staging
- Highlight differences (added/removed/changed)

### 5. Validation
- Warn on common issues:
  - Quotes inside values
  - Trailing whitespace
  - Duplicate keys
  - Empty values for required vars

## Components

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│  Fyne UI        │────▶│  Go Backend     │────▶│  File System    │
│  (widgets)      │     │  (parsing)      │     │  (.env files)   │
│                 │◀────│                 │◀────│                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │  os.Environ()   │
                        │  (shell vars)   │
                        └─────────────────┘
```

## Files

```
envui/
├── main.go              # App entry, window setup
├── ui/
│   ├── app.go           # Main application widget
│   ├── table.go         # Variable table widget
│   ├── editor.go        # Inline edit dialog
│   └── compare.go       # File comparison view
├── env/
│   ├── parser.go        # .env file parsing
│   ├── writer.go        # .env file writing
│   ├── shell.go         # Shell environment reading
│   └── secrets.go       # Secret detection patterns
├── go.mod
├── go.sum
└── ARCHITECTURE.md
```

## Dependencies

```go
module github.com/mrf/envui

go 1.21

require (
    fyne.io/fyne/v2 v2.4.0
    github.com/joho/godotenv v1.5.1  // .env parsing
)
```

## Data Structures

```go
// Variable represents a single env var
type Variable struct {
    Key      string
    Value    string
    Source   Source    // File, Shell, Missing
    IsSecret bool
    Line     int       // Original line number (for preservation)
}

type Source int

const (
    SourceFile Source = iota
    SourceShell
    SourceMissing
)

// EnvFile represents a parsed .env file
type EnvFile struct {
    Path      string
    Variables []Variable
    Comments  map[int]string  // Preserve comments by line
}

// Project represents the current working directory
type Project struct {
    Path        string
    Files       []EnvFile     // .env, .env.example, .env.local, etc.
    ActiveFile  *EnvFile      // Currently selected
    ShellVars   []Variable    // From os.Environ()
}
```

## Secret Detection

```go
var secretPatterns = []string{
    "PASSWORD",
    "SECRET",
    "TOKEN",
    "API_KEY",
    "APIKEY",
    "API_SECRET",
    "PRIVATE_KEY",
    "ACCESS_KEY",
    "AUTH",
    "CREDENTIAL",
    "DSN",  // Database connection strings often have passwords
}

func isSecret(key string) bool {
    upper := strings.ToUpper(key)
    for _, pattern := range secretPatterns {
        if strings.Contains(upper, pattern) {
            return true
        }
    }
    return false
}
```

## .env Parsing

Use `godotenv` for parsing but handle edge cases:

```go
func ParseEnvFile(path string) (*EnvFile, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    envFile := &EnvFile{
        Path:     path,
        Comments: make(map[int]string),
    }

    lines := strings.Split(string(content), "\n")
    for i, line := range lines {
        line = strings.TrimSpace(line)

        // Preserve comments
        if strings.HasPrefix(line, "#") {
            envFile.Comments[i] = line
            continue
        }

        // Skip empty lines
        if line == "" {
            continue
        }

        // Parse KEY=value
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        envFile.Variables = append(envFile.Variables, Variable{
            Key:      strings.TrimSpace(parts[0]),
            Value:    unquote(strings.TrimSpace(parts[1])),
            Source:   SourceFile,
            IsSecret: isSecret(parts[0]),
            Line:     i,
        })
    }

    return envFile, nil
}
```

## File Discovery

```go
func discoverEnvFiles(projectPath string) []string {
    patterns := []string{
        ".env",
        ".env.local",
        ".env.development",
        ".env.staging",
        ".env.production",
        ".env.example",
        ".env.sample",
        ".env.template",
    }

    var found []string
    for _, pattern := range patterns {
        path := filepath.Join(projectPath, pattern)
        if _, err := os.Stat(path); err == nil {
            found = append(found, path)
        }
    }
    return found
}
```

## Fyne UI Layout

```go
func (a *App) buildUI() fyne.CanvasObject {
    // Left: File list
    fileList := widget.NewList(
        func() int { return len(a.project.Files) },
        func() fyne.CanvasObject { return widget.NewLabel("") },
        func(id widget.ListItemID, obj fyne.CanvasObject) {
            obj.(*widget.Label).SetText(filepath.Base(a.project.Files[id].Path))
        },
    )

    // Right: Variable table
    varTable := widget.NewTable(
        func() (int, int) { return len(a.activeVars()), 3 },
        func() fyne.CanvasObject { return widget.NewLabel("") },
        func(id widget.TableCellID, obj fyne.CanvasObject) {
            v := a.activeVars()[id.Row]
            switch id.Col {
            case 0:
                obj.(*widget.Label).SetText(v.Key)
            case 1:
                if v.IsSecret && !a.showSecrets {
                    obj.(*widget.Label).SetText("••••••••••••")
                } else {
                    obj.(*widget.Label).SetText(truncate(v.Value, 30))
                }
            case 2:
                obj.(*widget.Label).SetText(v.Source.String())
            }
        },
    )

    // Bottom: Action buttons
    actions := container.NewHBox(
        widget.NewButton("+ Add Variable", a.onAddVariable),
        widget.NewButton("Compare Files...", a.onCompare),
        widget.NewCheck("Show Secrets", func(show bool) { a.showSecrets = show }),
        layout.NewSpacer(),
        widget.NewButton("Save", a.onSave),
    )

    // Layout
    split := container.NewHSplit(
        fileList,
        varTable,
    )
    split.SetOffset(0.2)

    return container.NewBorder(
        nil,      // top
        actions,  // bottom
        nil,      // left
        nil,      // right
        split,    // center
    )
}
```

## Compare View

```go
type CompareView struct {
    left  *EnvFile
    right *EnvFile
}

type DiffResult struct {
    Key    string
    Left   string  // empty if missing
    Right  string  // empty if missing
    Status DiffStatus
}

type DiffStatus int

const (
    DiffSame DiffStatus = iota
    DiffChanged
    DiffLeftOnly
    DiffRightOnly
)

func Compare(left, right *EnvFile) []DiffResult {
    // Build map of right file
    rightMap := make(map[string]string)
    for _, v := range right.Variables {
        rightMap[v.Key] = v.Value
    }

    var results []DiffResult
    seen := make(map[string]bool)

    // Check all left vars
    for _, v := range left.Variables {
        seen[v.Key] = true
        rightVal, exists := rightMap[v.Key]

        if !exists {
            results = append(results, DiffResult{
                Key: v.Key, Left: v.Value, Status: DiffLeftOnly,
            })
        } else if v.Value != rightVal {
            results = append(results, DiffResult{
                Key: v.Key, Left: v.Value, Right: rightVal, Status: DiffChanged,
            })
        }
    }

    // Check for right-only vars
    for _, v := range right.Variables {
        if !seen[v.Key] {
            results = append(results, DiffResult{
                Key: v.Key, Right: v.Value, Status: DiffRightOnly,
            })
        }
    }

    return results
}
```

## MVP Scope (2 hours)

1. ✅ Open directory, discover .env files
2. ✅ Parse and display variables in table
3. ✅ Show source (file vs shell)
4. ✅ Mask secrets by default
5. ✅ Edit variable values inline
6. ✅ Save changes back to file
7. ⏭️ Compare view (v2)
8. ⏭️ Missing variable detection (v2)

## Building

```bash
# Development
go run .

# Build binary
go build -o envui

# macOS app bundle (optional)
fyne package -os darwin -icon icon.png
```

## Future Ideas

- [ ] Compare view with diff highlighting
- [ ] Detect missing vars from .env.example
- [ ] Export to different formats (JSON, YAML, shell export)
- [ ] Variable groups/sections
- [ ] Search/filter variables
- [ ] Duplicate to new file (.env → .env.staging)
- [ ] Git integration (show uncommitted changes)
- [ ] Encrypt secrets at rest

## Why Fyne Over Electron?

| Fyne | Electron |
|------|----------|
| ~10MB binary | ~150MB+ |
| Native Go | Node.js |
| Simpler | More flexible |
| Less polished | More polished |

For a developer utility, Fyne's simplicity wins. Save Electron/Flutter for user-facing apps.
