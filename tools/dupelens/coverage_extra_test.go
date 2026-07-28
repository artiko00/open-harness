package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig_ReadError(t *testing.T) {
	dir := t.TempDir()
	_, err := loadConfig(dir)
	if err == nil {
		t.Error("expected error loading directory as config")
	}
}

func TestLoadConfig_ZeroMinTokensFillsDefault(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cfg.json")
	os.WriteFile(p, []byte(`{"default":{"minLines":7}}`), 0644)
	cfg, err := loadConfig(p)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}
	if cfg.Default.MinTokens != 50 {
		t.Errorf("MinTokens should fill from default=50, got %d", cfg.Default.MinTokens)
	}
}

func TestMergeOnePair_SingleElement(t *testing.T) {
	m := Match{FileA: "a.go", StartLineA: 1, EndLineA: 5, FileB: "b.go", StartLineB: 10, EndLineB: 14, Tokens: 5}
	result := mergeOnePair([]Match{m}, 5)
	if len(result) != 1 || result[0].StartLineA != 1 {
		t.Errorf("single-element mergeOnePair = %v, want unchanged", result)
	}
}

func TestMergeOnePair_NonOverlapping(t *testing.T) {
	matches := []Match{
		{FileA: "a.go", StartLineA: 1, EndLineA: 5, FileB: "b.go", StartLineB: 1, EndLineB: 5, Tokens: 3},
		{FileA: "a.go", StartLineA: 20, EndLineA: 25, FileB: "b.go", StartLineB: 20, EndLineB: 25, Tokens: 3},
	}
	result := mergeOnePair(matches, 3)
	if len(result) != 2 {
		t.Errorf("non-overlapping mergeOnePair = %d matches, want 2", len(result))
	}
}

func TestMergeOnePair_SameStartLineA_SortByB(t *testing.T) {
	matches := []Match{
		{FileA: "a.go", StartLineA: 10, EndLineA: 15, FileB: "b.go", StartLineB: 50, EndLineB: 55, Tokens: 3},
		{FileA: "a.go", StartLineA: 10, EndLineA: 15, FileB: "b.go", StartLineB: 1, EndLineB: 6, Tokens: 3},
	}
	result := mergeOnePair(matches, 3)
	if len(result) != 2 {
		t.Errorf("same StartLineA non-overlapping B: expected 2, got %d", len(result))
	}
}

func TestSortMatches_SameFilePair(t *testing.T) {
	matches := []Match{
		{FileA: "a.go", FileB: "b.go", StartLineA: 20, EndLineA: 25},
		{FileA: "a.go", FileB: "b.go", StartLineA: 5, EndLineA: 10},
	}
	sortMatches(matches)
	if matches[0].StartLineA != 5 {
		t.Errorf("sortMatches same pair: expected StartLineA=5 first, got %d", matches[0].StartLineA)
	}
}

func TestReportConsole_WithColor(t *testing.T) {
	var buf strings.Builder
	reportConsole(sampleMatches(), ReportOpts{ScannedCount: 5, NoColor: false}, &buf)
	if buf.Len() == 0 {
		t.Error("reportConsole with color should produce output")
	}
}

func TestReportConsole_EmptyWithColor(t *testing.T) {
	var buf strings.Builder
	reportConsole(nil, ReportOpts{ScannedCount: 5, NoColor: false}, &buf)
	if !containsSubstr(buf.String(), "OK") {
		t.Error("empty reportConsole with color should show OK")
	}
}

func TestTopDuplicatedFiles_IntraFileMatch(t *testing.T) {
	matches := []Match{
		{FileA: "a.go", FileB: "a.go", StartLineA: 1, EndLineA: 5, StartLineB: 100, EndLineB: 104},
	}
	top := topDuplicatedFiles(matches, 5)
	if len(top) != 1 || top[0].File != "a.go" || top[0].Count != 1 {
		t.Errorf("intra-file topDuplicatedFiles = %v, want [{a.go 1}]", top)
	}
}

