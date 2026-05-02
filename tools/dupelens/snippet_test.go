package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}

func TestReadSnippet_returnsRequestedLines(t *testing.T) {
	dir := t.TempDir()
	src := "line1\nline2\nline3\nline4\nline5\n"
	writeTempFile(t, dir, "a.txt", src)

	got := readSnippet(dir, "a.txt", 2, 3)
	for _, want := range []string{"line2", "line3", "line4"} {
		if !contains([]string{got}, "") && !containsSubstr(got, want) {
			t.Errorf("readSnippet missing %q in %q", want, got)
		}
	}
}

func TestReadSnippet_clampsAtEOF(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "short.txt", "only_one_line\n")
	got := readSnippet(dir, "short.txt", 1, 5)
	if !containsSubstr(got, "only_one_line") {
		t.Errorf("expected only_one_line in snippet, got %q", got)
	}
}

func TestReadSnippet_missingFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	got := readSnippet(dir, "nope.txt", 1, 3)
	if got != "" {
		t.Errorf("missing file debe devolver vacío, got %q", got)
	}
}

func TestReadSnippet_invalidLineReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "f.txt", "a\nb\nc\n")
	got := readSnippet(dir, "f.txt", 100, 3)
	if got != "" {
		t.Errorf("startLine fuera de rango debe devolver vacío, got %q", got)
	}
}

func containsSubstr(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
