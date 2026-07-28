package main

import (
	"os"
	"path/filepath"
	"testing"
)

// fixturesCfg usa ventana chica y umbral bajo para que los pares cortos de la
// carpeta testdata/fixtures se detecten.
func fixturesCfg() Config {
	return Config{
		Default: DefaultConfig{MinTokens: 10, MinLines: 3, WindowSize: 8},
		Exclude: []string{".git"},
	}
}

func scanFixtures(t *testing.T) []Match {
	t.Helper()
	root := filepath.Join("testdata", "fixtures")
	matches, _, _, err := scan(root, fixturesCfg(), 0)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	return matches
}

// findMatch busca un match entre dos archivos (en cualquier orden).
func findMatch(matches []Match, a, b string) *Match {
	for i := range matches {
		if (matches[i].FileA == a && matches[i].FileB == b) ||
			(matches[i].FileA == b && matches[i].FileB == a) {
			return &matches[i]
		}
	}
	return nil
}

func TestFixtures_exactPairLabeledExact(t *testing.T) {
	m := findMatch(scanFixtures(t), "exact_a.go", "exact_b.go")
	if m == nil {
		t.Fatal("no se detectó el par exacto exact_a.go ↔ exact_b.go")
	}
	if m.Kind != "exact" {
		t.Errorf("Kind = %q; want exact", m.Kind)
	}
}

func TestFixtures_renamedPairLabeledRenamed(t *testing.T) {
	m := findMatch(scanFixtures(t), "renamed_user.go", "renamed_product.go")
	if m == nil {
		t.Fatal("no se detectó el par renamed renamed_user.go ↔ renamed_product.go")
	}
	if m.Kind != "renamed" {
		t.Errorf("Kind = %q; want renamed", m.Kind)
	}
}

func TestFixtures_dataAndJsonNeverMatched(t *testing.T) {
	matches := scanFixtures(t)
	for _, m := range matches {
		for _, f := range []string{m.FileA, m.FileB} {
			if f == "data.csv" || f == "fixtures.json" {
				t.Errorf("archivo de datos %q no debe aparecer en matches", f)
			}
		}
	}
}

func TestFixtures_rustDeriveNotComment(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "fixtures", "derive.rs"))
	if err != nil {
		t.Fatalf("leyendo derive.rs: %v", err)
	}
	got := tokenValuesExt(string(data), ".rs")
	if !contains(got, "#[derive") && !contains(got, "derive") {
		t.Errorf("#[derive] no debe tratarse como comentario en Rust: %v", got)
	}
}

func TestFixtures_cssHexNotComment(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "fixtures", "colors.css"))
	if err != nil {
		t.Fatalf("leyendo colors.css: %v", err)
	}
	got := tokenValuesExt(string(data), ".css")
	if !contains(got, "#fff") {
		t.Errorf("color #fff no debe tratarse como comentario en CSS: %v", got)
	}
}
