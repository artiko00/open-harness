package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViolationLine(t *testing.T) {
	plain := violationLine("foo.go", "- no test found", true)
	if strings.Contains(plain, "\033") {
		t.Errorf("noColor line must not contain ANSI: %q", plain)
	}
	if !strings.Contains(plain, "foo.go - no test found") {
		t.Errorf("plain line missing content: %q", plain)
	}

	colored := violationLine("foo.go", "- no test found", false)
	if !strings.Contains(colored, "\033") {
		t.Errorf("colored line must contain ANSI: %q", colored)
	}
}

func TestSummaryLine(t *testing.T) {
	cases := []struct {
		violations int
		noColor    bool
		wantANSI   bool
		want       string
	}{
		{0, true, false, "All source files have tests"},
		{0, false, true, "All source files have tests"},
		{3, true, false, "3 file(s) without tests"},
		{3, false, true, "3 file(s) without tests"},
	}
	for _, c := range cases {
		got := summaryLine(c.violations, c.noColor)
		if !strings.Contains(got, c.want) {
			t.Errorf("summaryLine(%d,%v) = %q, want substring %q", c.violations, c.noColor, got, c.want)
		}
		if strings.Contains(got, "\033") != c.wantANSI {
			t.Errorf("summaryLine(%d,%v) ANSI presence = %v, want %v", c.violations, c.noColor, !c.wantANSI, c.wantANSI)
		}
	}
}

// captureStdout runs fn and returns everything it wrote to os.Stdout.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestRunCheckNoColorOutput(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "foo.ts"), []byte("export const x = 1"), 0644)

	out := captureStdout(func() {
		runCheck([]string{"--lang", "typescript", "--dir", tmpDir, "--no-color"})
	})

	if strings.Contains(out, "\033") {
		t.Errorf("--no-color output must not contain ANSI escapes: %q", out)
	}
	if !strings.Contains(out, "foo.ts - no test found") {
		t.Errorf("expected violation line in output: %q", out)
	}
}

func TestRunCheckColorOutput(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "foo.ts"), []byte("export const x = 1"), 0644)

	out := captureStdout(func() {
		runCheck([]string{"--lang", "typescript", "--dir", tmpDir})
	})

	if !strings.Contains(out, "\033") {
		t.Errorf("default output must contain ANSI escapes: %q", out)
	}
}
