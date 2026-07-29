package main

import (
	"bytes"
	"strings"
	"testing"
)

func sampleRes() scanResult {
	return scanResult{
		Branch:    "feature/x",
		Base:      "origin/main",
		MergeBase: "a1b2c3d",
		Files: []string{
			"go.sum", "pkg/a.go", "pkg/a_test.go", "pnpm-lock.yaml", "src/b.ts",
		},
	}
}

func TestBuildReportClasificaYCuenta(t *testing.T) {
	r := buildReport(sampleRes(), defaultExclude, 15, 0, "or", "changed", false, false)
	if len(r.Source) != 2 || len(r.Test) != 1 || len(r.Excluded) != 2 {
		t.Fatalf("clasificación: src=%v test=%v exc=%v", r.Source, r.Test, r.Excluded)
	}
	if r.Countable != 3 {
		t.Fatalf("contable = %d, want 3", r.Countable)
	}
}

func TestBuildReportExcludeTests(t *testing.T) {
	r := buildReport(sampleRes(), defaultExclude, 15, 0, "or", "changed", true, false)
	if r.Countable != 2 {
		t.Fatalf("con --exclude-tests contable = %d, want 2", r.Countable)
	}
}

func TestPrintReportNoColor(t *testing.T) {
	r := buildReport(sampleRes(), defaultExclude, 2, 0, "or", "changed", false, false)
	var buf bytes.Buffer
	printReport(r, true, &buf)
	out := buf.String()
	if strings.Contains(out, "\033[") {
		t.Fatal("no-color no debe emitir ANSI")
	}
	for _, want := range []string{
		"feature/x", "origin/main", "a1b2c3d", "FAIL: 3 files (max 2)",
		"source (2)", "test (1)", "excluded (2)", "pnpm-lock.yaml", "lockfile",
		"SUMMARY: 3 counted, 2 excluded, limit 2",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("falta %q en:\n%s", want, out)
		}
	}
}

func TestPrintReportOKColor(t *testing.T) {
	r := buildReport(sampleRes(), defaultExclude, 15, 0, "or", "changed", false, false)
	var buf bytes.Buffer
	printReport(r, false, &buf)
	out := buf.String()
	if !strings.Contains(out, colorGreen) || !strings.Contains(out, "OK") {
		t.Fatalf("esperaba OK en verde:\n%s", out)
	}
}

func TestPrintReportStagedHeader(t *testing.T) {
	res := scanResult{Branch: "main", Files: []string{"a.go"}}
	r := buildReport(res, defaultExclude, 15, 0, "or", "changed", false, true)
	var buf bytes.Buffer
	printReport(r, true, &buf)
	if !strings.Contains(buf.String(), "main (sólo staged)") {
		t.Fatalf("encabezado staged:\n%s", buf.String())
	}
}

func TestPrintReportSinArchivos(t *testing.T) {
	res := scanResult{Branch: "main", MergeBase: "abc"}
	r := buildReport(res, defaultExclude, 15, 0, "or", "changed", false, false)
	var buf bytes.Buffer
	printReport(r, true, &buf)
	out := buf.String()
	if strings.Contains(out, "source (") || strings.Contains(out, "excluded (") {
		t.Fatalf("categorías vacías no deben imprimirse:\n%s", out)
	}
}

func TestDeterminismo20Corridas(t *testing.T) {
	r := buildReport(sampleRes(), defaultExclude, 2, 0, "or", "changed", false, false)
	var first string
	for i := 0; i < 20; i++ {
		var buf bytes.Buffer
		printReport(r, true, &buf)
		if i == 0 {
			first = buf.String()
			continue
		}
		if buf.String() != first {
			t.Fatalf("corrida %d difiere:\n%s\nvs\n%s", i, buf.String(), first)
		}
	}
}
