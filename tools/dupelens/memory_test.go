package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// genSource genera ~sizeBytes de código sintético con identificadores
// globalmente únicos (un contador por token) repartido en varios archivos. Al
// ser todos distintos casi no hay duplicados: la detección queda lineal y el
// test mide memoria, no el peor caso cuadrático de agrupación por hash.
func genSource(t *testing.T, dir string, sizeBytes int) {
	t.Helper()
	const perFile = 1 << 20 // 1 MB por archivo
	files := sizeBytes/perFile + 1
	g := 0
	for f := 0; f < files; f++ {
		var b strings.Builder
		b.WriteString("package gen\n")
		i := 0
		for b.Len() < perFile {
			// Longitud de línea variable para no alinear ventanas.
			cols := 4 + i%9
			for c := 0; c < cols; c++ {
				fmt.Fprintf(&b, "token_%d ", g)
				g++
			}
			b.WriteByte('\n')
			i++
		}
		p := filepath.Join(dir, fmt.Sprintf("gen_%d.go", f))
		if err := os.WriteFile(p, []byte(b.String()), 0644); err != nil {
			t.Fatalf("escribiendo fuente sintética: %v", err)
		}
	}
}

// TestMemory_scanStaysProportional verifica que el escaneo de ~24 MB de fuente
// mantiene el heap MUY por debajo del viejo factor N*windowSize (que reventaba
// a cientos de MB por pocos MB de fuente). Con fingerprints por índice (sin
// copiar ventanas) la memoria es proporcional al fuente.
//
// Método: se mide runtime.MemStats.HeapAlloc inmediatamente después de scan
// (la basura del escaneo todavía cuenta, dando una cota superior del pico) y se
// asienta un techo de 512 MB. Es -short-aware para no lentificar el ciclo.
func TestMemory_scanStaysProportional(t *testing.T) {
	if testing.Short() {
		t.Skip("omitido en -short: genera ~24 MB de fuente")
	}
	dir := t.TempDir()
	const target = 24 << 20 // 24 MB
	genSource(t, dir, target)

	cfg := Config{Default: DefaultConfig{MinTokens: 50, MinLines: 5}, Exclude: []string{".git"}}

	var before runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	matches, scanned, _, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}

	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	heap := after.HeapAlloc

	const ceiling = 512 << 20
	if heap > ceiling {
		t.Errorf("HeapAlloc=%d MB excede el techo de %d MB (scanned=%d)",
			heap>>20, ceiling>>20, scanned)
	}
	// Sanidad: se escaneó realmente el volumen esperado.
	if scanned < 20 {
		t.Errorf("scanned=%d; se esperaban ≥20 archivos de ~1 MB", scanned)
	}
	t.Logf("scanned=%d, matches=%d, HeapAlloc=%d MB (techo %d MB)",
		scanned, len(matches), heap>>20, ceiling>>20)
}
