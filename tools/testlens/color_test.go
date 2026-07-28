package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// skipsNoGate arma n omisiones que NO rompen el gate (binarios).
func skipsNoGate(n int) []pathmatch.Skip {
	var s []pathmatch.Skip
	for i := 0; i < n; i++ {
		s = append(s, pathmatch.Skip{Path: "b", Reason: pathmatch.ReasonBinary})
	}
	return s
}

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
		skipped    int
		noColor    bool
		wantANSI   bool
		want       string
	}{
		{0, 0, true, false, "OK: all source files have tests"},
		{0, 0, false, true, "OK: all source files have tests"},
		{3, 0, true, false, "SUMMARY: 3 file(s) without tests"},
		{3, 0, false, true, "SUMMARY: 3 file(s) without tests"},
		// El resumen debe informar los archivos omitidos (tarea 2.9).
		{0, 1, true, false, "OK: all source files have tests (1 skipped)"},
		{0, 2, false, true, "(2 skipped)"},
		{3, 1, true, false, "SUMMARY: 3 file(s) without tests, 1 skipped"},
		{3, 2, false, true, ", 2 skipped"},
	}
	for _, c := range cases {
		got := summaryLine(c.violations, skipsNoGate(c.skipped), c.noColor)
		if !strings.Contains(got, c.want) {
			t.Errorf("summaryLine(%d,%d,%v) = %q, want substring %q", c.violations, c.skipped, c.noColor, got, c.want)
		}
		if strings.Contains(got, "\033") != c.wantANSI {
			t.Errorf("summaryLine(%d,%d,%v) ANSI presence = %v, want %v", c.violations, c.skipped, c.noColor, !c.wantANSI, c.wantANSI)
		}
	}
}

// Con 0 violaciones pero un skip que rompe el gate, la línea es SUMMARY, no OK.
func TestSummaryLine_GateRotoSinViolaciones(t *testing.T) {
	skips := []pathmatch.Skip{{Path: "x", Reason: pathmatch.ReasonReadError}}
	got := summaryLine(0, skips, true)
	if !strings.HasPrefix(got, "SUMMARY:") {
		t.Errorf("gate roto debe dar SUMMARY, got %q", got)
	}
	if strings.Contains(got, "OK:") {
		t.Errorf("no debe contener OK: cuando el gate se rompe, got %q", got)
	}
}

func TestUntestedHeader(t *testing.T) {
	plain := untestedHeader(2, true)
	if plain != "UNTESTED (2 file(s) without tests):" {
		t.Errorf("header sin color = %q", plain)
	}
	if strings.Contains(plain, "\033") {
		t.Errorf("header noColor no debe llevar ANSI: %q", plain)
	}
	if !strings.Contains(untestedHeader(2, false), "\033") {
		t.Error("header con color debe llevar ANSI")
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
