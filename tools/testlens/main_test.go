package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Main and run tests
func TestMain(t *testing.T) {
	orig := osExit
	defer func() { osExit = orig }()
	
	var exited int
	osExit = func(code int) { exited = code }
	
	main()
	if exited != 1 {
		t.Errorf("main() with no args: expected exit 1, got %d", exited)
	}
}

func TestRunNoArgs(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := run([]string{})

	w.Close()
	os.Stdout = old

	if code != 1 {
		t.Errorf("run([]string{}) = %d, want 1", code)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if !bytes.Contains(buf.Bytes(), []byte("testlens")) {
		t.Error("run([]string{}) should print usage with tool name")
	}
}

func TestRunHelp(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := run([]string{"help"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Errorf("run([\"help\"]) = %d, want 0", code)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if !bytes.Contains(buf.Bytes(), []byte("COMMANDS")) {
		t.Error("help should print commands section")
	}
}

func TestRunVersion(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := run([]string{"version"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Errorf("run([\"version\"]) = %d, want 0", code)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if !bytes.Contains(buf.Bytes(), []byte("testlens")) {
		t.Error("version should print testlens")
	}
}

func TestRunUnknown(t *testing.T) {
	old := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	code := run([]string{"unknown-cmd"})

	w.Close()
	os.Stderr = old

	if code != 1 {
		t.Errorf("run([\"unknown-cmd\"]) = %d, want 1", code)
	}
}

func TestRunInitCmd(t *testing.T) {
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "testlens.json")
	
	code := run([]string{"init", "--output", output})
	
	if code != 0 {
		t.Errorf("run([\"init\"]) = %d, want 0", code)
	}
	
	if _, err := os.Stat(output); err != nil {
		t.Errorf("init should create file: %v", err)
	}
}

func TestRunCheckCmd(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	
	code := run([]string{"check", "--lang", "go", "--dir", tmpDir})

	if code != 0 && code != 1 {
		t.Errorf("run([\"check\"]) = %d, want 0 or 1", code)
	}
}

// runCheck tests
func TestRunCheckNoArgs(t *testing.T) {
	code := runCheck([]string{})
	if code != 0 && code != 1 {
		t.Errorf("runCheck([]) = %d, want 0 or 1", code)
	}
}

func TestRunCheckWithLang(t *testing.T) {
	code := runCheck([]string{"--lang", "go", "--dir", os.TempDir()})
	
	if code != 0 && code != 1 {
		t.Errorf("runCheck with valid args: expected 0 or 1, got %d", code)
	}
}

func TestRunCheckFailFlag(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "foo.go"), []byte("package main"), 0644)
	
	code := runCheck([]string{"--dir", tmpDir})
	if code != 0 && code != 1 {
		t.Errorf("runCheck without --fail should return 0 or 1, got %d", code)
	}
}

func TestRunCheckInvalidLang(t *testing.T) {
	// Invalid language falls back to allExtensions, so it succeeds with 0
	code := runCheck([]string{"--lang", "invalid-lang", "--dir", os.TempDir()})
	// Should not crash, returns 0 (no violations) or 1 (something found)
	if code != 0 && code != 1 {
		t.Errorf("runCheck with invalid lang = %d, want 0 or 1", code)
	}
}

// runInit tests
func TestRunInit(t *testing.T) {
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "testlens.json")
	
	code := runInit([]string{"--output", output})
	if code != 0 {
		t.Errorf("runInit with valid output = %d, want 0", code)
	}
	
	if _, err := os.Stat(output); err != nil {
		t.Errorf("runInit should create file, got error: %v", err)
	}
}

func TestRunInitFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "testlens.json")
	os.WriteFile(output, []byte("{}"), 0644)
	
	old := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w
	
	code := runInit([]string{"--output", output})
	
	os.Stderr = old
	
	if code != 1 {
		t.Errorf("runInit on existing file = %d, want 1", code)
	}
}

// Coverage helpers tests
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

func TestDetectLanguageFromFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	for i := 0; i < 6; i++ {
		os.WriteFile(filepath.Join(tmpDir, "file"+string(rune('a'+i))+".go"), []byte("package main"), 0644)
	}
	
	exts := detectLanguageFromFiles(tmpDir)
	if exts == nil {
		t.Error("detectLanguageFromFiles should detect .go from 6 files")
	}
	if len(exts) != 1 || exts[0] != ".go" {
		t.Errorf("detectLanguageFromFiles = %v, want [.go]", exts)
	}
}

