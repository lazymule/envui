package env

import (
	"strings"
	"testing"
)

func TestParseLineSyntax(t *testing.T) {
	tests := []struct {
		line    string
		key     string
		value   string
		export  bool
		comment string
	}{
		{`A=plain`, "A", "plain", false, ""},
		{`export B=1`, "B", "1", true, ""},
		{`  export   C = spaced `, "C", "spaced", true, ""},
		{`D="hello world"`, "D", "hello world", false, ""},
		{`E='single "quoted"'`, "E", `single "quoted"`, false, ""},
		{`F="say \"hi\" \\ done"`, "F", `say "hi" \ done`, false, ""},
		{`G="line1\nline2"`, "G", "line1\nline2", false, ""},
		{`H=8080 # http port`, "H", "8080", false, "# http port"},
		{`I="a # not comment" # real comment`, "I", "a # not comment", false, "# real comment"},
		{`J=http://x/#anchor`, "J", "http://x/#anchor", false, ""},
		{`K=`, "K", "", false, ""},
		{`L="unterminated`, "L", `"unterminated`, false, ""},
		{`URL=a=b=c`, "URL", "a=b=c", false, ""},
	}
	for _, tt := range tests {
		v, ok := parseLine(tt.line)
		if !ok {
			t.Errorf("parseLine(%q) failed", tt.line)
			continue
		}
		if v.Key != tt.key || v.Value != tt.value || v.Export != tt.export || v.InlineComment != tt.comment {
			t.Errorf("parseLine(%q) = key %q value %q export %v comment %q; want %q %q %v %q",
				tt.line, v.Key, v.Value, v.Export, v.InlineComment, tt.key, tt.value, tt.export, tt.comment)
		}
	}

	for _, bad := range []string{"no equals", "=value", "BAD KEY=x"} {
		if _, ok := parseLine(bad); ok {
			t.Errorf("parseLine(%q) should fail", bad)
		}
	}
}

func TestTrailingSpaceDetected(t *testing.T) {
	v, _ := parseLine("A=value   ")
	if !v.TrailingSpace || v.Value != "value" {
		t.Errorf("got value %q trailing %v", v.Value, v.TrailingSpace)
	}
	v, _ = parseLine("A=value   # comment")
	if v.TrailingSpace {
		t.Error("space before a comment is not trailing whitespace")
	}
}

// The original writer escaped quotes but the parser didn't unescape them, so
// every save added another backslash. Values must survive any number of
// save/load cycles.
func TestValuesSurviveRepeatedSaves(t *testing.T) {
	values := []string{
		`say "hi"`, `C:\path\to\file`, `it's`, "two\nlines", `#hash`, ` padded `,
		`a\"b`, `trailing\`, `=starts`, "tab\there", `ünicode çğş`, ``,
	}
	for _, want := range values {
		f := &EnvFile{Path: "x", Comments: map[int]string{}, Variables: []Variable{{Key: "K", Value: want, Line: -1}}}
		for round := 0; round < 3; round++ {
			f = ParseEnv("x", FormatEnvFile(f))
			if len(f.Variables) != 1 || f.Variables[0].Value != want {
				t.Fatalf("value %q after round %d: %+v", want, round, f.Variables)
			}
			f.Variables[0].Value += "" // keep
			f.Variables[0].raw = ""    // force re-formatting next round
		}
	}
}

func TestUnchangedLinesAreWrittenVerbatim(t *testing.T) {
	in := strings.Join([]string{
		"# header",
		"export   A = 'x'   # keep my spacing",
		"weird line without equals",
		"B=\"q\"",
		"",
		"C=3",
	}, "\r\n") + "\r\n"
	f := ParseEnv(".env", in)
	if got := FormatEnvFile(f); got != in {
		t.Fatalf("unchanged file was rewritten:\n%q\nwant\n%q", got, in)
	}

	// change only C: everything else (incl. CRLF and the odd line) stays
	f.Variables[2].Value = "4"
	out := FormatEnvFile(f)
	want := strings.Replace(in, "C=3", "C=4", 1)
	if out != want {
		t.Fatalf("got\n%q\nwant\n%q", out, want)
	}
}

func TestChangedVariableKeepsStyle(t *testing.T) {
	f := ParseEnv(".env", "export A='old' # note\n")
	f.Variables[0].Value = "new value"
	if got := FormatEnvFile(f); got != "export A='new value' # note\n" {
		t.Errorf("got %q", got)
	}
}

func TestMarkEditedDropsTrailingSpace(t *testing.T) {
	f := ParseEnv(".env", "A=1   \n")
	if issues := Validate(f, nil); len(issues) != 1 {
		t.Fatalf("expected trailing-space issue, got %v", issues)
	}
	f.Variables[0].MarkEdited()
	if got := FormatEnvFile(f); got != "A=1\n" {
		t.Errorf("got %q", got)
	}
	if issues := Validate(f, nil); len(issues) != 0 {
		t.Errorf("issue should be gone: %v", issues)
	}
}
