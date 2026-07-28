package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestScan_DetectsAWSKey(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "config.go", `
package main
const key = "AKIAQNB7Q7AIIVSMBGPF"
`)
	findings, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Error("expected AWS key finding, got none")
	}
	if findings[0].Severity != "critical" {
		t.Errorf("expected critical, got %s", findings[0].Severity)
	}
}

func TestScan_DetectsGenericSecret(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "app.py", `
DB_PASSWORD = "supersecretpassword123"
`)
	findings, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Error("expected generic secret finding, got none")
	}
}

func TestScan_AllowlistSkipsPlaceholders(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "example.env", `
API_KEY="your_key_here"
SECRET="changeme"
`)
	findings, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for placeholders, got %d", len(findings))
	}
}

func TestScan_SkipsExcludedDirs(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "node_modules/lib/config.js", `
const token = "ghp_abcdefghijklmnopqrstuvwxyzABCDEFGHIJ"
`)
	findings, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings in node_modules, got %d", len(findings))
	}
}

func TestScan_DetectsPEMKey(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "keys/id_rsa", `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA1234...
-----END RSA PRIVATE KEY-----
`)
	findings, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Error("expected PEM key finding, got none")
	}
	if findings[0].Severity != "critical" {
		t.Errorf("expected critical severity, got %s", findings[0].Severity)
	}
}

func TestScan_DetectsGitHubToken(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "ci.yml", `
env:
  TOKEN: ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij
`)
	findings, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Error("expected GitHub token finding, got none")
	}
}

func TestScan_CleanFileNoFindings(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", `
package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
`)
	findings, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings in clean file, got %d", len(findings))
	}
}

func TestScan_RootNotExist(t *testing.T) {
	_, _, _, err := scan("/nonexistent/path/abc123", defaultConfig)
	if err == nil {
		t.Error("expected error for nonexistent root")
	}
}

func TestScan_InvalidPatterns(t *testing.T) {
	dir := t.TempDir()
	badCfg := Config{
		Patterns:  []PatternRule{{Name: "bad", Pattern: `[invalid`, Severity: "high"}},
		Allowlist: []string{},
		Exclude:   []string{},
	}
	_, _, _, err := scan(dir, badCfg)
	if err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestScan_SkipsBinaryPath(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "image.jpg", "not really a jpg")
	writeTestFile(t, dir, "clean.go", "package main\n")

	findings, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range findings {
		if f.RelPath == "image.jpg" {
			t.Error("jpg should be skipped as binary path")
		}
	}
}

func TestScan_SkipsBinaryContent(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "disguised.go", string([]byte{0x00, 'h', 'i'}))
	writeTestFile(t, dir, "clean.go", "package main\n")

	findings, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range findings {
		if f.RelPath == "disguised.go" {
			t.Error("file with null bytes should be skipped")
		}
	}
}

func TestScan_SkipsExcludedFile(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "yarn.lock", "lock content")
	writeTestFile(t, dir, "clean.go", "package main\n")

	cfg := Config{
		Patterns:  defaultPatterns(),
		Allowlist: defaultConfig.Allowlist,
		Exclude:   []string{"*.lock"},
	}
	findings, _, _, err := scan(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range findings {
		if f.RelPath == "yarn.lock" {
			t.Error("lock file should be excluded")
		}
	}
}

func TestScan_ErrorInSubdir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "restricted")
	os.Mkdir(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "file.go"), []byte("package main"), 0644)
	os.Chmod(subdir, 0000)
	defer os.Chmod(subdir, 0755)

	_, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
}

func TestScan_ScanFileError(t *testing.T) {
	dir := t.TempDir()
	locked := filepath.Join(dir, "locked.go")
	writeTestFile(t, dir, "locked.go", `const key = "AKIAQNB7Q7AIIVSMBGPF"`)
	writeTestFile(t, dir, "clean.go", "package main\n")
	os.Chmod(locked, 0000)
	defer os.Chmod(locked, 0644)

	_, _, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
}
