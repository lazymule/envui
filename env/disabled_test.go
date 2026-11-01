package env

import (
	"strings"
	"testing"
)

func TestDisabledVariables(t *testing.T) {
	in := strings.Join([]string{
		"# only in DEBUG builds:",
		"# dev.lisans.web=5",
		"#dev.lisans.mobil=3",
		"# export TOKEN='a b'",
		"# enabled=false: jobs wait in the queue",
		"# Mode: Bearer Token. Login: POST /auth/login",
		"# see kuyruk.enabled=false below",
		"# ============",
		"PORT=8080",
		"# PORT=9090",
		"",
	}, "\n")
	f := ParseEnv(".env", in)
	if got := FormatEnvFile(f); got != in {
		t.Fatalf("round trip changed the file:\n%s", got)
	}

	var keys []string
	for _, v := range f.Variables {
		if v.Disabled {
			keys = append(keys, v.Key+"="+v.Value)
		}
	}
	want := "dev.lisans.web=5 dev.lisans.mobil=3 TOKEN=a b PORT=9090"
	if strings.Join(keys, " ") != want {
		t.Errorf("disabled = %q, want %q", strings.Join(keys, " "), want)
	}
	if d := varByKey(f, "dev.lisans.web"); d.Doc != "only in DEBUG builds:" {
		t.Errorf("doc of first disabled: %q", d.Doc)
	}
	if d := varByKey(f, "dev.lisans.mobil"); d.GroupDoc != "only in DEBUG builds:" {
		t.Errorf("group doc of second disabled: %q", d.GroupDoc)
	}

	// enable one, disable PORT=8080
	for i := range f.Variables {
		v := &f.Variables[i]
		switch {
		case v.Key == "dev.lisans.web":
			v.Disabled = false
		case v.Key == "PORT" && !v.Disabled:
			v.Disabled = true
		}
	}
	out := FormatEnvFile(f)
	if !strings.Contains(out, "\ndev.lisans.web=5\n") || !strings.Contains(out, "\n# PORT=8080\n") {
		t.Errorf("toggling failed:\n%s", out)
	}

	// commented-out variables don't count as duplicates or in comparisons
	g := ParseEnv(".env", "A=1\n# A=2\n")
	if len(Validate(g, nil)) != 0 {
		t.Error("disabled duplicate reported")
	}
	rows := Compare(g, ParseEnv("b", "A=1\n"))
	if len(rows) != 1 || rows[0].Kind != DiffSame {
		t.Errorf("compare: %+v", rows)
	}
}
