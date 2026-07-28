package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Regresión fase 8: un candidato de test que resulta ser un FIFO no debe colgar
// la verificación de contenido (contieneMarcador hace os.Open, que bloquea sobre
// un named pipe sin escritor). fileExists debe exigir archivo regular.
func TestTestExists_CandidatoFifoNoCuelga(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.ts")
	if err := os.WriteFile(src, []byte("export const f = 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	creaFifo(t, filepath.Join(dir, "a.test.ts"))

	lang := mapLanguageExtensions()["typescript"]
	done := make(chan string, 1)
	go func() {
		done <- testExists(src, dir, findTestCandidates("a.ts", lang), lang)
	}()
	select {
	case got := <-done:
		if got != "" {
			t.Fatalf("un FIFO no es un test válido, got %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("testExists colgó sobre un candidato FIFO")
	}
}
