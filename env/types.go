package env

// Source indicates where a variable originated from
type Source int

const (
	SourceFile Source = iota
	SourceShell
	SourceMissing
)

func (s Source) String() string {
	switch s {
	case SourceFile:
		return "file"
	case SourceShell:
		return "shell"
	case SourceMissing:
		return "missing"
	default:
		return "unknown"
	}
}

// Variable represents a single environment variable
type Variable struct {
	Key      string
	Value    string
	Source   Source
	IsSecret bool
	Line     int // Original line number (for preservation)

	Export        bool   // line started with "export "
	Quote         byte   // quote character used in the file: 0, '"' or '\''
	InlineComment string // trailing "# ..." comment on the same line
	TrailingSpace bool   // unquoted value had trailing whitespace in the file

	// Disabled marks a commented-out variable ("# KEY=value"). It is kept
	// in the file as a comment and ignored by loaders until enabled.
	Disabled bool

	// Doc is the description written in "#" comment lines directly above
	// the variable (without the "#"), one line per comment line.
	Doc string
	// Section is the title of the "# ==== / # Title / # ====" block the
	// variable is in ("" if the file has no such blocks).
	Section string
	// GroupDoc is the Doc of the variable directly above when this one has
	// none of its own and sits in the same block (no blank line between),
	// e.g. a comment above db.username that also covers db.password.
	GroupDoc string

	// raw is the original line and parsed* the values read from it. If the
	// variable is unchanged when saving, raw is written back verbatim so
	// untouched lines never change.
	raw          string
	parsedKey    string
	parsedValue  string
	parsedExport bool
	parsedOff    bool
	parsedDoc    string
	docLines     []int // line numbers of the Doc comment lines
}

// DocChanged reports whether the description differs from the file.
func (v *Variable) DocChanged() bool { return v.Doc != v.parsedDoc }

// MarkEdited forces the variable to be re-formatted when saving, e.g. to
// drop trailing whitespace even though the value itself didn't change.
func (v *Variable) MarkEdited() {
	v.raw = ""
	v.TrailingSpace = false
}

// Unchanged reports whether the variable still has the key, value and
// export flag it was read with.
func (v *Variable) Unchanged() bool {
	return v.raw != "" && v.Key == v.parsedKey && v.Value == v.parsedValue && v.Export == v.parsedExport &&
		v.Disabled == v.parsedOff
}

// Active returns the variables that are not commented out.
func (f *EnvFile) Active() []Variable {
	out := make([]Variable, 0, len(f.Variables))
	for _, v := range f.Variables {
		if !v.Disabled {
			out = append(out, v)
		}
	}
	return out
}

// EnvFile represents a parsed .env file
type EnvFile struct {
	Path      string
	Variables []Variable
	Comments  map[int]string // Preserve comments by line
	Modified  bool           // Track if file has unsaved changes
	CRLF      bool           // file uses Windows (\r\n) line endings

	// set by the parser, never modified afterwards (so undo snapshots of
	// Variables stay consistent with them)
	varLines map[int]bool // lines that held a variable when parsed
	docOwner map[int]int  // doc comment line -> line of the variable it describes
}

// Project represents the current working directory
type Project struct {
	Path       string
	Files      []*EnvFile
	ActiveFile *EnvFile
	ShellVars  []Variable
}
