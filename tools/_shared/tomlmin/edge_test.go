package tomlmin

import (
	"strings"
	"testing"
)

func TestEdge_DottedKeyInsideInlineTableOverScalar(t *testing.T) {
	if _, _, err := ExtractAsJSON([]byte("[a]\nt = { x = 1, x.y = 2 }\n"), "a"); err == nil {
		t.Error("expected error: x is already a scalar inside the inline table")
	}
}

func TestEdge_TableHeaderWithTrailingText(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte("[a] x\nk = 1\n"), "a")
	if err == nil || !strings.Contains(err.Error(), "unclosed table header") {
		t.Errorf("want unclosed-table-header error, got: %v", err)
	}
}

func TestEdge_KeySegmentsMalformed(t *testing.T) {
	cases := map[string]string{
		"quoted segment followed by text": "[a]\n\"x\" y = 1\n",
		"trailing dot":                    "[a]\nx. = 1\n",
		"leading dot":                     "[a]\n.x = 1\n",
		"unterminated quoted key":         "[a]\n\"x = 1\n",
		"invalid escape in quoted key":    "[a]\n\"x\\z\" = 1\n",
	}
	for name, src := range cases {
		if _, _, err := ExtractAsJSON([]byte(src), "a"); err == nil {
			t.Errorf("%s: expected key error for %q", name, src)
		}
	}
}

func TestEdge_ValueStartingWithDelimiter(t *testing.T) {
	if _, _, err := ExtractAsJSON([]byte("[a]\nk = ,\n"), "a"); err == nil {
		t.Error("expected error for a value that is only a delimiter")
	}
}

func TestEdge_AlmostADate(t *testing.T) {
	if _, _, err := ExtractAsJSON([]byte("[a]\nd = 1234-56-7x\n"), "a"); err == nil {
		t.Error("expected error: it looks like a date but is not one")
	}
}

func TestEdge_InvalidEscapeInMultilineString(t *testing.T) {
	if _, _, err := ExtractAsJSON([]byte("[a]\nk = \"\"\"x\\zy\"\"\"\n"), "a"); err == nil {
		t.Error("expected invalid-escape error inside a multiline string")
	}
}

func TestEdge_EscapedTripleQuoteInsideMultiline(t *testing.T) {
	// El delimitador escapado no cierra el string.
	m, _ := mustParse(t, "[a]\nk = \"\"\"x\\\"\"\"y\"\"\"\nj = 1\n", "a")
	if m["k"] != `x"""y` {
		t.Errorf("escaped delimiter: %q", m["k"])
	}
	if m["j"].(float64) != 1 {
		t.Errorf("parsing did not resume after the multiline string: %v", m)
	}
}

func TestEdge_MultilineEndingInBackslash(t *testing.T) {
	if _, _, err := ExtractAsJSON([]byte("[a]\nk = \"\"\"x\\\"\"\"\n"), "a"); err == nil {
		t.Error("expected unterminated-escape error")
	}
}
