package tomlmin

import (
	"strings"
	"testing"
)

func TestSubset_LiteralString(t *testing.T) {
	m, _ := mustParse(t, "[a]\nk = 'C:\\path\\n'\n", "a")
	if m["k"] != `C:\path\n` {
		t.Errorf("literal string must not unescape: %q", m["k"])
	}
}

func TestSubset_LiteralStringUnterminated(t *testing.T) {
	if _, _, err := ExtractAsJSON([]byte("[a]\nk = 'abc\n"), "a"); err == nil {
		t.Error("expected error for unterminated literal string")
	}
}

func TestSubset_MultilineBasicString(t *testing.T) {
	m, _ := mustParse(t, "[a]\nk = \"\"\"\nhola\nmundo\n\"\"\"\n", "a")
	if m["k"] != "hola\nmundo\n" {
		t.Errorf("multiline basic: %q", m["k"])
	}
}

func TestSubset_MultilineLiteralString(t *testing.T) {
	m, _ := mustParse(t, "[a]\nk = '''\nx\\ny\n'''\n", "a")
	if m["k"] != "x\\ny\n" {
		t.Errorf("multiline literal: %q", m["k"])
	}
}

func TestSubset_MultilineStringUnterminated(t *testing.T) {
	for _, src := range []string{"[a]\nk = \"\"\"abc\n", "[a]\nk = '''abc\n"} {
		if _, _, err := ExtractAsJSON([]byte(src), "a"); err == nil {
			t.Errorf("expected unterminated error for %q", src)
		}
	}
}

func TestSubset_IntegerForms(t *testing.T) {
	m, _ := mustParse(t, "[a]\nu = 1_000\nh = 0x1f\no = 0o17\nb = 0b101\n", "a")
	if m["u"].(float64) != 1000 || m["h"].(float64) != 31 {
		t.Errorf("underscore/hex: %v", m)
	}
	if m["o"].(float64) != 15 || m["b"].(float64) != 5 {
		t.Errorf("octal/binary: %v", m)
	}
}

func TestSubset_InvalidIntegerPrefix(t *testing.T) {
	for _, src := range []string{"[a]\nk = 0x\n", "[a]\nk = 0xzz\n", "[a]\nk = 0b2\n"} {
		if _, _, err := ExtractAsJSON([]byte(src), "a"); err == nil {
			t.Errorf("expected error for %q", src)
		}
	}
}

func TestSubset_DatesAsStrings(t *testing.T) {
	src := "[a]\nd = 2026-01-01\nts = 2026-01-01T10:00:00Z\nlocal = 2026-01-01 10:00:00\nt = 10:32:00\n"
	m, _ := mustParse(t, src, "a")
	if m["d"] != "2026-01-01" || m["ts"] != "2026-01-01T10:00:00Z" {
		t.Errorf("dates: %v", m)
	}
	if m["local"] != "2026-01-01 10:00:00" || m["t"] != "10:32:00" {
		t.Errorf("local datetime / time: %v", m)
	}
}

func TestSubset_DateInsideArray(t *testing.T) {
	m, _ := mustParse(t, "[a]\narr = [2026-01-01, 2026-02-02]\n", "a")
	arr := m["arr"].([]any)
	if len(arr) != 2 || arr[1] != "2026-02-02" {
		t.Errorf("dates in array: %v", arr)
	}
}

func TestSubset_UnrecognizedTokenStillErrors(t *testing.T) {
	_, _, err := ExtractAsJSON([]byte("[a]\nk = @foo\n"), "a")
	if err == nil || !strings.Contains(err.Error(), "@foo") {
		t.Errorf("want error naming the token, got: %v", err)
	}
}

func TestSubset_DottedKeyCreatesTable(t *testing.T) {
	m, _ := mustParse(t, "[tool.linelens]\ndefault.maxLines = 42\n", "tool.linelens")
	if m["default"].(map[string]any)["maxLines"].(float64) != 42 {
		t.Errorf("dotted key: %v", m)
	}
}

func TestSubset_DottedKeysShareParent(t *testing.T) {
	m, _ := mustParse(t, "[a]\nx.b = 1\nx.c = 2\n", "a")
	x := m["x"].(map[string]any)
	if x["b"].(float64) != 1 || x["c"].(float64) != 2 {
		t.Errorf("dotted keys must share the parent table: %v", x)
	}
}

func TestSubset_QuotedKeyIsLiteral(t *testing.T) {
	for _, src := range []string{"[a]\n\"m.sub\" = 1\n", "[a]\n'm.sub' = 1\n"} {
		m, _ := mustParse(t, src, "a")
		if m["m.sub"].(float64) != 1 {
			t.Errorf("quoted key must not nest (%q): %v", src, m)
		}
	}
}

func TestSubset_QuotedKeyWithDottedSuffix(t *testing.T) {
	m, _ := mustParse(t, "[a]\n\"m.sub\".c = 1\n", "a")
	if m["m.sub"].(map[string]any)["c"].(float64) != 1 {
		t.Errorf("quoted+dotted key: %v", m)
	}
}

func TestSubset_MalformedKeys(t *testing.T) {
	cases := []string{
		"[a]\na..b = 1\n",
		"[a]\n\"unterminated = 1\n",
		"[a]\n. = 1\n",
	}
	for _, src := range cases {
		if _, _, err := ExtractAsJSON([]byte(src), "a"); err == nil {
			t.Errorf("expected key error for %q", src)
		}
	}
}

func TestSubset_DottedKeyOverExistingScalarIsError(t *testing.T) {
	if _, _, err := ExtractAsJSON([]byte("[a]\nx = 1\nx.y = 2\n"), "a"); err == nil {
		t.Error("expected error: x is already a scalar")
	}
}

func TestSubset_QuotedKeyInsideInlineTable(t *testing.T) {
	m, _ := mustParse(t, "[a]\nt = { \"k.x\" = 1 }\n", "a")
	if m["t"].(map[string]any)["k.x"].(float64) != 1 {
		t.Errorf("quoted key in inline table: %v", m)
	}
}
