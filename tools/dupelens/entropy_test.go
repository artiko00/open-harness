package main

import (
	"fmt"
	"strings"
	"testing"
)

// toks arma tokens con una línea por grupo separado por "|".
func toks(spec string) []Token {
	var out []Token
	for i, line := range strings.Split(spec, "|") {
		for _, v := range strings.Fields(line) {
			out = append(out, Token{Value: v, Line: i + 1})
		}
	}
	return out
}

func TestLowEntropyWindow_repetitiveBlockIsLowEntropy(t *testing.T) {
	tk := toks("id name|id name|id name|id name|id name")
	if !lowEntropyWindow(tk, 0, len(tk)) {
		t.Error("un bloque donde todas las líneas empiezan igual es de baja entropía")
	}
}

func TestLowEntropyWindow_realCodeIsNotLowEntropy(t *testing.T) {
	tk := toks("if err|return nil|for row|acc +=|return acc")
	if lowEntropyWindow(tk, 0, len(tk)) {
		t.Error("código con líneas de formas distintas no es de baja entropía")
	}
}

func TestLowEntropyWindow_belowMinimumLinesIsNotJudged(t *testing.T) {
	tk := toks("id name|id name")
	if lowEntropyWindow(tk, 0, len(tk)) {
		t.Errorf("con menos de %d líneas la ventana no se juzga", minEntropyLines)
	}
}

func TestLowEntropyWindow_windowStartingMidLineSkipsThePartialLine(t *testing.T) {
	// La ventana arranca en el segundo token de la primera línea: esa línea está
	// cortada y su token no debe usarse como ancla.
	tk := toks("head tail|id name|id name|id name|id name")
	if !lowEntropyWindow(tk, 1, len(tk)-1) {
		t.Error("la línea inicial parcial debe ignorarse al elegir el ancla")
	}
}

func TestLowEntropyWindow_windowFullyInsideOneLineIsNotJudged(t *testing.T) {
	tk := toks("alpha beta gamma delta")
	if lowEntropyWindow(tk, 1, 2) {
		t.Error("una ventana contenida en una sola línea parcial no se juzga")
	}
}

func TestLowEntropyWindow_singleLineWindowIsNotJudged(t *testing.T) {
	tk := toks("id name value other")
	if lowEntropyWindow(tk, 0, len(tk)) {
		t.Error("una ventana de una sola línea no se juzga")
	}
}

func TestLowEntropyWindow_minorityRepetitionIsNotEnough(t *testing.T) {
	tk := toks("id a|id b|const c|return d|for e")
	if lowEntropyWindow(tk, 0, len(tk)) {
		t.Error("2 de 5 líneas iguales no alcanza el umbral")
	}
}

// dataArray genera un array de objetos literales: una entrada por línea, todas
// con la misma forma y la misma clave inicial. Es el patrón de los seeds y
// catálogos embebidos en código.
func dataArray(name string, offset int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "export const %s = [\n", name)
	for i := 0; i < 30; i++ {
		fmt.Fprintf(&b, "  { id: %d, label: 'row%d', weight: %d, active: true },\n",
			i+offset, i, (i+offset)*7)
	}
	b.WriteString("];\n")
	return b.String()
}

func TestScan_dataArraysDoNotProduceRenamedMatches(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "seed_a.ts", dataArray("alphaSeed", 0))
	writeTempFile(t, dir, "seed_b.ts", dataArray("betaSeed", 100))

	cfg := defaultConfig
	cfg.Default.MinTokens = 20
	matches, _, _, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	for _, m := range matches {
		if m.Kind == "renamed" {
			t.Fatalf("un array de datos no debe producir clones renombrados: %+v", m)
		}
	}
}

// literalSwitch genera un switch con veinte case: bloque repetitivo, pero cuya
// copia literal sí es un hallazgo legítimo.
func literalSwitch() string {
	var b strings.Builder
	b.WriteString("function classify(code) {\n  switch (code) {\n")
	for i := 0; i < 20; i++ {
		fmt.Fprintf(&b, "    case %d: return handler_%d();\n", i, i)
	}
	b.WriteString("  }\n}\n")
	return b.String()
}

func TestScan_literalRepetitiveBlockStillReportedAsExact(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "a.ts", literalSwitch())
	writeTempFile(t, dir, "b.ts", literalSwitch())

	cfg := defaultConfig
	cfg.Default.MinTokens = 20
	matches, _, _, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	found := false
	for _, m := range matches {
		if m.Kind == "exact" {
			found = true
		}
	}
	if !found {
		t.Fatalf("una copia literal debe seguir reportándose como exact, got %+v", matches)
	}
}
