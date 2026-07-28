package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copiaFixture copia testdata/<name> al directorio dir con el mismo nombre.
func copiaFixture(t *testing.T, dir, name string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("no se pudo leer fixture %s: %v", name, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func soloViolaciones(results []FileResult) []FileResult {
	var v []FileResult
	for _, r := range results {
		if r.IsViolation() {
			v = append(v, r)
		}
	}
	return v
}

// TAREA 7.2: 382 líneas físicas (380 de licencia + 2 de código) NO viola con maxLines 100.
func TestScan_LicenseHeaderNoViola(t *testing.T) {
	dir := t.TempDir()
	copiaFixture(t, dir, "license_header.go")

	results, _, err := scan(dir, Config{Default: DefaultConfig{MaxLines: 100}, Exclude: []string{}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("esperaba 1 archivo medido, hay %d", len(results))
	}
	r := results[0]
	if r.IsViolation() {
		t.Errorf("cabecera de licencia NO debe violar: code=%d physical=%d", r.Lines, r.Physical)
	}
	if r.Lines != 2 {
		t.Errorf("esperaba 2 líneas de código, got %d", r.Lines)
	}
	if r.Physical != 382 {
		t.Errorf("esperaba 382 líneas físicas, got %d", r.Physical)
	}
}

// TAREA 7.3: 250 líneas de código (sin comentarios) SÍ viola con maxLines 100.
func TestScan_CodigoDensoViola(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "denso.go", lines(250))

	results, _, err := scan(dir, Config{Default: DefaultConfig{MaxLines: 100}, Exclude: []string{}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	v := soloViolaciones(results)
	if len(v) != 1 {
		t.Fatalf("esperaba 1 violación, hay %d", len(v))
	}
	if v[0].Lines != 250 {
		t.Errorf("esperaba 250 líneas de código, got %d", v[0].Lines)
	}
}

// TAREA 7.4: el reporte (consola y JSON) muestra líneas de CÓDIGO y total FÍSICO.
func TestReport_MuestraCodigoYFisico(t *testing.T) {
	// 150 líneas de código + 100 de comentario = 250 físicas.
	var b strings.Builder
	for i := 0; i < 150; i++ {
		b.WriteString("real := 1\n")
	}
	for i := 0; i < 100; i++ {
		b.WriteString("// comentario\n")
	}
	dir := t.TempDir()
	writeFile(t, dir, "mix.go", b.String())

	results, _, err := scan(dir, Config{Default: DefaultConfig{MaxLines: 100}, Exclude: []string{}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	v := soloViolaciones(results)
	if len(v) != 1 || v[0].Lines != 150 || v[0].Physical != 250 {
		t.Fatalf("esperaba code=150 physical=250, got %+v", v)
	}

	// Consola: ambas métricas visibles.
	out := captureStdout(t, func() { reportConsole(results, nil, true) })
	if !strings.Contains(out, "150 code") || !strings.Contains(out, "250 lines") {
		t.Errorf("consola debe mostrar código y físico:\n%s", out)
	}

	// JSON: campos lines y physical.
	var buf bytes.Buffer
	reportJSON(results, nil, &buf)
	var parsed struct {
		Violations []struct {
			Lines    int `json:"lines"`
			Physical int `json:"physical"`
		} `json:"violations"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if len(parsed.Violations) != 1 || parsed.Violations[0].Lines != 150 || parsed.Violations[0].Physical != 250 {
		t.Errorf("JSON debe llevar lines=150 y physical=250, got %+v", parsed.Violations)
	}
}

// TAREA 7.5: i18n.json y schema.sql NO se reportan (no son extensiones de código).
func TestScan_DatosNoSeReportan(t *testing.T) {
	dir := t.TempDir()
	copiaFixture(t, dir, "i18n.json")
	copiaFixture(t, dir, "schema.sql")
	writeFile(t, dir, "app.go", lines(10))

	results, skips, err := scan(dir, Config{Default: DefaultConfig{MaxLines: 100}, Exclude: []string{}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].RelPath != "app.go" {
		t.Fatalf("solo app.go debe medirse, got %+v", results)
	}
	for _, s := range skips {
		if s.Path == "i18n.json" || s.Path == "schema.sql" {
			t.Errorf("los datos no deben aparecer como skip: %s", s.Path)
		}
	}
}

// TAREA 7.6: 83 líneas con 80 anidamientos VIOLA con umbral 5; plano largo NO.
func TestScan_AnidamientoViola(t *testing.T) {
	dir := t.TempDir()
	copiaFixture(t, dir, "nightmare.go")
	writeFile(t, dir, "flat.go", lines(250))

	cfg := Config{Default: DefaultConfig{MaxLines: 1000, MaxNesting: 5}, Exclude: []string{}}
	results, _, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		switch r.RelPath {
		case "nightmare.go":
			if !r.IsViolation() {
				t.Errorf("nightmare.go debe violar por anidamiento (nesting=%d)", r.Nesting)
			}
			if r.Nesting <= 5 {
				t.Errorf("esperaba nesting > 5, got %d", r.Nesting)
			}
		case "flat.go":
			if r.IsViolation() {
				t.Errorf("flat.go plano NO debe violar por anidamiento (nesting=%d)", r.Nesting)
			}
		}
	}
}

// La regla skip también aplica a archivos de código (no solo a datos filtrados).
func TestScan_SkipRuleCodigo(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "gen/api.go", lines(300))

	cfg := Config{
		Default: DefaultConfig{MaxLines: 100},
		Rules:   []Rule{{Pattern: "**/gen/**", Skip: true}},
		Exclude: []string{},
	}
	results, _, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("el .go bajo gen/ debe omitirse por regla skip, got %d", len(results))
	}
}

// Un path de más de 60 caracteres se recorta y la columna se topea.
func TestReport_PathLargoSeRecorta(t *testing.T) {
	long := "src/muy/profundamente/anidado/camino/hacia/un/archivo/que/supera/sesenta.go"
	results := []FileResult{{RelPath: long, Lines: 150, Physical: 150, MaxLines: 100}}
	out := captureStdout(t, func() {
		if n := reportConsole(results, nil, true); n != 1 {
			t.Errorf("violaciones = %d, quiero 1", n)
		}
	})
	if !strings.Contains(out, "...") {
		t.Errorf("el path largo debe recortarse con '...':\n%s", out)
	}
}

// TAREA 7.10: la violación por anidamiento se detalla en consola.
func TestReport_AnidamientoEnConsola(t *testing.T) {
	results := []FileResult{
		{RelPath: "deep.go", Lines: 10, Physical: 12, Nesting: 9, MaxLines: 100, MaxNesting: 5},
	}
	out := captureStdout(t, func() {
		if n := reportConsole(results, nil, false); n != 1 {
			t.Errorf("violaciones = %d, quiero 1", n)
		}
	})
	if !strings.Contains(out, "nesting 9 (max: 5)") {
		t.Errorf("consola debe detallar el anidamiento:\n%s", out)
	}
}
