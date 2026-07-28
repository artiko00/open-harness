package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDart_Registered(t *testing.T) {
	m := mapLanguageExtensions()
	lang, ok := m["dart"]
	if !ok {
		t.Fatal("expected 'dart' to be a registered language")
	}
	if len(lang.extensions) != 1 || lang.extensions[0] != ".dart" {
		t.Errorf("expected .dart extension, got %v", lang.extensions)
	}
}

func TestDart_InSupportedLanguages(t *testing.T) {
	found := false
	for _, l := range supportedLanguages() {
		if l == "dart" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'dart' in supportedLanguages()")
	}
}

func TestDart_TestMirrorLibToTest(t *testing.T) {
	// Flutter/Dart canonical layout: lib/services/foo.dart ↔ test/services/foo_test.dart
	root := t.TempDir()
	srcDir := filepath.Join(root, "lib", "services")
	tstDir := filepath.Join(root, "test", "services")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(tstDir, 0755)
	src := filepath.Join(srcDir, "foo.dart")
	os.WriteFile(src, []byte(""), 0644)
	os.WriteFile(filepath.Join(tstDir, "foo_test.dart"),
		[]byte("void main(){ test('x', (){}); }"), 0644)

	lang := mapLanguageExtensions()["dart"]
	cands := findTestCandidates("foo.dart", lang)
	if got := testExists(src, root, cands, lang); got == "" {
		t.Errorf("expected test/services/foo_test.dart to match lib/services/foo.dart")
	}
}

func TestDart_Colocated(t *testing.T) {
	// Some Dart projects colocate: foo.dart ↔ foo_test.dart
	dir := t.TempDir()
	src := filepath.Join(dir, "foo.dart")
	os.WriteFile(src, []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "foo_test.dart"),
		[]byte("void main(){ test('x', (){}); }"), 0644)

	lang := mapLanguageExtensions()["dart"]
	cands := findTestCandidates("foo.dart", lang)
	if got := testExists(src, dir, cands, lang); got == "" {
		t.Errorf("expected colocated foo_test.dart to match")
	}
}

func TestDart_NoTestIsViolation(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "orphan.dart")
	os.WriteFile(src, []byte(""), 0644)

	lang := mapLanguageExtensions()["dart"]
	cands := findTestCandidates("orphan.dart", lang)
	if got := testExists(src, dir, cands, lang); got != "" {
		t.Errorf("expected no match for orphan.dart, got %q", got)
	}
}
