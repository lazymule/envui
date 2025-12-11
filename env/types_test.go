package env

import "testing"

func TestSourceString(t *testing.T) {
	tests := []struct {
		source   Source
		expected string
	}{
		{SourceFile, "file"},
		{SourceShell, "shell"},
		{SourceMissing, "missing"},
		{Source(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.source.String()
			if got != tt.expected {
				t.Errorf("Source(%d).String() = %q, want %q", tt.source, got, tt.expected)
			}
		})
	}
}
