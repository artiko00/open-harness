package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestParseNumstat(t *testing.T) {
	out := []byte("10\t2\tsrc/a.go\n" + // normal
		"-\t-\tbin.png\n" + // binario → 0/0
		"5\t3\t{old => new}/x.go\n" + // rename con prefijo
		"4\t1\told.go => dst/new.go\n" + // rename simple
		"7\t0\t\"tab\\ttab.go\"\n" + // path citado (octal/escape)
		"malformada\n") // sin tabs → se ignora
	m := parseNumstat(out)
	if m["src/a.go"] != (lineStat{10, 2}) {
		t.Fatalf("src/a.go = %+v", m["src/a.go"])
	}
	if m["bin.png"] != (lineStat{0, 0}) {
		t.Fatalf("binario = %+v", m["bin.png"])
	}
	if m["new/x.go"] != (lineStat{5, 3}) {
		t.Fatalf("rename prefijo = %+v", m["new/x.go"])
	}
	if m["dst/new.go"] != (lineStat{4, 1}) {
		t.Fatalf("rename simple = %+v", m["dst/new.go"])
	}
	if m["tab\ttab.go"] != (lineStat{7, 0}) {
		t.Fatalf("path citado = %+v", m)
	}
	if len(m) != 5 {
		t.Fatalf("líneas malformadas deben ignorarse: %+v", m)
	}
}

func TestParseNumstatAcumula(t *testing.T) {
	// El mismo archivo en dos filas acumula.
	m := parseNumstat([]byte("3\t1\ta.go\n2\t4\ta.go\n"))
	if m["a.go"] != (lineStat{5, 5}) {
		t.Fatalf("acumulación = %+v", m["a.go"])
	}
}

func TestAtoiOrZero(t *testing.T) {
	if atoiOrZero("42") != 42 || atoiOrZero("-") != 0 || atoiOrZero("x") != 0 {
		t.Fatal("atoiOrZero")
	}
}

func TestCountLinesFor(t *testing.T) {
	churn := map[string]lineStat{"a.go": {10, 4}, "b.go": {2, 8}}
	files := []string{"a.go", "b.go", "ausente.go"}
	if got := countLinesFor(churn, files, "changed"); got != 24 {
		t.Fatalf("changed = %d, want 24", got)
	}
	if got := countLinesFor(churn, files, "added"); got != 12 {
		t.Fatalf("added = %d, want 12", got)
	}
}

func TestExceededSoloArchivos(t *testing.T) {
	// maxLines 0: sólo cuenta el presupuesto de archivos (retrocompat).
	over := report{Countable: 16, Max: 15, MaxLines: 0}
	under := report{Countable: 10, Max: 15, MaxLines: 0, Lines: 99999}
	if !over.exceeded() || under.exceeded() {
		t.Fatal("con maxLines 0 sólo importan los archivos")
	}
}

func TestExceededOr(t *testing.T) {
	base := func(f, l int) report {
		return report{Countable: f, Max: 15, Lines: l, MaxLines: 100, Mode: "or"}
	}
	if base(20, 10).exceeded() != true {
		t.Fatal("or: archivos excedidos")
	}
	if base(5, 200).exceeded() != true {
		t.Fatal("or: líneas excedidas")
	}
	if base(5, 10).exceeded() != false {
		t.Fatal("or: ninguno excedido")
	}
}

func TestExceededAnd(t *testing.T) {
	base := func(f, l int) report {
		return report{Countable: f, Max: 15, Lines: l, MaxLines: 100, Mode: "and"}
	}
	if base(20, 200).exceeded() != true {
		t.Fatal("and: ambos excedidos")
	}
	if base(20, 10).exceeded() != false {
		t.Fatal("and: sólo archivos no basta")
	}
	if base(5, 200).exceeded() != false {
		t.Fatal("and: sólo líneas no basta")
	}
}

func TestValidateFlags(t *testing.T) {
	dir := t.TempDir()
	if validateFlags(15, 0, "or", "changed", dir) != 0 {
		t.Fatal("flags válidos → 0")
	}
	if validateFlags(15, -1, "or", "changed", dir) != 2 {
		t.Fatal("max-lines negativo → 2")
	}
	if validateFlags(15, 0, "xor", "changed", dir) != 2 {
		t.Fatal("mode inválido → 2")
	}
	if validateFlags(15, 0, "or", "raro", dir) != 2 {
		t.Fatal("line-metric inválido → 2")
	}
}

