package main

import (
	"path/filepath"
	"testing"
)

// e2eFixtureCfg construye la config con la que corremos el E2E:
// minTokens=30 (el bloque duplicado de la fixture es ~80 tokens),
// minLines=5, sin reglas de skip más allá de los .test/.spec habituales.
func e2eFixtureCfg() Config {
	return Config{
		Default: DefaultConfig{MinTokens: 30, MinLines: 5},
		Rules: []Rule{
			{Pattern: "**/*_test.go", Skip: true},
		},
		Exclude: []string{".git"},
	}
}

func TestE2E_findsKnownDuplicateInFixture(t *testing.T) {
	root := filepath.Join("testdata", "e2e_fixture")
	matches, scanned, _, err := scan(root, e2eFixtureCfg(), 0)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if scanned < 3 {
		t.Errorf("expected ≥3 archivos escaneados (pair_a, pair_b, unique), got %d", scanned)
	}
	if len(matches) == 0 {
		t.Fatal("expected al menos 1 match entre pair_a.go y pair_b.go, got 0")
	}
	// Buscar el match esperado.
	found := false
	for _, m := range matches {
		hasA := m.FileA == "pair_a.go" || m.FileB == "pair_a.go"
		hasB := m.FileA == "pair_b.go" || m.FileB == "pair_b.go"
		if hasA && hasB {
			found = true
			if m.Tokens < 30 {
				t.Errorf("match a-b debe tener ≥30 tokens, got %d", m.Tokens)
			}
		}
	}
	if !found {
		t.Errorf("no se encontró match entre pair_a.go y pair_b.go en: %+v", matches)
	}
}

func TestE2E_uniqueFileNotInAnyMatch(t *testing.T) {
	root := filepath.Join("testdata", "e2e_fixture")
	matches, _, _, err := scan(root, e2eFixtureCfg(), 0)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	for _, m := range matches {
		if m.FileA == "unique.go" || m.FileB == "unique.go" {
			t.Errorf("unique.go no debe aparecer en matches, pero apareció: %+v", m)
		}
	}
}

func TestE2E_minTokensTooHigh_findsNothing(t *testing.T) {
	root := filepath.Join("testdata", "e2e_fixture")
	cfg := e2eFixtureCfg()
	matches, _, _, err := scan(root, cfg, 500)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("con minTokens=500 no debería haber matches en una fixture chica, got %d", len(matches))
	}
}

func BenchmarkScan_e2eFixture(b *testing.B) {
	root := filepath.Join("testdata", "e2e_fixture")
	cfg := e2eFixtureCfg()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, err := scan(root, cfg, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}
