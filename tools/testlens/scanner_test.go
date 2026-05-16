package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Extension mapping tests
func TestSourceExtensions(t *testing.T) {
	mappings := mapLanguageExtensions()
	
	// Verify at least one mapping per language
	langs := []string{"go", "typescript", "python", "ruby", "rust", "java", "csharp"}
	for _, lang := range langs {
		if _, ok := mappings[lang]; !ok {
			t.Errorf("missing language mapping for %s", lang)
		}
	}
	
	// Verify go has correct test pattern
	goSrc := mappings["go"]
	if len(goSrc.extensions) != 1 || goSrc.extensions[0] != ".go" {
		t.Errorf("go should have .go extension, got %v", goSrc.extensions)
	}
	if len(goSrc.testSuffixes) != 1 || goSrc.testSuffixes[0] != "_test" {
		t.Errorf("go test suffix should be _test, got %v", goSrc.testSuffixes)
	}
}

func TestSupportedLanguages(t *testing.T) {
	langs := supportedLanguages()
	if len(langs) < 7 {
		t.Errorf("expected at least 7 supported languages, got %d", len(langs))
	}
	
	// Check key languages are present
	found := false
	for _, l := range langs {
		if l == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("go should be in supported languages")
	}
}

// Match candidate tests
func TestFindTestCandidates(t *testing.T) {
	lang := languageMapping{
		extensions:   []string{".ts", ".tsx"},
		testSuffixes: []string{".test", ".spec"},
		testPrefixes: []string{"test_"},
	}
	
	cases := []struct {
		filename string
		want     []string
	}{
		{"foo.ts", []string{"foo.test.ts", "foo.spec.ts", "test_foo.ts"}},
		{"bar.tsx", []string{"bar.test.tsx", "bar.spec.tsx", "test_bar.tsx"}},
		// test files themselves should return empty (they don't need candidates)
		{"foo.test.ts", []string{}},
		{"foo.spec.tsx", []string{}},
	}

	for _, c := range cases {
		got := findTestCandidates(c.filename, lang)
		if !equalSlice(got, c.want) {
			t.Errorf("findTestCandidates(%s) = %v, want %v", c.filename, got, c.want)
		}
	}
}

// File type detection tests
func TestIsTestFile(t *testing.T) {
	// Go pattern
	if !isTestFile("util_test.go", []string{".go"}, []string{"_test"}) {
		t.Error("util_test.go should be a test file for go")
	}
	if isTestFile("util.go", []string{".go"}, []string{"_test"}) {
		t.Error("util.go should NOT be a test file for go")
	}
	
	// TypeScript pattern
	if !isTestFile("auth.test.ts", []string{".ts", ".tsx"}, []string{".test", ".spec"}) {
		t.Error("auth.test.ts should be a test file for typescript")
	}
}

func TestIsSourceFile(t *testing.T) {
	extensions := []string{".ts", ".tsx"}
	
	if !isSourceFile("foo.ts", extensions) {
		t.Error("foo.ts should be a source file")
	}
	if !isSourceFile("bar.tsx", extensions) {
		t.Error("bar.tsx should be a source file")
	}
	if isSourceFile("foo.test.ts", extensions) {
		t.Error("foo.test.ts should NOT be a source file (it's a test)")
	}
	if isSourceFile("foo.spec.ts", extensions) {
		t.Error("foo.spec.ts should NOT be a source file (it's a test)")
	}
}

// Directory scanning tests
func TestSkipDirectories(t *testing.T) {
	skipDirs := []string{"node_modules", ".git", "vendor", "dist", "build"}

	cases := []struct {
		path string
		want bool
	}{
		{"node_modules/lodash", true},
		{".git/hooks", true},
		{"vendor/pkg", true},
		{"src/components", false},
	}

	for _, c := range cases {
		got := shouldSkipDir(c.path, skipDirs)
		if got != c.want {
			t.Errorf("shouldSkipDir(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

// File existence tests
func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a file
	filePath := filepath.Join(tmpDir, "exists.txt")
	os.WriteFile(filePath, []byte("hello"), 0644)
	
	if !fileExists(filePath) {
		t.Error("fileExists should return true for existing file")
	}
	
	if fileExists(filepath.Join(tmpDir, "doesnt-exist.txt")) {
		t.Error("fileExists should return false for non-existing file")
	}
}

// Test existence checking
func TestTestExists(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create source file
	os.WriteFile(filepath.Join(tmpDir, "foo.go"), []byte("package main"), 0644)
	
	// Create test file
	os.WriteFile(filepath.Join(tmpDir, "foo_test.go"), []byte("package main"), 0644)
	
	// Test exists
	candidates := []string{"foo_test.go", "foo_spec.go"}
	found := testExists(filepath.Join(tmpDir, "foo.go"), candidates, mapLanguageExtensions()["go"])
	if found != "foo_test.go" {
		t.Errorf("expected foo_test.go, got %s", found)
	}
	
	// Test doesn't exist
	candidates2 := []string{"bar_test.go", "bar_spec.go"}
	found2 := testExists(filepath.Join(tmpDir, "bar.go"), candidates2, mapLanguageExtensions()["go"])
	if found2 != "" {
		t.Errorf("expected empty, got %s", found2)
	}
}

// Integration test with temp directory
func TestScanDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "src", "foo.ts"), []byte("export const x = 1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "bar.ts"), []byte("export const y = 2"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "foo.test.ts"), []byte("describe('foo', () => {})"), 0644)

	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Hello"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "config.json"), []byte("{}"), 0644)

	files, err := scanDirectory(tmpDir, []string{"node_modules"}, extensionsForLanguage("typescript"))
	if err != nil {
		t.Fatalf("scanDirectory failed: %v", err)
	}

	// All .ts files found
	if len(files) != 3 {
		t.Errorf("expected 3 .ts files, got %d: %v", len(files), files)
	}

	// After filtering, only non-test source files remain
	sourceFiles := filterSourceFiles(files, extensionsForLanguage("typescript"))
	if len(sourceFiles) != 2 {
		t.Errorf("expected 2 source files (foo.ts, bar.ts), got %d: %v", len(sourceFiles), sourceFiles)
	}
}

// Helper functions
func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}