func TestValidateAndDefaultLineas(t *testing.T) {
	if _, err := validateAndDefault(Config{MaxLines: -1}, "x"); err == nil {
		t.Fatal("maxLines negativo debe fallar")
	}
	if _, err := validateAndDefault(Config{Mode: "xor"}, "x"); err == nil {
		t.Fatal("mode inválido debe fallar")
	}
	if _, err := validateAndDefault(Config{LineMetric: "raro"}, "x"); err == nil {
		t.Fatal("lineMetric inválido debe fallar")
	}
	cfg, err := validateAndDefault(Config{MaxLines: 800, Mode: "and", LineMetric: "added"}, "x")
	if err != nil || cfg.Mode != "and" || cfg.LineMetric != "added" || cfg.MaxLines != 800 {
		t.Fatalf("valores válidos preservados: %+v err=%v", cfg, err)
	}
	def, _ := validateAndDefault(Config{}, "x")
	if def.Mode != "or" || def.LineMetric != "changed" {
		t.Fatalf("defaults mode/lineMetric: %+v", def)
	}
}

func TestBuildReportCuentaLineas(t *testing.T) {
	res := scanResult{
		Files: []string{"a.go", "a_test.go"},
		Churn: map[string]lineStat{"a.go": {100, 20}, "a_test.go": {30, 10}},
	}
	r := buildReport(res, nil, 15, 500, "or", "changed", false, false)
	if r.Lines != 160 { // (100+20)+(30+10)
		t.Fatalf("changed incluye tests = %d, want 160", r.Lines)
	}
	r2 := buildReport(res, nil, 15, 500, "or", "added", true, false)
	if r2.Lines != 100 { // sólo source, sólo added
		t.Fatalf("added exclude-tests = %d, want 100", r2.Lines)
	}
}

func TestPrintReportConLineas(t *testing.T) {
	res := scanResult{Branch: "f", Files: []string{"a.go"}, Churn: map[string]lineStat{"a.go": {200, 0}}}
	r := buildReport(res, nil, 15, 100, "or", "changed", false, false)
	var b strings.Builder
	printReport(r, true, &b)
	out := b.String()
	if !strings.Contains(out, "1 files (max 15), 200 lines (max 100)") {
		t.Fatalf("estado con líneas:\n%s", out)
	}
	if !strings.Contains(out, "SUMMARY: 1 files, 200 lines counted, 0 excluded; limit 15 files / 100 lines (or)") {
		t.Fatalf("SUMMARY con líneas:\n%s", out)
	}
	if !strings.Contains(out, "FAIL") {
		t.Fatalf("200 líneas > 100 debe ser FAIL:\n%s", out)
	}
}

// TestCheckUnArchivoMuchasLineas es el caso motivador: 1 archivo con 11000
// líneas excede el presupuesto de líneas aunque esté dentro del de archivos.
func TestCheckUnArchivoMuchasLineas(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "scopelens.json", `{"maxFiles":15,"maxLines":500}`)
	f := &fakeGit{
		refsOK:       map[string]bool{"HEAD": true, "main": true},
		committed:    &resp{out: "huge.go\n"},
		committedNum: &resp{out: "11000\t0\thuge.go\n"},
	}
	code, out := checkWith(t, f, "--dir", dir, "--fail", "--no-color")
	if code != 1 {
		t.Fatalf("1 archivo/11000 líneas code=%d, want 1\n%s", code, out)
	}
	if !strings.Contains(out, "11000 lines (max 500)") {
		t.Fatalf("estado con líneas:\n%s", out)
	}
}

func TestCheckFlagsLineas(t *testing.T) {
	// Flags sobreescriben la config; mode=and exige exceder ambos: 3 archivos
	// (<15) con 11000 líneas (>500) NO debe fallar en modo and.
	dir := t.TempDir()
	f := &fakeGit{
		refsOK:       map[string]bool{"HEAD": true, "main": true},
		committed:    &resp{out: "a.go\nb.go\nc.go\n"},
		committedNum: &resp{out: "11000\t0\ta.go\n"},
	}
	code, out := checkWith(t, f, "--dir", dir, "--max-lines", "500", "--mode", "and", "--line-metric", "added", "--fail", "--no-color")
	if code != 0 {
		t.Fatalf("and con sólo líneas excedidas code=%d, want 0\n%s", code, out)
	}
}

func TestCheckModeInvalido(t *testing.T) {
	if code, _ := checkWith(t, fullRepo("", ""), "--dir", t.TempDir(), "--mode", "xor"); code != 2 {
		t.Fatalf("mode inválido code=%d, want 2", code)
	}
}

func TestMeasureChurnStagedError(t *testing.T) {
	// El numstat del índice falla → measure propaga el error (exit 2).
	f := &fakeGit{
		refsOK:    map[string]bool{"HEAD": true, "main": true},
		staged:    &resp{out: "a.go\n"},
		stagedNum: &resp{err: errors.New("boom")},
	}
	if _, err := measure(f.runner(), "", "", false); err == nil {
		t.Fatal("error de numstat staged debe propagarse")
	}
}

func TestAccumulateChurnError(t *testing.T) {
	run := func(context.Context, ...string) ([]byte, error) { return nil, errors.New("boom") }
	if err := accumulateChurn(context.Background(), run, "--cached", map[string]lineStat{}); err == nil {
		t.Fatal("error de git debe propagarse")
	}
}
