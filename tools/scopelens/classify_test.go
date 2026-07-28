package main

import "testing"

func TestClassifyTestLayouts(t *testing.T) {
	tests := []string{
		"src/a.test.ts", "src/a.spec.js", "pkg/__tests__/x.ts", "app/tests/util.ts",
		"tests/test_b.py", "svc/mod_test.py", "pkg/c_test.go",
	}
	for _, p := range tests {
		if c := classify(p, defaultExclude); c != catTest {
			t.Errorf("classify(%q) = %q, want test", p, c)
		}
	}
}

func TestClassifySource(t *testing.T) {
	for _, p := range []string{"src/main.ts", "pkg/handler.go", "app/service.py"} {
		if c := classify(p, defaultExclude); c != catSource {
			t.Errorf("classify(%q) = %q, want source", p, c)
		}
	}
}

func TestClassifyExcludedPrioridad(t *testing.T) {
	// go.sum matchea exclude aunque no sea test; excluded gana.
	if c := classify("go.sum", defaultExclude); c != catExcluded {
		t.Fatalf("go.sum debería ser excluded, got %q", c)
	}
}

func TestExcludeReason(t *testing.T) {
	cases := map[string]string{
		"pnpm-lock.yaml":            "lockfile",
		"poetry.lock":               "lockfile",
		"go.sum":                    "lockfile",
		"api/service.pb.go":         "generated",
		"pkg/zz_generated_deep.go":  "generated",
		"node_modules/left-pad.js":  "excluded",
	}
	for p, want := range cases {
		if got := excludeReason(p); got != want {
			t.Errorf("excludeReason(%q) = %q, want %q", p, got, want)
		}
	}
}

func TestExcludeUsuarioReemplaza(t *testing.T) {
	// Con exclude propio, go.sum ya no matchea y docs/api.md sí.
	custom := []string{"docs/**"}
	if classify("docs/api.md", custom) != catExcluded {
		t.Error("docs/api.md debería ser excluded con exclude propio")
	}
	if classify("go.sum", custom) == catExcluded {
		t.Error("go.sum no debería excluirse cuando el usuario reemplaza exclude")
	}
}
