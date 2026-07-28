package main

import (
	"strings"
	"testing"
)

func TestRun_Tutorial_Exit0(t *testing.T) {
	for _, arg := range []string{"--tutorial", "tutorial"} {
		code := 99
		out := captureStdout(t, func() { code = run([]string{arg}) })
		if code != 0 {
			t.Errorf("run(%q) = %d, want 0", arg, code)
		}
		if len(strings.TrimSpace(out)) == 0 {
			t.Errorf("run(%q) no imprimio nada", arg)
		}
	}
}

func TestRun_Tutorial_MencionaClavesConfig(t *testing.T) {
	out := captureStdout(t, func() { run([]string{"--tutorial"}) })
	// Las claves JSON del struct Config deben aparecer en el tutorial para
	// que no se desincronice con config.go.
	for _, key := range []string{"default", "maxLines", "maxNesting", "rules", "pattern", "skip", "exclude"} {
		if !strings.Contains(out, key) {
			t.Errorf("tutorial no menciona la clave de config %q", key)
		}
	}
}

func TestRun_Tutorial_MencionaFlags(t *testing.T) {
	out := captureStdout(t, func() { run([]string{"--tutorial"}) })
	for _, flag := range []string{"--config", "--fail", "--no-color", "--format"} {
		if !strings.Contains(out, flag) {
			t.Errorf("tutorial no menciona el flag %q", flag)
		}
	}
}

func TestRun_Tutorial_NoColorSinANSI(t *testing.T) {
	out := captureStdout(t, func() { run([]string{"--tutorial", "--no-color"}) })
	if strings.Contains(out, "\033") {
		t.Errorf("tutorial --no-color contiene ANSI")
	}
}

func TestRun_Tutorial_ColorTieneANSI(t *testing.T) {
	out := captureStdout(t, func() { run([]string{"--tutorial"}) })
	if !strings.Contains(out, "\033") {
		t.Errorf("tutorial con color deberia contener ANSI")
	}
}
