package tomlmin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Los cuatro pyproject.toml de referencia cubren los estilos que un usuario
// real trae consigo: poetry, setuptools, hatch y uv. Ninguno debe abortar la
// extraccion de [tool.linelens], aunque el resto del documento use sintaxis
// fuera del subset.
func TestGolden_RealPyprojects(t *testing.T) {
	files, err := filepath.Glob("testdata/*.toml")
	if err != nil || len(files) != 4 {
		t.Fatalf("testdata glob: %v (%d files)", err, len(files))
	}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		raw, found, err := ExtractAsJSON(src, "tool.linelens")
		if err != nil {
			t.Errorf("%s: unexpected error: %v", f, err)
			continue
		}
		if !found {
			t.Errorf("%s: [tool.linelens] not found", f)
			continue
		}
		var cfg map[string]any
		if err := json.Unmarshal(raw, &cfg); err != nil {
			t.Errorf("%s: json: %v", f, err)
			continue
		}
		if cfg["maxLines"].(float64) != 120 {
			t.Errorf("%s: maxLines = %v, want 120", f, cfg["maxLines"])
		}
	}
}

func TestGolden_PoetryDetails(t *testing.T) {
	src, _ := os.ReadFile("testdata/poetry.toml")
	m, _ := mustParse(t, string(src), "tool.linelens")
	arr := m["exclude"].([]any)
	if len(arr) != 2 || arr[0] != "vendor/**" || arr[1] != "build/**" {
		t.Errorf("exclude: %v", arr)
	}
}

func TestGolden_HatchNestedTables(t *testing.T) {
	src, _ := os.ReadFile("testdata/hatch.toml")
	m, _ := mustParse(t, string(src), "tool.linelens")
	if m["default"].(map[string]any)["maxLines"].(float64) != 120 {
		t.Errorf("default: %v", m["default"])
	}
	rules := m["rules"].([]any)
	if len(rules) != 1 || rules[0].(map[string]any)["maxLines"].(float64) != 300 {
		t.Errorf("rules: %v", rules)
	}
}

func TestGolden_UvDateAsString(t *testing.T) {
	src, _ := os.ReadFile("testdata/uv.toml")
	m, _ := mustParse(t, string(src), "tool.linelens")
	if m["since"] != "2026-01-01" {
		t.Errorf("since = %v, want the raw date string", m["since"])
	}
}
