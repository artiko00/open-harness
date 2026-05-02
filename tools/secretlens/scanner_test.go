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
	findings, err := scan(dir, defaultConfig)
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
	findings, err := scan(dir, defaultConfig)
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
	findings, err := scan(dir, defaultConfig)
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
	findings, err := scan(dir, defaultConfig)
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
	findings, err := scan(dir, defaultConfig)
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
	findings, err := scan(dir, defaultConfig)
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
	findings, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings in clean file, got %d", len(findings))
	}
}