func TestTopDuplicatedFiles_MoreThanN(t *testing.T) {
	matches := []Match{
		{FileA: "a.go", FileB: "b.go"}, {FileA: "c.go", FileB: "d.go"},
		{FileA: "e.go", FileB: "f.go"}, {FileA: "g.go", FileB: "h.go"},
		{FileA: "i.go", FileB: "j.go"}, {FileA: "k.go", FileB: "l.go"},
	}
	top := topDuplicatedFiles(matches, 5)
	if len(top) != 5 {
		t.Errorf("topDuplicatedFiles n=5 with 12 unique files: expected 5, got %d", len(top))
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("write error") }

func TestReportJSON_WriteError(t *testing.T) {
	if exit := reportJSON(nil, ReportOpts{}, errorWriter{}); exit != 0 {
		t.Errorf("reportJSON with write error = %d, want 0", exit)
	}
}

func TestScan_ExcludedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	excluded := filepath.Join(tmpDir, "vendor")
	os.MkdirAll(excluded, 0755)
	os.WriteFile(filepath.Join(excluded, "lib.go"), []byte("package foo\nfunc Lib() {}"), 0644)
	cfg := Config{Default: DefaultConfig{MinTokens: 5, MinLines: 1}, Exclude: []string{"vendor"}}
	_, scanned, _, err := scan(tmpDir, cfg, 0)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if scanned != 0 {
		t.Errorf("excluded dir: expected 0 scanned, got %d", scanned)
	}
}

func TestScan_ExcludedFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "vendor"), []byte("not a dir"), 0644)
	cfg := Config{Default: DefaultConfig{MinTokens: 5, MinLines: 1}, Exclude: []string{"vendor"}}
	_, scanned, _, err := scan(tmpDir, cfg, 0)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if scanned != 0 {
		t.Errorf("excluded file: expected 0 scanned, got %d", scanned)
	}
}

func TestScan_BinaryExtension(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "img.jpg"), []byte{0xFF, 0xD8, 0xFF}, 0644)
	cfg := Config{Default: DefaultConfig{MinTokens: 5, MinLines: 1}}
	_, scanned, _, err := scan(tmpDir, cfg, 0)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if scanned != 0 {
		t.Errorf("binary extension: expected 0 scanned, got %d", scanned)
	}
}

func TestScan_SkippedByRule(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "foo_test.go"), []byte("package foo\nfunc TestFoo(t *testing.T) {}"), 0644)
	cfg := Config{
		Default: DefaultConfig{MinTokens: 5, MinLines: 1},
		Rules:   []Rule{{Pattern: "**/*_test.go", Skip: true}},
	}
	_, scanned, _, err := scan(tmpDir, cfg, 0)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if scanned != 0 {
		t.Errorf("skipped by rule: expected 0 scanned, got %d", scanned)
	}
}

func TestScan_RootNotExist(t *testing.T) {
	cfg := Config{Default: DefaultConfig{MinTokens: 5, MinLines: 1}}
	if _, _, _, err := scan("/nonexistent/path/xyz123abc", cfg, 0); err == nil {
		t.Error("scan with nonexistent root should return error")
	}
}

func TestScan_SubdirError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod test not reliable as root")
	}
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "restricted")
	os.Mkdir(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "foo.go"), []byte("package main"), 0644)
	os.Chmod(subdir, 0000)
	defer os.Chmod(subdir, 0755)
	cfg := Config{Default: DefaultConfig{MinTokens: 5, MinLines: 1}}
	if _, _, _, err := scan(tmpDir, cfg, 0); err != nil {
		t.Fatalf("scan should ignore subdir errors, got: %v", err)
	}
}

func TestScan_UnreadableFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod test not reliable as root")
	}
	tmpDir := t.TempDir()
	locked := filepath.Join(tmpDir, "locked.go")
	os.WriteFile(locked, []byte("package main"), 0644)
	os.Chmod(locked, 0000)
	defer os.Chmod(locked, 0644)
	cfg := Config{Default: DefaultConfig{MinTokens: 5, MinLines: 1}}
	if _, _, _, err := scan(tmpDir, cfg, 0); err != nil {
		t.Fatalf("scan with unreadable file should not error: %v", err)
	}
}

func TestScan_FileWithNoTokens(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "tiny.go"), []byte("x"), 0644)
	cfg := Config{Default: DefaultConfig{MinTokens: 50, MinLines: 5}}
	_, scanned, _, err := scan(tmpDir, cfg, 0)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if scanned != 1 {
		t.Errorf("tiny file should be scanned, got scanned=%d", scanned)
	}
}

func TestReadSnippet_ZeroStartLine(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "f.txt", "line1\n")
	if got := readSnippet(dir, "f.txt", 0, 3); got != "" {
		t.Errorf("startLine=0 should return empty, got %q", got)
	}
}
