package tomlmin

import (
	"encoding/json"
	"strings"
	"testing"
)

func mustParse(t *testing.T, src, section string) (map[string]any, bool) {
	t.Helper()
	data, found, err := ExtractAsJSON([]byte(src), section)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		return nil, false
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("json unmarshal failed: %v\nraw: %s", err, data)
	}
	return out, true
}

func TestExtract_SectionAbsent(t *testing.T) {
	_, found := mustParse(t, `# empty doc`, "tool.linelens")
	if found {
		t.Error("expected section absent")
	}
}

func TestExtract_EmptyTable(t *testing.T) {
	m, found := mustParse(t, "[tool.linelens]\n", "tool.linelens")
	if !found || len(m) != 0 {
		t.Errorf("expected empty map, got %v (found=%v)", m, found)
	}
}

func TestExtract_ScalarTypes(t *testing.T) {
	src := `
[tool.linelens]
name = "linelens"
maxLines = 100
ratio = 0.75
enabled = true
disabled = false
`
	m, _ := mustParse(t, src, "tool.linelens")
	if m["name"] != "linelens" {
		t.Errorf("name: %v", m["name"])
	}
	if m["maxLines"].(float64) != 100 {
		t.Errorf("maxLines: %v", m["maxLines"])
	}
	if m["ratio"].(float64) != 0.75 {
		t.Errorf("ratio: %v", m["ratio"])
	}
	if m["enabled"] != true || m["disabled"] != false {
		t.Errorf("bools wrong: %v %v", m["enabled"], m["disabled"])
	}
}

func TestExtract_NestedTables(t *testing.T) {
	src := `
[tool.linelens]
top = "a"

[tool.linelens.default]
maxLines = 100
`
	m, _ := mustParse(t, src, "tool.linelens")
	if m["top"] != "a" {
		t.Errorf("top missing: %v", m)
	}
	d := m["default"].(map[string]any)
	if d["maxLines"].(float64) != 100 {
		t.Errorf("nested maxLines: %v", d)
	}
}

func TestExtract_InlineTable(t *testing.T) {
	src := `
[tool.linelens]
default = { maxLines = 100, enabled = true }
`
	m, _ := mustParse(t, src, "tool.linelens")
	d := m["default"].(map[string]any)
	if d["maxLines"].(float64) != 100 || d["enabled"] != true {
		t.Errorf("inline table wrong: %v", d)
	}
}

func TestExtract_ArrayOfStrings(t *testing.T) {
	src := `
[tool.linelens]
exclude = ["node_modules", "vendor", "dist"]
`
	m, _ := mustParse(t, src, "tool.linelens")
	arr := m["exclude"].([]any)
	if len(arr) != 3 || arr[0] != "node_modules" || arr[2] != "dist" {
		t.Errorf("array of strings: %v", arr)
	}
}

func TestExtract_ArrayOfInlineTables(t *testing.T) {
	src := `
[tool.linelens]
rules = [{ pattern = "**/*_test.go", maxLines = 300 }, { pattern = "**/migrations/**", skip = true }]
`
	m, _ := mustParse(t, src, "tool.linelens")
	arr := m["rules"].([]any)
	r0 := arr[0].(map[string]any)
	if r0["pattern"] != "**/*_test.go" || r0["maxLines"].(float64) != 300 {
		t.Errorf("rule[0]: %v", r0)
	}
	r1 := arr[1].(map[string]any)
	if r1["skip"] != true {
		t.Errorf("rule[1].skip: %v", r1)
	}
}

func TestExtract_ArrayOfTables(t *testing.T) {
	src := `
[tool.linelens]
top = "x"

[[tool.linelens.rules]]
pattern = "**/*_test.go"
maxLines = 300

[[tool.linelens.rules]]
pattern = "**/migrations/**"
skip = true
`
	m, _ := mustParse(t, src, "tool.linelens")
	arr := m["rules"].([]any)
	if len(arr) != 2 {
		t.Fatalf("want 2 rules, got %d: %v", len(arr), arr)
	}
	r0 := arr[0].(map[string]any)
	if r0["pattern"] != "**/*_test.go" || r0["maxLines"].(float64) != 300 {
		t.Errorf("rule[0]: %v", r0)
	}
	r1 := arr[1].(map[string]any)
	if r1["skip"] != true {
		t.Errorf("rule[1].skip: %v", r1)
	}
}

