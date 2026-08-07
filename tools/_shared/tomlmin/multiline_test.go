package tomlmin

import (
	"strings"
	"testing"
)

func TestMultiline_ArrayAcrossLines(t *testing.T) {
	src := `[project]
name = "demo"
dependencies = [
    "requests>=2.0",
    "pydantic>=2.0",
]

[tool.linelens]
maxLines = 200
`
	m, _ := mustParse(t, src, "tool.linelens")
	if m["maxLines"].(float64) != 200 {
		t.Errorf("maxLines: %v", m)
	}
	root, _ := mustParse(t, src, "project")
	deps := root["dependencies"].([]any)
	if len(deps) != 2 || deps[0] != "requests>=2.0" {
		t.Errorf("dependencies: %v", deps)
	}
}

func TestMultiline_TrailingCommaInArray(t *testing.T) {
	m, _ := mustParse(t, "[a]\narr = [\"x\", \"y\",]\n", "a")
	arr := m["arr"].([]any)
	if len(arr) != 2 || arr[1] != "y" {
		t.Errorf("trailing comma: %v", arr)
	}
}

func TestMultiline_TrailingCommaInInlineTableIsError(t *testing.T) {
	if _, _, err := ExtractAsJSON([]byte("[a]\nt = { k = 1, }\n"), "a"); err == nil {
		t.Error("expected error: inline tables reject a trailing comma")
	}
}

func TestMultiline_ArrayOfInlineTables(t *testing.T) {
	src := `[a]
authors = [
  { name = "Ada", email = "ada@example.com" },
  { name = "Bob" },
]
`
	m, _ := mustParse(t, src, "a")
	arr := m["authors"].([]any)
	if len(arr) != 2 || arr[0].(map[string]any)["name"] != "Ada" {
		t.Errorf("authors: %v", arr)
	}
}

func TestMultiline_CommentInsideArray(t *testing.T) {
	src := `[a]
arr = [
  "x",   # primero
  # solo comentario
  "y",
]
`
	m, _ := mustParse(t, src, "a")
	if len(m["arr"].([]any)) != 2 {
		t.Errorf("comments inside array: %v", m["arr"])
	}
}

func TestMultiline_ErrorReportsFirstPhysicalLine(t *testing.T) {
	src := "[a]\nx = 1\narr = [\n  1,\n  @,\n]\n"
	_, _, err := ExtractAsJSON([]byte(src), "a")
	if err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Errorf("want error anchored at line 3, got: %v", err)
	}
}

func TestMultiline_UnclosedDelimiterAtEOF(t *testing.T) {
	for _, src := range []string{"[a]\narr = [1, 2\n", "[a]\nt = { k = 1\n"} {
		if _, _, err := ExtractAsJSON([]byte(src), "a"); err == nil {
			t.Errorf("expected unclosed-delimiter error for %q", src)
		}
	}
}

func TestMultiline_BracketInsideStringDoesNotOpen(t *testing.T) {
	src := "[a]\nk = \"a ] b [ c\"\nj = 2\n"
	m, _ := mustParse(t, src, "a")
	if m["k"] != "a ] b [ c" || m["j"].(float64) != 2 {
		t.Errorf("brackets inside string leaked: %v", m)
	}
}

func TestMultiline_TableHeaderInsideMultilineString(t *testing.T) {
	src := "[a]\nk = \"\"\"\n[tool.linelens]\nmaxLines = 1\n\"\"\"\nj = 2\n"
	m, _ := mustParse(t, src, "a")
	if m["j"].(float64) != 2 {
		t.Errorf("j missing: %v", m)
	}
	if _, found := mustParse(t, src, "tool.linelens"); found {
		t.Error("a table header inside a multiline string must not open a table")
	}
}

func TestMultiline_HashInsideStrings(t *testing.T) {
	src := "[a]\nb = \"x # y\"\nl = 'z # w'\nm = \"\"\"p # q\"\"\"\n"
	got, _ := mustParse(t, src, "a")
	if got["b"] != "x # y" || got["l"] != "z # w" || got["m"] != "p # q" {
		t.Errorf("hash inside strings: %v", got)
	}
}
