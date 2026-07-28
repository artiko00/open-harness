package main

import (
	"reflect"
	"strings"
	"testing"
)

// TestTutorial_MencionaClavesConfig verifica exit 0 y que la guía nombre cada
// clave JSON del struct Config (para que no se desincronice del código).
func TestTutorial_MencionaClavesConfig(t *testing.T) {
	var code int
	out := capturaSalida(t, func() { code = run([]string{"--tutorial"}) })
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("el tutorial no imprimió nada")
	}
	tp := reflect.TypeOf(Config{})
	for i := 0; i < tp.NumField(); i++ {
		key := tp.Field(i).Tag.Get("json")
		if !strings.Contains(out, key) {
			t.Errorf("el tutorial no menciona la clave de config %q", key)
		}
	}
}

// TestTutorial_NoColorSinANSI garantiza que --no-color no emite códigos ANSI.
func TestTutorial_NoColorSinANSI(t *testing.T) {
	var code int
	out := capturaSalida(t, func() { code = run([]string{"tutorial", "--no-color"}) })
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Contains(out, "\033") {
		t.Errorf("--no-color no debe contener ANSI:\n%s", out)
	}
}

// TestTutorial_ColorPorDefecto verifica que sin --no-color la guía lleva ANSI.
func TestTutorial_ColorPorDefecto(t *testing.T) {
	out := capturaSalida(t, func() { run([]string{"tutorial"}) })
	if !strings.Contains(out, "\033") {
		t.Errorf("sin --no-color la guía debe incluir códigos ANSI:\n%s", out)
	}
}
