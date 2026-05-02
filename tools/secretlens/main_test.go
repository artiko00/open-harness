package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain_CallsRun(t *testing.T) {
	orig := osExit
	var code int
	osExit = func(c int) { code = c }
	defer func() { osExit = orig }()

	origArgs := os.Args
	os.Args = []string{"secretlens"}
	defer func() { os.Args = origArgs }()

	main()
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRun_NoArgs(t *testing.T) {
	if code := run([]string{}); code != 1 {
		t.Errorf("expected 1, got %d", code)
	}
}

func TestRun_Version(t *testing.T) {
	for _, arg := range []string{"version", "--version", "-v"} {
		if code := run([]string{arg}); code != 0 {
			t.Errorf("run(%q) = %d, want 0", arg, code)
		}
	}
}

func TestRun_Help(t *testing.T) {
	for _, arg := range []string{"help", "--help", "-h"} {
		if code := run([]string{arg}); code != 0 {
			t.Errorf("run(%q) = %d, want 0", arg, code)
		}
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	if code := run([]string{"foobar"}); code != 1 {
		t.Errorf("expected 1, got %d", code)
	}
}

func TestRun_Check(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "clean.go"), []byte("package main\n"), 0644)
	if code := run([]string{"check", "--dir", dir, "--no-color"}); code != 0 {
		t.Errorf("expected 0, got %d", code)
	}
}

func TestRun_Init(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "secretlens.json")
	if code := run([]string{"init", "--output", output}); code != 0 {
		t.Errorf("expected 0, got %d", code)
	}
}

func TestRunCheck_BadFlag(t *testing.T) {
	if code := runCheck([]string{"--notaflag"}); code != 1 {
		t.Errorf("expected 1 for bad flag, got %d", code)
	}
}

func TestRunCheck_BadConfigJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.json")
	os.WriteFile(cfgPath, []byte("{invalid json"), 0644)
	if code := runCheck([]string{"--config", cfgPath, "--dir", dir}); code != 1 {
		t.Errorf("expected 1 for bad config JSON, got %d", code)
	}
}

func TestRunCheck_ScanError(t *testing.T) {
	if code := runCheck([]string{"--dir", "/nonexistent/path/abc123xyz"}); code != 1 {
		t.Errorf("expected 1 for nonexistent dir, got %d", code)
	}
}

func TestRunCheck_NoFindings(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "clean.go"), []byte("package main\nfunc main() {}\n"), 0644)
	if code := runCheck([]string{"--dir", dir, "--fail", "--no-color"}); code != 0 {
		t.Errorf("expected 0 for clean file, got %d", code)
	}
}

func TestRunCheck_WithFindings(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.go"), []byte(`const key = "AKIAQNB7Q7AIIVSMBGPF"`), 0644)
	if code := runCheck([]string{"--dir", dir, "--fail", "--no-color"}); code != 1 {
		t.Errorf("expected 1 for findings with --fail, got %d", code)
	}
}

func TestRunInit_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "secretlens.json")
	if code := runInit([]string{"--output", output}); code != 0 {
		t.Errorf("expected 0, got %d", code)
	}
	if _, err := os.Stat(output); err != nil {
		t.Errorf("expected file to be created: %v", err)
	}
}

func TestRunInit_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "secretlens.json")
	os.WriteFile(output, []byte("{}"), 0644)
	if code := runInit([]string{"--output", output}); code != 1 {
		t.Errorf("expected 1 when file exists, got %d", code)
	}
}

func TestRunInit_WriteError(t *testing.T) {
	if code := runInit([]string{"--output", "/nonexistent/dir/secretlens.json"}); code != 1 {
		t.Errorf("expected 1 for write error, got %d", code)
	}
}

func TestRunInit_BadFlag(t *testing.T) {
	if code := runInit([]string{"--notaflag"}); code != 1 {
		t.Errorf("expected 1 for bad flag, got %d", code)
	}
}
