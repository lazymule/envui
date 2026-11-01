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

	// raw is the original line and parsed* the values read from it. If the
	// variable is unchanged when saving, raw is written back verbatim so
	// untouched lines never change.
	raw          string
	parsedKey    string
	parsedValue  string
	parsedExport bool
}

// MarkEdited forces the variable to be re-formatted when saving, e.g. to
// drop trailing whitespace even though the value itself didn't change.
func (v *Variable) MarkEdited() {
	v.raw = ""
	v.TrailingSpace = false
}

// Unchanged reports whether the variable still has the key, value and
// export flag it was read with.
func (v *Variable) Unchanged() bool {
	return v.raw != "" && v.Key == v.parsedKey && v.Value == v.parsedValue && v.Export == v.parsedExport
}

// EnvFile represents a parsed .env file
type EnvFile struct {
	Path      string
	Variables []Variable
	Comments  map[int]string // Preserve comments by line
	Modified  bool           // Track if file has unsaved changes
	CRLF      bool           // file uses Windows (\r\n) line endings
}

// Project represents the current working directory
type Project struct {
	Path       string
	Files      []*EnvFile
	ActiveFile *EnvFile
	ShellVars  []Variable
}
