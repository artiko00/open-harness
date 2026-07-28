package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// captureRun ejecuta run redirigiendo stdout y devuelve lo capturado y el exit.
func captureRun(t *testing.T, args []string) (stdout string, code int) {
	t.Helper()
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)
	go func() { var b bytes.Buffer; b.ReadFrom(r); outC <- b.String() }()
	code = run(args)
	w.Close()
	stdout = <-outC
	os.Stdout = orig
	return
}

func TestTutorial_imprime_y_exit0(t *testing.T) {
	stdout, code := captureRun(t, []string{"--tutorial"})
	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("tutorial no imprimió nada")
	}
}

// TestTutorial_menciona_claves garantiza que el tutorial no se desincronice del
// struct Config: cada clave JSON debe aparecer en el texto.
func TestTutorial_menciona_claves(t *testing.T) {
	stdout, _ := captureRun(t, []string{"tutorial"})
	for _, key := range []string{
		"default", "minTokens", "minLines", "windowSize",
		"rules", "pattern", "skip", "exclude",
	} {
		if !strings.Contains(stdout, key) {
			t.Errorf("el tutorial no menciona la clave %q", key)
		}
	}
}

func TestTutorial_menciona_flags(t *testing.T) {
	stdout, _ := captureRun(t, []string{"--tutorial"})
	for _, flag := range []string{"--config", "--fail", "--no-color", "--format"} {
		if !strings.Contains(stdout, flag) {
			t.Errorf("el tutorial no menciona el flag %q", flag)
		}
	}
}

func TestTutorial_noColor_sin_ansi(t *testing.T) {
	stdout, code := captureRun(t, []string{"--tutorial", "--no-color"})
	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if strings.Contains(stdout, "\033") {
		t.Errorf("--no-color no debe contener ANSI, got %q", stdout)
	}
}

func TestTutorial_color_por_defecto(t *testing.T) {
	stdout, _ := captureRun(t, []string{"--tutorial"})
	if !strings.Contains(stdout, "\033") {
		t.Error("sin --no-color se espera al menos un código ANSI")
	}
}
