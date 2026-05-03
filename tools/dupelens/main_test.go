package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMain_CallsRun(t *testing.T) {
	orig := osExit
	defer func() { osExit = orig }()
	var code int
	osExit = func(c int) { code = c }
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"dupelens"}
	main()
	if code != 1 {
		t.Errorf("main() with no args: expected exit 1, got %d", code)
	}
}

func TestRun_NoArgs(t *testing.T) {
	code := run([]string{})
	if code != 1 {
		t.Errorf("run([]) = %d, want 1", code)
	}
}

func TestRun_Version(t *testing.T) {
	code := run([]string{"version"})
	if code != 0 {
		t.Errorf("run([version]) = %d, want 0", code)
	}
}

func TestRun_VersionFlags(t *testing.T) {
	for _, v := range []string{"--version", "-v"} {
		code := run([]string{v})
		if code != 0 {
			t.Errorf("run([%s]) = %d, want 0", v, code)
		}
	}
}

func TestRun_Help(t *testing.T) {
	for _, h := range []string{"help", "--help", "-h"} {
		code := run([]string{h})
		if code != 0 {
			t.Errorf("run([%s]) = %d, want 0", h, code)
		}
	}
}

func TestRun_Unknown(t *testing.T) {
	code := run([]string{"bogus-cmd"})
	if code != 1 {
		t.Errorf("run([bogus-cmd]) = %d, want 1", code)
	}
}

func TestRun_InitCmd(t *testing.T) {
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "dupelens.json")
	code := run([]string{"init", "--output", output})
	if code != 0 {
		t.Errorf("run init = %d, want 0", code)
	}
	if _, err := os.Stat(output); err != nil {
		t.Errorf("init should create file: %v", err)
	}
}

func TestRun_CheckCmd(t *testing.T) {
	code := run([]string{"check", "--dir", "testdata/e2e_fixture"})
	if code != 0 && code != 1 {
		t.Errorf("run check = %d, want 0 or 1", code)
	}
}

func TestRunCheck_ParseError(t *testing.T) {
	code := runCheck([]string{"--not-a-real-flag-xyz"})
	if code != 1 {
		t.Errorf("runCheck with bad flag = %d, want 1", code)
	}
}

func TestRunCheck_DirNotAccessible(t *testing.T) {
	code := runCheck([]string{"--dir", "/nonexistent/path/xyz123abc"})
	if code != 1 {
		t.Errorf("runCheck with nonexistent dir = %d, want 1", code)
	}
}

func TestRunCheck_ScanError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod test not reliable as root")
	}
	tmpDir := t.TempDir()
	// El directorio existe (pasa Stat) pero es ilegible (scan falla)
	os.Chmod(tmpDir, 0000)
	defer os.Chmod(tmpDir, 0755)
	code := runCheck([]string{"--dir", tmpDir})
	if code != 1 {
		t.Errorf("runCheck scan error = %d, want 1", code)
	}
}

func TestRunCheck_ConfigError(t *testing.T) {
	tmpDir := t.TempDir()
	badConfig := filepath.Join(tmpDir, "bad.json")
	os.WriteFile(badConfig, []byte("not json"), 0644)
	code := runCheck([]string{"--config", badConfig, "--dir", "testdata/e2e_fixture"})
	if code != 1 {
		t.Errorf("runCheck with bad config = %d, want 1", code)
	}
}

func TestRunCheck_WithMinTokens(t *testing.T) {
	code := runCheck([]string{"--min-tokens", "30", "--dir", "testdata/e2e_fixture"})
	if code != 0 && code != 1 {
		t.Errorf("runCheck --min-tokens = %d, want 0 or 1", code)
	}
}

func TestRunCheck_FormatJSON(t *testing.T) {
	code := runCheck([]string{"--format", "json", "--dir", "testdata/e2e_fixture"})
	if code != 0 && code != 1 {
		t.Errorf("runCheck --format json = %d, want 0 or 1", code)
	}
}

func TestRunCheck_Verbose(t *testing.T) {
	code := runCheck([]string{"--verbose", "--dir", "testdata/e2e_fixture"})
	if code != 0 && code != 1 {
		t.Errorf("runCheck --verbose = %d, want 0 or 1", code)
	}
}

func TestRunCheck_FailWithViolations(t *testing.T) {
	code := runCheck([]string{"--dir", "testdata/e2e_fixture", "--fail", "--min-tokens", "30"})
	if code != 1 {
		t.Errorf("runCheck --fail with violations = %d, want 1", code)
	}
}

func TestRunInit_ParseError(t *testing.T) {
	code := runInit([]string{"--not-a-real-flag-xyz"})
	if code != 1 {
		t.Errorf("runInit with bad flag = %d, want 1", code)
	}
}

func TestRunInit_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "dupelens.json")
	os.WriteFile(output, []byte("{}"), 0644)
	code := runInit([]string{"--output", output})
	if code != 1 {
		t.Errorf("runInit on existing file = %d, want 1", code)
	}
}

func TestRunInit_WriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod test not reliable as root")
	}
	tmpDir := t.TempDir()
	os.Chmod(tmpDir, 0555)
	defer os.Chmod(tmpDir, 0755)
	output := filepath.Join(tmpDir, "new.json")
	// Stat falla (no existe), WriteFile falla (directorio solo lectura)
	code := runInit([]string{"--output", output})
	if code != 1 {
		t.Errorf("runInit in readonly dir = %d, want 1", code)
	}
}

func TestDefaultConfigJSON(t *testing.T) {
	cfg := defaultConfigJSON()
	if cfg == "" {
		t.Error("defaultConfigJSON should return non-empty string")
	}
	if !containsSubstr(cfg, "minTokens") {
		t.Error("defaultConfigJSON should contain minTokens")
	}
}

func TestMakeSnippetFunc(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "test.txt", "line1\nline2\nline3\n")
	fn := makeSnippetFunc(dir)
	result := fn("test.txt", 1, 2)
	if !containsSubstr(result, "line1") {
		t.Errorf("makeSnippetFunc should return snippet, got %q", result)
	}
}

func TestPrintUsage_Stdout(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printUsage()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	if buf.Len() == 0 {
		t.Error("printUsage should print something")
	}
}
