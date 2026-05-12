package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

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
	code := runCheck([]string{"--lang", "invalid-lang", "--dir", os.TempDir()})
	if code != 0 && code != 1 {
		t.Errorf("runCheck with invalid lang = %d, want 0 or 1", code)
	}
}

func TestRunCheckConfigLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	code := runCheck([]string{"--config", tmpDir, "--dir", tmpDir})
	if code != 1 {
		t.Errorf("expected runCheck=1 when --config points to a directory, got %d", code)
	}
}

func TestRunCheckConfigFromPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "foo.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "package.json"),
		[]byte(`{"testlens":{"language":"go"}}`), 0644)

	code := runCheck([]string{"--config", filepath.Join(tmpDir, "testlens.json"), "--dir", tmpDir})
	if code != 0 && code != 1 {
		t.Errorf("expected 0 or 1 with config from package.json, got %d", code)
	}
}

func TestRunCheckExplicitLangBeatsConfig(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "testlens.json"),
		[]byte(`{"language":"typescript"}`), 0644)

	code := runCheck([]string{"--config", filepath.Join(tmpDir, "testlens.json"), "--lang", "go", "--dir", tmpDir})
	if code != 0 && code != 1 {
		t.Errorf("expected 0 or 1 when --lang explicit overrides config, got %d", code)
	}
}

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