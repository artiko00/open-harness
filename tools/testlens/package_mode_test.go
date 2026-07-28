package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageHasTests_WithTestFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "main_test.go"),
		[]byte("package main\nfunc TestMain(t *testing.T){}"), 0644)
	if !packageHasTests(dir, mapLanguageExtensions()["go"]) {
		t.Error("dir with *_test.go should return true")
	}
}

func TestPackageHasTests_NoTestFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "scanner.go"), []byte("package main"), 0644)
	if packageHasTests(dir, mapLanguageExtensions()["go"]) {
		t.Error("dir with only source files should return false")
	}
}

func TestPackageHasTests_EmptyTestFile(t *testing.T) {
	// Un *_test.go sin marcador (solo 'package svc') NO cubre (8.9, 8.15).
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "svc.go"), []byte("package svc"), 0644)
	os.WriteFile(filepath.Join(dir, "zzz_test.go"), []byte("package svc"), 0644)
	if packageHasTests(dir, mapLanguageExtensions()["go"]) {
		t.Error("test file without markers should NOT count as covering the package")
	}
}

func TestPackageHasTests_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	if packageHasTests(dir, mapLanguageExtensions()["go"]) {
		t.Error("empty dir should return false")
	}
}

func TestPackageHasTests_UnreadableDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod test not reliable as root")
	}
	dir := t.TempDir()
	os.Chmod(dir, 0000)
	defer os.Chmod(dir, 0755)
	if packageHasTests(dir, mapLanguageExtensions()["go"]) {
		t.Error("unreadable dir should return false")
	}
}

func TestCheckCoverage_PackageMode_NoViolation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "scanner.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "scanner_test.go"),
		[]byte("package main\nfunc TestScan(t *testing.T){}"), 0644)
	cfg := config{language: "go", root: dir}
	violations, _, err := checkCoverageCounts(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}
	if violations != 0 {
		t.Errorf("Go package with any test file: expected 0 violations, got %d", violations)
	}
}

func TestCheckCoverage_PackageMode_OneViolation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "scanner.go"), []byte("package main"), 0644)
	cfg := config{language: "go", root: dir}
	violations, _, err := checkCoverageCounts(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}
	if violations != 1 {
		t.Errorf("Go package with no tests: expected 1 violation (package-level), got %d", violations)
	}
}

func TestCheckCoverage_PackageMode_MixedSubdirs(t *testing.T) {
	dir := t.TempDir()
	subA := filepath.Join(dir, "pkga")
	os.MkdirAll(subA, 0755)
	os.WriteFile(filepath.Join(subA, "foo.go"), []byte("package pkga"), 0644)
	os.WriteFile(filepath.Join(subA, "foo_test.go"),
		[]byte("package pkga\nfunc TestFoo(t *testing.T){}"), 0644)
	subB := filepath.Join(dir, "pkgb")
	os.MkdirAll(subB, 0755)
	os.WriteFile(filepath.Join(subB, "bar.go"), []byte("package pkgb"), 0644)
	os.WriteFile(filepath.Join(subB, "baz.go"), []byte("package pkgb"), 0644)
	cfg := config{language: "go", root: dir}
	violations, _, err := checkCoverageCounts(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}
	if violations != 1 {
		t.Errorf("mixed subdirs: expected 1 violation (pkgb), got %d", violations)
	}
}

func TestCheckCoverage_Python_PerFileMode(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "foo.py"), []byte("x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "bar.py"), []byte("y = 2"), 0644)
	cfg := config{language: "python", root: dir}
	violations, _, err := checkCoverageCounts(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}
	if violations != 2 {
		t.Errorf("Python per-file mode: expected 2 violations (one per file), got %d", violations)
	}
}

func TestCheckCoverage_PerFile_SkipsNodeModules(t *testing.T) {
	tmpDir := t.TempDir()
	nm := filepath.Join(tmpDir, "node_modules")
	os.MkdirAll(nm, 0755)
	os.WriteFile(filepath.Join(nm, "foo.py"), []byte("x = 1"), 0644)
	cfg := config{language: "python", root: tmpDir}
	violations, _, err := checkCoverageCounts(cfg)
	if err != nil {
		t.Fatalf("checkCoverage failed: %v", err)
	}
	if violations != 0 {
		t.Errorf("node_modules should be skipped in per-file mode, got %d violations", violations)
	}
}

func TestCheckCoverage_PerFile_WalkError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod test not reliable as root")
	}
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "restricted")
	os.Mkdir(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "foo.py"), []byte("x = 1"), 0644)
	os.Chmod(subdir, 0000)
	defer os.Chmod(subdir, 0755)
	cfg := config{language: "python", root: tmpDir}
	if _, _, err := checkCoverageCounts(cfg); err == nil {
		t.Error("checkCoverage per-file should return error for unreadable subdir")
	}
}

func TestLanguageMapping_GoIsPackageBased(t *testing.T) {
	m := mapLanguageExtensions()
	if !m["go"].packageBased {
		t.Error("go language mapping should be packageBased=true")
	}
}

func TestLanguageMapping_PythonIsNotPackageBased(t *testing.T) {
	m := mapLanguageExtensions()
	if m["python"].packageBased {
		t.Error("python language mapping should be packageBased=false")
	}
}
