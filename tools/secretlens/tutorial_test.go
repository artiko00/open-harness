package main

import (
	"strings"
	"testing"
)

// claves del struct Config que el tutorial DEBE mencionar (nombres JSON).
var tutorialConfigKeys = []string{
	"patterns", "allowlist", "exclude", "minEntropy", "disableDefaultPatterns",
}

func TestTutorial_ExitZeroAndPrints(t *testing.T) {
	var code int
	out := capturarSalida(t, func() { code = run([]string{"--tutorial"}) })
	if code != 0 {
		t.Fatalf("exit = %d, quiero 0", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("no imprimio nada")
	}
}

func TestTutorial_MencionaCadaClave(t *testing.T) {
	out := capturarSalida(t, func() { run([]string{"tutorial"}) })
	for _, k := range tutorialConfigKeys {
		if !strings.Contains(out, k) {
			t.Errorf("el tutorial no menciona la clave %q", k)
		}
	}
}

func TestTutorial_MencionaFlags(t *testing.T) {
	out := capturarSalida(t, func() { run([]string{"--tutorial"}) })
	for _, f := range []string{"--config", "--fail", "--no-color", "--format", "--dir"} {
		if !strings.Contains(out, f) {
			t.Errorf("el tutorial no menciona el flag %q", f)
		}
	}
}

func TestTutorial_NoColorSinANSI(t *testing.T) {
	out := capturarSalida(t, func() { run([]string{"--tutorial", "--no-color"}) })
	if strings.Contains(out, "\033") {
		t.Error("con --no-color no debe haber secuencias ANSI")
	}
}

func TestTutorial_ColorContieneANSI(t *testing.T) {
	out := capturarSalida(t, func() { run([]string{"--tutorial"}) })
	if !strings.Contains(out, "\033") {
		t.Error("sin --no-color se espera color ANSI")
	}
}
