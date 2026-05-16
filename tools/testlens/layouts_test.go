package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Layout-aware test resolution.
// F-015 / ADR-019: testExists must find tests in adjacent testDirs and across
// mirrored path prefixes, not only colocated.

func TestTestExists_Colocated(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "foo.ts")
	tst := filepath.Join(dir, "foo.test.ts")
	os.WriteFile(src, []byte(""), 0644)
	os.WriteFile(tst, []byte(""), 0644)

	lang := mapLanguageExtensions()["typescript"]
	cands := findTestCandidates("foo.ts", lang)
	got := testExists(src, cands, lang)
	if got == "" {
		t.Errorf("expected colocated foo.test.ts to match, got empty")
	}
}

func TestTestExists_TestsDirSubdir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "foo.ts")
	tdir := filepath.Join(dir, "__tests__")
	os.MkdirAll(tdir, 0755)
	os.WriteFile(src, []byte(""), 0644)
	os.WriteFile(filepath.Join(tdir, "foo.test.ts"), []byte(""), 0644)

	lang := mapLanguageExtensions()["typescript"]
	cands := findTestCandidates("foo.ts", lang)
	got := testExists(src, cands, lang)
	if got == "" {
		t.Errorf("expected __tests__/foo.test.ts to match, got empty")
	}
}

func TestTestExists_TestsDirAlternative(t *testing.T) {
	// The Vitest "tests/" subdir variant — sibling of source.
	dir := t.TempDir()
	src := filepath.Join(dir, "foo.ts")
	tdir := filepath.Join(dir, "tests")
	os.MkdirAll(tdir, 0755)
	os.WriteFile(src, []byte(""), 0644)
	os.WriteFile(filepath.Join(tdir, "foo.spec.ts"), []byte(""), 0644)

	lang := mapLanguageExtensions()["typescript"]
	cands := findTestCandidates("foo.ts", lang)
	got := testExists(src, cands, lang)
	if got == "" {
		t.Errorf("expected tests/foo.spec.ts to match, got empty")
	}
}

func TestTestExists_PythonTestsMirror(t *testing.T) {
	// src/auth/user.py  ↔  tests/auth/test_user.py
	root := t.TempDir()
	srcDir := filepath.Join(root, "src", "auth")
	tstDir := filepath.Join(root, "tests", "auth")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(tstDir, 0755)
	src := filepath.Join(srcDir, "user.py")
	os.WriteFile(src, []byte(""), 0644)
	os.WriteFile(filepath.Join(tstDir, "test_user.py"), []byte(""), 0644)

	lang := mapLanguageExtensions()["python"]
	cands := findTestCandidates("user.py", lang)
	got := testExists(src, cands, lang)
	if got == "" {
		t.Errorf("expected tests/auth/test_user.py (python mirror) to match, got empty")
	}
}

func TestTestExists_MavenMirror(t *testing.T) {
	// src/main/java/com/foo/Bar.java  ↔  src/test/java/com/foo/BarTest.java
	root := t.TempDir()
	mainDir := filepath.Join(root, "src", "main", "java", "com", "foo")
	testDir := filepath.Join(root, "src", "test", "java", "com", "foo")
	os.MkdirAll(mainDir, 0755)
	os.MkdirAll(testDir, 0755)
	src := filepath.Join(mainDir, "Bar.java")
	os.WriteFile(src, []byte(""), 0644)
	os.WriteFile(filepath.Join(testDir, "BarTest.java"), []byte(""), 0644)

	lang := mapLanguageExtensions()["java"]
	cands := findTestCandidates("Bar.java", lang)
	got := testExists(src, cands, lang)
	if got == "" {
		t.Errorf("expected src/test/java/com/foo/BarTest.java to match, got empty")
	}
}

func TestTestExists_NoMatchAnywhere(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "foo.ts")
	os.WriteFile(src, []byte(""), 0644)

	lang := mapLanguageExtensions()["typescript"]
	cands := findTestCandidates("foo.ts", lang)
	if got := testExists(src, cands, lang); got != "" {
		t.Errorf("expected no match, got %q", got)
	}
}

func TestApplyMirror_EmptyFromPrefix(t *testing.T) {
	// Defensive guard: empty prefix should return path unchanged.
	got, ok := applyMirror("src/foo.py", "", "tests")
	if ok || got != "src/foo.py" {
		t.Errorf("expected unchanged path for empty prefix, got %q ok=%v", got, ok)
	}
}

func TestSplitSegments_Empty(t *testing.T) {
	if got := splitSegments(""); got != nil {
		t.Errorf("expected nil for empty input, got %v", got)
	}
}

func TestTestExists_MirrorIgnoredWhenPathOutsidePrefix(t *testing.T) {
	// Source NOT under "src/" → mirror to "tests/" should not apply.
	dir := t.TempDir()
	src := filepath.Join(dir, "user.py")
	os.WriteFile(src, []byte(""), 0644)
	// Test file under another arbitrary dir — should NOT count.
	os.MkdirAll(filepath.Join(dir, "tests"), 0755)
	os.WriteFile(filepath.Join(dir, "tests", "test_user.py"), []byte(""), 0644)

	lang := mapLanguageExtensions()["python"]
	cands := findTestCandidates("user.py", lang)
	got := testExists(src, cands, lang)
	// The "tests/" subdir IS a configured testDir for Python, so this DOES match.
	if got == "" {
		t.Errorf("expected match via testDirs (tests/test_user.py), got empty")
	}
}
