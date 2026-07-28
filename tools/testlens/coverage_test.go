package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllExtensions(t *testing.T) {
	mappings := mapLanguageExtensions()
	exts := allExtensions(mappings)

	if len(exts) == 0 {
		t.Error("allExtensions should return non-empty slice")
	}

	seen := make(map[string]bool)
	for _, e := range exts {
		if seen[e] {
			t.Errorf("duplicate extension: %s", e)
		}
		seen[e] = true
	}
}

func TestDetectLanguage(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 6; i++ {
		os.WriteFile(filepath.Join(tmpDir, "file"+string(rune('a'+i))+".go"), []byte("package main"), 0644)
	}

	if got := detectLanguage(tmpDir, nil); got != "go" {
		t.Errorf("detectLanguage = %q, want go", got)
	}
}

func TestDetectLanguage_SinUmbral(t *testing.T) {
	// Sin umbral: un solo archivo alcanza para detectar el lenguaje.
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("package main"), 0644)

	if got := detectLanguage(tmpDir, nil); got != "go" {
		t.Errorf("detectLanguage con 1 archivo = %q, want go", got)
	}
}

func TestDetectLanguage_SinArchivos(t *testing.T) {
	if got := detectLanguage(t.TempDir(), nil); got != "" {
		t.Errorf("detectLanguage sin fuentes = %q, want \"\"", got)
	}
}

func TestExtensionsForLanguageGo(t *testing.T) {
	exts := extensionsForLanguage("go")
	if len(exts) != 1 || exts[0] != ".go" {
		t.Errorf("extensionsForLanguage(go) = %v, want [.go]", exts)
	}
}

func TestExtensionsForLanguageUnknown(t *testing.T) {
	exts := extensionsForLanguage("nonexistent")
	if exts != nil {
		t.Errorf("extensionsForLanguage(unknown) = %v, want nil", exts)
	}
}

func TestCheckCoverageConfig(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "foo.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "foo_test.go"), []byte("package main\nfunc TestFoo(t *testing.T){}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "bar.go"), []byte("package main"), 0644)

	cfg := config{language: "go", root: tmpDir, fail: false}
	violations, _, err := checkCoverageCounts(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}

	// Go package mode: foo_test.go covers the whole package, including bar.go
	if violations != 0 {
		t.Errorf("checkCoverage = %d violations, want 0 (foo_test.go covers package)", violations)
	}
}

func TestCheckCoverageAllTested(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "foo.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "foo_test.go"), []byte("package main\nfunc TestFoo(t *testing.T){}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "bar_test.go"), []byte("package main\nfunc TestBar(t *testing.T){}"), 0644)

	cfg := config{language: "go", root: tmpDir, fail: false}
	violations, _, err := checkCoverageCounts(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}

	if violations != 0 {
		t.Errorf("checkCoverage = %d violations, want 0", violations)
	}
}

func TestIsTestFilePython(t *testing.T) {
	if !isTestFile("foo_test.py", []string{".py"}, []string{"_test"}) {
		t.Error("foo_test.py should be a test file for python")
	}
	if isTestFile("foo.py", []string{".py"}, []string{"_test"}) {
		t.Error("foo.py should NOT be a test file for python")
	}
}

func TestIsTestFileRuby(t *testing.T) {
	if !isTestFile("foo_spec.rb", []string{".rb"}, []string{"_spec", "_test"}) {
		t.Error("foo_spec.rb should be a test file for ruby")
	}
}
