package main

import (
	"strings"
	"testing"
)

// Un run de tokens idénticos (p. ej. keywords.go tras stripear strings deja
// una tira de "true") no aporta señal estructural: no debe generar
// fingerprints raw ni, por ende, auto-matches ni colisiones contra otros datos.
func TestAddFile_dropsMonotoneRawRuns(t *testing.T) {
	var files []fileData
	var raw, norm []Fingerprint
	content := strings.Repeat("aa ", 60)
	addFile(&files, &raw, &norm, "x.go", content, scanOpts{windowSize: 25})
	if len(raw) != 0 {
		t.Errorf("run monótono no debe producir fingerprints raw; got %d", len(raw))
	}
	if len(files) != 0 {
		t.Errorf("archivo sin señal no debe registrarse; got %d", len(files))
	}
}

func TestNormalizeValue(t *testing.T) {
	cases := map[string]string{
		"func":  "func", // keyword se preserva
		"42":    "NUM",  // literal numérico
		"3.14":  "NUM",  // numérico con punto
		"foo":   "ID",   // identificador
		"foo_1": "ID",   // identificador con dígito y guion bajo
		"a+b":   "a+b",  // operador glued: ni keyword ni número ni id → literal
	}
	for in, want := range cases {
		if got := normalizeValue(in); got != want {
			t.Errorf("normalizeValue(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestIsIdentifier(t *testing.T) {
	if !isIdentifier("abc1") {
		t.Error("abc1 debe ser identificador")
	}
	if isIdentifier("1abc") {
		t.Error("1abc no debe ser identificador (empieza con dígito)")
	}
	if isIdentifier("a-b") {
		t.Error("a-b no debe ser identificador (guion)")
	}
	if isIdentifier("") {
		t.Error("vacío no debe ser identificador")
	}
}

func TestNormalizeTokens_preservesLines(t *testing.T) {
	in := tok([]string{"foo", "42"}, []int{3, 7})
	out := normalizeTokens(in)
	if out[0].Value != "ID" || out[0].Line != 3 {
		t.Errorf("token 0 = %+v; want {ID 3}", out[0])
	}
	if out[1].Value != "NUM" || out[1].Line != 7 {
		t.Errorf("token 1 = %+v; want {NUM 7}", out[1])
	}
}

func TestDeriveWindow(t *testing.T) {
	if got := deriveWindow(0); got != defaultWindow {
		t.Errorf("deriveWindow(0) = %d; want %d", got, defaultWindow)
	}
	if got := deriveWindow(-3); got != defaultWindow {
		t.Errorf("deriveWindow(-3) = %d; want %d", got, defaultWindow)
	}
	if got := deriveWindow(10); got != 10 {
		t.Errorf("deriveWindow(10) = %d; want 10", got)
	}
}
