package env

import (
	"os"
	"strings"
	"testing"
)

func loadDocumented(t *testing.T) (*EnvFile, string) {
	t.Helper()
	data, err := os.ReadFile("testdata/documented.env")
	if err != nil {
		t.Fatal(err)
	}
	return ParseEnv(".env.example", string(data)), string(data)
}

func varByKey(f *EnvFile, key string) *Variable {
	for i := range f.Variables {
		if f.Variables[i].Key == key {
			return &f.Variables[i]
		}
	}
	return nil
}

func TestDocsAndSections(t *testing.T) {
	f, original := loadDocumented(t)
	if FormatEnvFile(f) != original {
		t.Fatal("unchanged documented file must round-trip exactly")
	}

	cases := []struct{ key, section, doc, group string }{
		{"dmvc.server.port", "Server", "", ""},
		{"dmvc.swagger.host", "Server", "Address Swagger sends requests to", ""},
		{"db.server", "Database (SQL Server login; Windows authentication is not used)", `Server: name or name\instance (e.g. .\SQL17)`, ""},
		{"db.database", "Database (SQL Server login; Windows authentication is not used)", "", `Server: name or name\instance (e.g. .\SQL17)`},
		{"db.username", "Database (SQL Server login; Windows authentication is not used)",
			"User and password are required. If the password contains #, $ or quotes,\nwrite it in single quotes: 'p#ro'", ""},
		{"db.password", "Database (SQL Server login; Windows authentication is not used)", "",
			"User and password are required. If the password contains #, $ or quotes,\nwrite it in single quotes: 'p#ro'"},
		{"logger.level", "Logging", "", ""},
	}
	for _, c := range cases {
		v := varByKey(f, c.key)
		if v == nil {
			t.Fatalf("%s missing", c.key)
		}
		if v.Section != c.section || v.Doc != c.doc || v.GroupDoc != c.group {
			t.Errorf("%s:\n section %q\n doc     %q\n group   %q", c.key, v.Section, v.Doc, v.GroupDoc)
		}
	}
	if v := varByKey(f, "logger.level"); v.InlineComment != "# debug, info, warning" {
		t.Errorf("inline comment %q", v.InlineComment)
	}
	if v := varByKey(f, "dmvc.analytics.folder"); v.Value != `logs\analytics` {
		t.Errorf("backslash value %q", v.Value)
	}
}

func TestEditDoc(t *testing.T) {
	f, _ := loadDocumented(t)
	v := varByKey(f, "dmvc.swagger.host")
	v.Doc = "Swagger target\n(host:port)"
	out := FormatEnvFile(f)
	want := "# Swagger target\n# (host:port)\ndmvc.swagger.host=localhost:8085\n"
	if !strings.Contains(out, want) || strings.Contains(out, "Address Swagger") {
		t.Errorf("doc not replaced:\n%s", out)
	}

	// removing a description removes its comment lines
	f, _ = loadDocumented(t)
	varByKey(f, "db.username").Doc = ""
	out = FormatEnvFile(f)
	if strings.Contains(out, "User and password") || !strings.Contains(out, "\ndb.username=\ndb.password=\n") {
		t.Errorf("doc not removed:\n%s", out)
	}

	// a new variable gets its description above it
	f, _ = loadDocumented(t)
	f.Variables = append(f.Variables, Variable{Key: "NEW", Value: "1", Doc: "a new one", Line: -1})
	if out := FormatEnvFile(f); !strings.HasSuffix(out, "# a new one\nNEW=1\n") {
		t.Errorf("new var doc:\n%s", out)
	}
}

func TestDeleteRemovesLineAndDoc(t *testing.T) {
	f, original := loadDocumented(t)
	var kept []Variable
	for _, v := range f.Variables {
		if v.Key != "dmvc.swagger.host" {
			kept = append(kept, v)
		}
	}
	f.Variables = kept
	out := FormatEnvFile(f)
	want := strings.Replace(original, "# Address Swagger sends requests to\ndmvc.swagger.host=localhost:8085\n", "", 1)
	if out != want {
		t.Errorf("got:\n%s\nwant:\n%s", out, want)
	}
}

func TestQuotingStyle(t *testing.T) {
	cases := map[string]string{
		`logs\analytics`:   `logs\analytics`,     // backslashes alone: unquoted
		`p#ro$x`:           `'p#ro$x'`,           // # and $: single quotes (literal everywhere)
		`hello world`:      `"hello world"`,      // only spaces: double quotes
		`C:\Program Files`: `'C:\Program Files'`, // spaces + backslash: single
		`it's`:             `"it's"`,             // apostrophe: double
		`it's "q"`:         `"it's \"q\""`,       // both quotes: escaped double
		"two\nlines":       `"two\nlines"`,
		`plain`:            `plain`,
	}
	for in, want := range cases {
		if got := quoteValue(in, 0); got != want {
			t.Errorf("quoteValue(%q) = %s, want %s", in, got, want)
		}
		f := ParseEnv("x", "K="+quoteValue(in, 0)+"\n")
		if f.Variables[0].Value != in {
			t.Errorf("round trip of %q gave %q", in, f.Variables[0].Value)
		}
	}
}
