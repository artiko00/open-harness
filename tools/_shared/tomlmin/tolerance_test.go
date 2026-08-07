package tomlmin

import (
	"strings"
	"testing"
)

// Un TOML fuera del subset en una seccion ajena no debe tumbar la extraccion
// de la seccion que si nos interesa.
func TestTolerance_ForeignSectionGarbageIgnored(t *testing.T) {
	src := `[tool.poetry.dependencies]
weird = @@@ nope
another = ???

[tool.dupelens]
minTokens = 50
`
	m, _ := mustParse(t, src, "tool.dupelens")
	if m["minTokens"].(float64) != 50 {
		t.Errorf("foreign garbage broke the load: %v", m)
	}
}

func TestTolerance_OwnSectionGarbageStillFails(t *testing.T) {
	src := "[tool.dupelens]\nminTokens = @@@\n"
	_, _, err := ExtractAsJSON([]byte(src), "tool.dupelens")
	if err == nil {
		t.Fatal("expected a hard error inside the target section")
	}
	if !strings.Contains(err.Error(), "line 2") || !strings.Contains(err.Error(), "minTokens") {
		t.Errorf("error must name line and key, got: %v", err)
	}
}

func TestTolerance_SubtablesOfTargetAreStrict(t *testing.T) {
	cases := []string{
		"[tool.linelens.default]\nmaxLines = @@@\n",
		"[[tool.linelens.rules]]\npattern = @@@\n",
	}
	for _, src := range cases {
		if _, _, err := ExtractAsJSON([]byte(src), "tool.linelens"); err == nil {
			t.Errorf("subtables of the target must be strict: %q", src)
		}
	}
}

func TestTolerance_SiblingPrefixIsNotASubtable(t *testing.T) {
	// tool.linelenses no es sub-tabla de tool.linelens pese a compartir prefijo.
	src := "[tool.linelenses]\nx = @@@\n\n[tool.linelens]\nmaxLines = 5\n"
	m, _ := mustParse(t, src, "tool.linelens")
	if m["maxLines"].(float64) != 5 {
		t.Errorf("sibling prefix treated as subtable: %v", m)
	}
}

func TestTolerance_RootSectionStrictOnlyBeforeFirstHeader(t *testing.T) {
	if _, _, err := ExtractAsJSON([]byte("k = @@@\n"), ""); err == nil {
		t.Error("root assignments must be strict when the target is the root table")
	}
	m, _ := mustParse(t, "name = \"ok\"\n[other]\nk = @@@\n", "")
	if m["name"] != "ok" {
		t.Errorf("garbage in a table broke the root extraction: %v", m)
	}
}

func TestTolerance_MalformedHeaderAlwaysFails(t *testing.T) {
	cases := []string{
		"[project\nname = 1\n\n[tool.linelens]\nmaxLines = 1\n",
		"[a..b]\nx = 1\n\n[tool.linelens]\nmaxLines = 1\n",
	}
	for _, src := range cases {
		if _, _, err := ExtractAsJSON([]byte(src), "tool.linelens"); err == nil {
			t.Errorf("a malformed header must fail anywhere: %q", src)
		}
	}
}

func TestTolerance_AbsentSectionWithForeignGarbage(t *testing.T) {
	src := "[project]\nk = @@@\n"
	raw, found, err := ExtractAsJSON([]byte(src), "tool.linelens")
	if err != nil || found || raw != nil {
		t.Errorf("want (nil,false,nil), got (%v,%v,%v)", raw, found, err)
	}
}

func TestTolerance_DiscardedLineLeavesNoResidue(t *testing.T) {
	// La asignacion descartada no debe dejar la clave a medio construir en la
	// tabla ajena; el resto de esa tabla se conserva.
	src := "[project]\ngood = 1\nbad = @@@\nworse.x = @@@\n\n[tool.linelens]\nmaxLines = 1\n"
	root, err := parseDocument(src, "tool.linelens")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	project := root["project"].(map[string]any)
	if project["good"].(float64) != 1 {
		t.Errorf("good lost: %v", project)
	}
	if _, ok := project["bad"]; ok {
		t.Errorf("bad should have been dropped: %v", project)
	}
	if _, ok := project["worse"]; ok {
		t.Errorf("worse must not be materialized when its value fails: %v", project)
	}
}
