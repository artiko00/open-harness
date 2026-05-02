package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompilePatterns_InvalidRegex(t *testing.T) {
	rules := []PatternRule{{Name: "bad", Pattern: `[invalid`, Severity: "high"}}
	_, err := compilePatterns(rules)
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestCompilePatterns_Valid(t *testing.T) {
	rules := []PatternRule{
		{Name: "aws", Pattern: `AKIA[0-9A-Z]{16}`, Severity: "critical"},
		{Name: "jwt", Pattern: `eyJ[A-Za-z0-9]+`, Severity: "high"},
	}
	compiled, err := compilePatterns(rules)
	if err != nil {
		t.Fatal(err)
	}
	if len(compiled) != 2 {
		t.Errorf("expected 2 compiled rules, got %d", len(compiled))
	}
}

func TestTruncate_Short(t *testing.T) {
	s := truncate("hello", 10)
	if s != "hello" {
		t.Errorf("expected 'hello', got %q", s)
	}
}

func TestTruncate_Long(t *testing.T) {
	s := truncate("abcdefghijklmnopqrstuvwxyz", 10)
	if len(s) != 10 {
		t.Errorf("expected len 10, got %d", len(s))
	}
	if s[7:] != "..." {
		t.Errorf("expected '...' suffix, got %q", s[7:])
	}
}

func TestIsAllowed_Match(t *testing.T) {
	if !isAllowed("API_KEY=your_key_here", []string{"your_key_here"}) {
		t.Error("expected line to be allowed")
	}
}

func TestIsAllowed_CaseInsensitive(t *testing.T) {
	if !isAllowed("SECRET=CHANGEME", []string{"changeme"}) {
		t.Error("expected case-insensitive match")
	}
}

func TestIsAllowed_NoMatch(t *testing.T) {
	if isAllowed(`key = "AKIAQNB7Q7AIIVSMBGPF"`, []string{"example", "placeholder"}) {
		t.Error("expected line NOT to be allowed")
	}
}

func TestScanFile_Error(t *testing.T) {
	compiled, _ := compilePatterns(defaultPatterns())
	_, err := scanFile("/nonexistent/file.go", "file.go", compiled, nil)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestScanFile_FindsSecrets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.go")
	os.WriteFile(path, []byte(`const key = "AKIAQNB7Q7AIIVSMBGPF"`), 0644)

	compiled, _ := compilePatterns(defaultPatterns())
	findings, err := scanFile(path, "config.go", compiled, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Error("expected findings")
	}
}

func TestScanFile_AllowlistFilters(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "example.env")
	os.WriteFile(path, []byte(`KEY="your_key_here"`), 0644)

	compiled, _ := compilePatterns(defaultPatterns())
	findings, err := scanFile(path, "example.env", compiled, []string{"your_key_here"})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}
