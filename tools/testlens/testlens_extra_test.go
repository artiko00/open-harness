package main

import (
	"os"
	"path/filepath"
	"testing"
)

// runCheck — ramas faltantes

func TestRunCheck_ParseError(t *testing.T) {
	code := runCheck([]string{"--not-a-real-flag-xyz"})
	if code != 1 {
		t.Errorf("runCheck with bad flag = %d, want 1", code)
	}
}

func TestRunCheck_CoverageError(t *testing.T) {
	// directorio inexistente → checkCoverage retorna error
	code := runCheck([]string{"--lang", "go", "--dir", "/nonexistent/path/abc123xyz"})
	if code != 1 {
		t.Errorf("runCheck with nonexistent dir = %d, want 1", code)
	}
}

func TestRunCheck_FailWithViolations(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "foo.go"), []byte("package main"), 0644)
	// sin test file → violación → con --fail debe retornar 1
	code := runCheck([]string{"--lang", "go", "--dir", tmpDir, "--fail"})
	if code != 1 {
		t.Errorf("runCheck --fail with violations = %d, want 1", code)
	}
}

func TestRunCheck_NoViolations(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "foo.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "foo_test.go"), []byte("package main"), 0644)
	// todos los archivos tienen tests → "All source files have tests"
	code := runCheck([]string{"--lang", "go", "--dir", tmpDir})
	if code != 0 {
		t.Errorf("runCheck with all files tested = %d, want 0", code)
	}
}

// runInit — ramas faltantes

func TestRunInit_ParseError(t *testing.T) {
	code := runInit([]string{"--not-a-real-flag-xyz"})
	if code != 1 {
		t.Errorf("runInit with bad flag = %d, want 1", code)
	}
}

func TestRunInit_WriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod test not reliable as root")
	}
	tmpDir := t.TempDir()
	os.Chmod(tmpDir, 0555)
	defer os.Chmod(tmpDir, 0755)
	output := filepath.Join(tmpDir, "testlens.json")
	// Stat falla (no existe), WriteFile falla (directorio solo lectura)
	code := runInit([]string{"--output", output})
	if code != 1 {
		t.Errorf("runInit in readonly dir = %d, want 1", code)
	}
}

// checkCoverage — rama return filepath.SkipDir faltante

func TestCheckCoverage_SkipsNodeModules(t *testing.T) {
	tmpDir := t.TempDir()
	nm := filepath.Join(tmpDir, "node_modules")
	os.MkdirAll(nm, 0755)
	os.WriteFile(filepath.Join(nm, "foo.go"), []byte("package foo"), 0644)

	cfg := config{language: "go", root: tmpDir, fail: false}
	violations, err := checkCoverage(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}
	if violations != 0 {
		t.Errorf("node_modules should be skipped, got %d violations", violations)
	}
}

// isTestFile — rama HasPrefix que retorna true

func TestIsTestFile_PrefixIndicator(t *testing.T) {
	// indicator "test_" → HasSuffix("test_foo", "test_") = false, HasPrefix("test_foo", "test_") = true
	if !isTestFile("test_foo.py", []string{".py"}, []string{"test_"}) {
		t.Error("test_foo.py with indicator 'test_' should be detected via HasPrefix")
	}
}

// isTestFile — rama return false cuando la extensión no coincide

func TestIsTestFile_WrongExtension(t *testing.T) {
	// .txt no está en extensions → hasSourceExt = false → return false
	if isTestFile("test_foo.txt", []string{".py"}, []string{"test_"}) {
		t.Error("test_foo.txt should not be a test file for python (wrong extension)")
	}
}

// scanDirectory — rama return err cuando err != nil en Walk callback

func TestScanDirectory_ErrorInSubdir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod test not reliable as root")
	}
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "restricted")
	os.Mkdir(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "foo.go"), []byte("package main"), 0644)
	os.Chmod(subdir, 0000)
	defer os.Chmod(subdir, 0755)

	_, err := scanDirectory(tmpDir, []string{}, []string{".go"})
	if err == nil {
		t.Error("expected error when subdirectory is unreadable")
	}
}