func TestExtract_Comments(t *testing.T) {
	src := `
# top-level comment
[tool.linelens]
# inside comment
maxLines = 100 # trailing
name = "linelens" # another
`
	m, _ := mustParse(t, src, "tool.linelens")
	if m["maxLines"].(float64) != 100 || m["name"] != "linelens" {
		t.Errorf("comments leaked: %v", m)
	}
}

func TestExtract_OtherSectionsIgnored(t *testing.T) {
	src := `
[tool.dupelens]
minTokens = 50

[tool.linelens]
maxLines = 100

[project]
name = "irrelevant"
`
	m, _ := mustParse(t, src, "tool.linelens")
	if len(m) != 1 || m["maxLines"].(float64) != 100 {
		t.Errorf("other sections leaked: %v", m)
	}
}

func TestExtract_StringEscapes(t *testing.T) {
	src := `
[tool.linelens]
path = "a\\b"
quote = "x\"y"
`
	m, _ := mustParse(t, src, "tool.linelens")
	if m["path"] != `a\b` || m["quote"] != `x"y` {
		t.Errorf("escapes: %v", m)
	}
}

func TestExtract_NegativeNumbers(t *testing.T) {
	src := `
[tool.linelens]
n = -5
f = -1.5
`
	m, _ := mustParse(t, src, "tool.linelens")
	if m["n"].(float64) != -5 || m["f"].(float64) != -1.5 {
		t.Errorf("negatives: %v", m)
	}
}

func TestExtract_MalformedTableHeader(t *testing.T) {
	cases := []string{
		"[unclosed",
		"]",
		"[[a]",
		"[]",
	}
	for _, src := range cases {
		if _, _, err := ExtractAsJSON([]byte(src), "x"); err == nil {
			t.Errorf("expected error for: %q", src)
		}
	}
}

func TestExtract_MalformedAssignment(t *testing.T) {
	cases := []string{
		"[a]\nkey",            // no equals
		"[a]\n= 5",            // no key
		`[a]` + "\nk = ",       // no value
		`[a]` + "\nk = \"unterm",
	}
	for _, src := range cases {
		if _, _, err := ExtractAsJSON([]byte(src), "a"); err == nil {
			t.Errorf("expected error for: %q", src)
		}
	}
}

func TestExtract_UnsupportedFeature_DottedKey(t *testing.T) {
	src := `
[tool.linelens]
default.maxLines = 100
`
	_, _, err := ExtractAsJSON([]byte(src), "tool.linelens")
	if err == nil || !strings.Contains(err.Error(), "dotted") {
		t.Errorf("expected dotted-key error, got: %v", err)
	}
}

func TestExtract_MalformedArray(t *testing.T) {
	cases := []string{
		`[a]` + "\narr = [1, 2",           // unclosed
		`[a]` + "\narr = [{ k = 1",        // unclosed inline table
		`[a]` + "\narr = [1, , 3]",        // empty slot
	}
	for _, src := range cases {
		if _, _, err := ExtractAsJSON([]byte(src), "a"); err == nil {
			t.Errorf("expected error for: %q", src)
		}
	}
}

func TestExtract_RootLevelKeys(t *testing.T) {
	// keys before any table header live at the root section
	src := `
name = "root-level"
[tool.linelens]
nested = true
`
	m, _ := mustParse(t, src, "")
	if m["name"] != "root-level" {
		t.Errorf("root key missing: %v", m)
	}
}

func TestExtract_DeeplyNestedPath(t *testing.T) {
	src := `
[a.b.c.d]
x = 1
`
	m, _ := mustParse(t, src, "a.b.c.d")
	if m["x"].(float64) != 1 {
		t.Errorf("deep path: %v", m)
	}
}
