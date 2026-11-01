package env

import "strings"

// IssueKind is the type of a problem found by Validate.
type IssueKind int

const (
	// IssueDuplicate: the key is defined more than once in the file.
	IssueDuplicate IssueKind = iota
	// IssueSpaces: the value starts or ends with whitespace (often a
	// copy/paste accident; unquoted trailing spaces are silently dropped by
	// most dotenv loaders).
	IssueSpaces
	// IssueEmptyRequired: the value is empty but the template has a value.
	IssueEmptyRequired
)

// Issue is a problem with one variable of a file.
type Issue struct {
	Kind  IssueKind
	Key   string
	Index int // index into EnvFile.Variables
}

// Validate checks a file for common mistakes. template may be nil.
func Validate(f *EnvFile, template *EnvFile) []Issue {
	if f == nil {
		return nil
	}
	var issues []Issue

	seen := make(map[string]bool)
	for i, v := range f.Variables {
		if v.Disabled {
			continue // commented out: not read by loaders
		}
		if seen[v.Key] {
			issues = append(issues, Issue{IssueDuplicate, v.Key, i})
		}
		seen[v.Key] = true

		if (v.TrailingSpace && v.Unchanged()) || (v.Value != "" && v.Value != strings.TrimSpace(v.Value)) {
			issues = append(issues, Issue{IssueSpaces, v.Key, i})
		}
	}

	if template != nil && template != f {
		example := make(map[string]string, len(template.Variables))
		for _, v := range template.Variables {
			if !v.Disabled {
				example[v.Key] = v.Value
			}
		}
		for i, v := range f.Variables {
			if ex, ok := example[v.Key]; ok && !v.Disabled && v.Value == "" && ex != "" {
				issues = append(issues, Issue{IssueEmptyRequired, v.Key, i})
			}
		}
	}
	return issues
}