func TestDetectLanguageFromFilesTooFew(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("package main"), 0644)
	
	exts := detectLanguageFromFiles(tmpDir)
	if exts != nil {
		t.Error("detectLanguageFromFiles should return nil for fewer than 6 files")
	}
}

func TestGetExtensions(t *testing.T) {
	mappings := mapLanguageExtensions()
	
	exts := getExtensions("go", os.TempDir(), mappings)
	if len(exts) != 1 || exts[0] != ".go" {
		t.Errorf("getExtensions(go) = %v, want [.go]", exts)
	}
	
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("package main"), 0644)
	exts = getExtensions("auto", tmpDir, mappings)
	if exts == nil {
		t.Error("getExtensions(auto) should return extensions from temp dir")
	}
}

func TestGetLanguageMapping(t *testing.T) {
	mappings := mapLanguageExtensions()
	
	lang := getLanguageMapping("go", mappings, []string{".go"})
	if len(lang.extensions) != 1 {
		t.Errorf("getLanguageMapping(go) extensions = %v", lang.extensions)
	}
	
	lang = getLanguageMapping("auto", mappings, []string{".ts", ".tsx"})
	if len(lang.extensions) != 2 {
		t.Errorf("getLanguageMapping(auto) extensions = %v", lang.extensions)
	}
	if len(lang.testSuffixes) == 0 {
		t.Error("getLanguageMapping(auto) should have testSuffixes")
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

// Reporter tests
func TestPrintViolations(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	results := []ViolationResult{
		{Path: "/src/foo.go", HasTest: false},
		{Path: "/src/bar.go", HasTest: true, TestFor: "bar_test.go"},
		{Path: "/src/baz.go", HasTest: false},
	}

	count := printViolations(results, "/src")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if count != 2 {
		t.Errorf("printViolations = %d, want 2", count)
	}

	if !bytes.Contains([]byte(output), []byte("foo.go - no test found")) {
		t.Error("should print foo.go violation")
	}
	if !bytes.Contains([]byte(output), []byte("baz.go - no test found")) {
		t.Error("should print baz.go violation")
	}
}

func TestPrintSummary(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSummary(0)
	printSummary(3)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("All source files have tests")) {
		t.Error("printSummary(0) should print success message")
	}
	if !bytes.Contains([]byte(output), []byte("3 file(s) without tests")) {
		t.Error("printSummary(3) should print violation count")
	}
}

func TestExitWithCode(t *testing.T) {
	if exitWithCode(true, 0) != 0 {
		t.Errorf("exitWithCode(true, 0) should return 0")
	}
	
	if exitWithCode(true, 1) != 1 {
		t.Errorf("exitWithCode(true, 1) should return 1")
	}
	
	if exitWithCode(false, 1) != 0 {
		t.Errorf("exitWithCode(false, 1) should return 0")
	}
	
	if exitWithCode(false, 0) != 0 {
		t.Errorf("exitWithCode(false, 0) should return 0")
	}
}

// checkCoverage integration tests
func TestCheckCoverageConfig(t *testing.T) {
	tmpDir := t.TempDir()
	
	os.WriteFile(filepath.Join(tmpDir, "foo.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "foo_test.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "bar.go"), []byte("package main"), 0644)
	
	cfg := config{language: "go", root: tmpDir, fail: false}
	violations, err := checkCoverage(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}
	
	if violations != 1 {
		t.Errorf("checkCoverage = %d violations, want 1 (bar.go)", violations)
	}
}

func TestCheckCoverageAllTested(t *testing.T) {
	tmpDir := t.TempDir()
	
	os.WriteFile(filepath.Join(tmpDir, "foo.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "foo_test.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "bar_test.go"), []byte("package main"), 0644)
	
	cfg := config{language: "go", root: tmpDir, fail: false}
	violations, err := checkCoverage(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}
	
	if violations != 0 {
		t.Errorf("checkCoverage = %d violations, want 0", violations)
	}
}

// isTestFile more coverage
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

// scanDirectory more coverage
func TestScanDirectoryEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	
	files, err := scanDirectory(tmpDir, []string{}, []string{".go"})
	if err != nil {
		t.Fatalf("scanDirectory failed: %v", err)
	}
	
	if len(files) != 0 {
		t.Errorf("scanDirectory empty dir = %d files, want 0", len(files))
	}
}

func TestScanDirectoryWithSkipDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "vendor"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "vendor", "foo.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	
	files, err := scanDirectory(tmpDir, []string{"vendor"}, []string{".go"})
	if err != nil {
		t.Fatalf("scanDirectory failed: %v", err)
	}
	
	if len(files) != 1 {
		t.Errorf("scanDirectory with skip = %d files, want 1", len(files))
	}
}