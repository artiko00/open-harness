package main

import (
	"reflect"
	"strings"
	"testing"
)

// jsonKeys extrae los tags json del struct Config para que el test falle si el
// tutorial se desincroniza del struct.
func jsonKeys() []string {
	t := reflect.TypeOf(Config{})
	keys := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag != "" {
			keys = append(keys, tag)
		}
	}
	return keys
}

func TestTutorialExit0(t *testing.T) {
	var code int
	out := captura(t, func() { code = run([]string{"--tutorial"}) })
	if code != 0 {
		t.Fatalf("--tutorial code=%d, want 0", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("--tutorial no imprimió nada")
	}
}

func TestTutorialAliasSubcomando(t *testing.T) {
	var code int
	captura(t, func() { code = run([]string{"tutorial"}) })
	if code != 0 {
		t.Fatalf("tutorial code=%d, want 0", code)
	}
}

func TestTutorialMencionaClaves(t *testing.T) {
	out := captura(t, func() { run([]string{"--tutorial"}) })
	for _, k := range jsonKeys() {
		if !strings.Contains(out, k) {
			t.Errorf("tutorial no menciona la clave %q", k)
		}
	}
}

func TestTutorialMencionaFlags(t *testing.T) {
	out := captura(t, func() { run([]string{"--tutorial"}) })
	for _, f := range []string{"--fail", "--no-color", "--max-files", "--base", "--dir"} {
		if !strings.Contains(out, f) {
			t.Errorf("tutorial no menciona el flag %q", f)
		}
	}
}

func TestTutorialConColor(t *testing.T) {
	out := captura(t, func() { run([]string{"--tutorial"}) })
	if !strings.Contains(out, "\033[") {
		t.Fatal("tutorial (con color) debería emitir ANSI")
	}
}

func TestTutorialNoColor(t *testing.T) {
	out := captura(t, func() { run([]string{"--tutorial", "--no-color"}) })
	if strings.Contains(out, "\033[") {
		t.Fatal("--tutorial --no-color no debe emitir ANSI")
	}
}
