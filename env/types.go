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
}

// EnvFile represents a parsed .env file
type EnvFile struct {
	Path      string
	Variables []Variable
	Comments  map[int]string // Preserve comments by line
	Modified  bool           // Track if file has unsaved changes
}

// Project represents the current working directory
type Project struct {
	Path       string
	Files      []*EnvFile
	ActiveFile *EnvFile
	ShellVars  []Variable
}
