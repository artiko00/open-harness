package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsBinaryPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"image.jpg", true},
		{"archive.zip", true},
		{"binary.exe", true},
		{"data.db", true},
		{"file.go", false},
		{"script.py", false},
		{"readme.txt", false},
	}
	for _, c := range cases {
		if got := isBinaryPath(c.path); got != c.want {
			t.Errorf("isBinaryPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestIsBinaryContent_TextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "text.go")
	os.WriteFile(path, []byte("package main\nfunc main() {}\n"), 0644)
	if isBinaryContent(path) {
		t.Error("expected false for text file")
	}
}

func TestIsBinaryContent_BinaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.bin")
	os.WriteFile(path, []byte{0x00, 0x01, 0x02, 'h', 'e', 'l', 'l', 'o'}, 0644)
	if !isBinaryContent(path) {
		t.Error("expected true for file with null bytes")
	}
}

func TestIsBinaryContent_NonExistent(t *testing.T) {
	if isBinaryContent("/nonexistent/file.bin") {
		t.Error("expected false for nonexistent file")
	}
}
