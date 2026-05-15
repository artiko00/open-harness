package tomlmin

import (
	"strings"
	"testing"
)

func TestExtract_AllStringEscapes(t *testing.T) {
	src := `
[a]
nl = "x\ny"
tab = "x\ty"
cr = "x\ry"
`
	m, _ := mustParse(t, src, "a")
	if m["nl"] != "x\ny" || m["tab"] != "x\ty" || m["cr"] != "x\ry" {
		t.Errorf("escapes: %v", m)
	}
}

func TestExtract_InvalidEscape(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\nk = \"x\\zy\""), "a")
	if err == nil {
		t.Error("expected error for invalid escape \\z")
	}
}

func TestExtract_UnterminatedEscape(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte("[a]\nk = \"x\\"), "a")
	if err == nil {
		t.Error("expected error for unterminated escape")
	}
}

func TestExtract_EmptyArray(t *testing.T) {
	src := `
[a]
arr = []
`
	m, _ := mustParse(t, src, "a")
	if len(m["arr"].([]any)) != 0 {
		t.Errorf("empty array expected, got: %v", m["arr"])
	}
}

func TestExtract_EmptyInlineTable(t *testing.T) {
	src := `
[a]
t = {}
`
	m, _ := mustParse(t, src, "a")
	if len(m["t"].(map[string]any)) != 0 {
		t.Errorf("empty inline table expected, got: %v", m["t"])
	}
}

func TestExtract_InvalidBool(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\nk = trueish"), "a")
	if err == nil {
		t.Error("expected error for invalid bool literal")
	}
}

func TestExtract_InvalidNumber(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\nk = 1.2.3"), "a")
	if err == nil {
		t.Error("expected error for malformed number 1.2.3")
	}
}

func TestExtract_TrailingGarbage(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\nk = 1 garbage"), "a")
	if err == nil {
		t.Error("expected error for trailing garbage after value")
	}
}

func TestExtract_InlineTableMissingEquals(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\nt = { k 1 }"), "a")
	if err == nil {
		t.Error("expected error for inline table without '='")
	}
}

func TestExtract_InlineTableEmptyKey(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\nt = { = 1 }"), "a")
	if err == nil {
		t.Error("expected error for inline table with empty key")
	}
}

func TestExtract_InlineTableNoComma(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\nt = { k = 1 v = 2 }"), "a")
	if err == nil {
		t.Error("expected error for inline table missing ','")
	}
}

func TestExtract_ArrayNoComma(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\narr = [1 2]"), "a")
	if err == nil {
		t.Error("expected error for array missing ','")
	}
}

func TestExtract_ArrayPropagatesInnerError(t *testing.T) {
	// Inner value parse fails (invalid escape) and the error must bubble up.
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\narr = [\"x\\zy\"]"), "a")
	if err == nil {
		t.Error("expected error propagated from array element")
	}
}

func TestExtract_InlineTableValueErrorPropagates(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte(`[a]`+"\nt = { k = \"x\\z\" }"), "a")
	if err == nil {
		t.Error("expected error propagated from inline table value")
	}
}

func TestExtract_EmptyValue(t *testing.T) {
	// A "= " with only whitespace after — caught by applyAssignment
	if _, _, err := ExtractAsJSON([]byte("[a]\nk =   "), "a"); err == nil {
		t.Error("expected error for empty value")
	}
}

func TestExtract_TableHeaderSubsequentRedefinesSameKey(t *testing.T) {
	// Re-opening an already-existing nested table must keep the previous values.
	src := `
[a.b]
x = 1
[a.b]
y = 2
`
	m, _ := mustParse(t, src, "a.b")
	if m["x"].(float64) != 1 || m["y"].(float64) != 2 {
		t.Errorf("table re-open lost data: %v", m)
	}
}

func TestExtract_TableHeaderEmptySegment(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte("[a..b]\nk = 1"), "a")
	if err == nil || !strings.Contains(err.Error(), "empty table segment") {
		t.Errorf("expected empty-segment error, got: %v", err)
	}
}

func TestExtract_InvalidLiteralStartingWithT(t *testing.T) {
	// parseBool returns the "invalid literal" error when prefix is t/f but
	// not 'true'/'false'.
	if _, _, err := ExtractAsJSON([]byte("[a]\nk = trois"), "a"); err == nil {
		t.Error("expected error for literal starting with 't' but not 'true'")
	}
	if _, _, err := ExtractAsJSON([]byte("[a]\nk = frue"), "a"); err == nil {
		t.Error("expected error for literal starting with 'f' but not 'false'")
	}
}

func TestExtract_ArrayOfTablesIntermediateMissing(t *testing.T) {
	// First time [[a.b.c]] is seen, neither a nor a.b exist yet — covers
	// the !ok branch of mountTableSlice for intermediate segments.
	src := `
[[a.b.c]]
v = 1
`
	root, _ := mustParse(t, src, "")
	if root["a"].(map[string]any)["b"].(map[string]any)["c"].([]any)[0].(map[string]any)["v"].(float64) != 1 {
		t.Errorf("deep array of tables failed: %v", root)
	}
}

func TestExtract_ArrayOfTablesAtRoot(t *testing.T) {
	// Single-segment array of tables: forces mountTableSlice(parts=[]) path.
	src := `
[[items]]
v = 1

[[items]]
v = 2
`
	root, _ := mustParse(t, src, "")
	arr := root["items"].([]any)
	if len(arr) != 2 {
		t.Errorf("want 2 items, got %d", len(arr))
	}
}

func TestExtract_ScalarValueErrorBubbles(t *testing.T) {
	// Triggers parseNumber returning "unrecognized value" via parseValue.
	if _, _, err := ExtractAsJSON([]byte("[a]\nk = @"), "a"); err == nil {
		t.Error("expected error for unrecognized scalar")
	}
}

func TestExtract_StringMissingOpenQuote(t *testing.T) {
	// parseString being called without the quote shouldn't happen via parseValue,
	// but the guard is part of the public guarantee — test it directly.
	if _, _, err := parseString("x"); err == nil {
		t.Error("expected error for missing opening quote")
	}
}

func TestExtract_ArrayOfTablesIntoExistingParent(t *testing.T) {
	// First the parent table is created, then [[parent.list]] appends.
	src := `
[parent]
name = "p"

[[parent.list]]
v = 1

[[parent.list]]
v = 2
`
	m, _ := mustParse(t, src, "parent")
	if m["name"] != "p" {
		t.Errorf("parent.name lost: %v", m)
	}
	arr := m["list"].([]any)
	if len(arr) != 2 || arr[1].(map[string]any)["v"].(float64) != 2 {
		t.Errorf("array of tables: %v", arr)
	}
}
