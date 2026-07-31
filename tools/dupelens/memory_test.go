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
// mantiene la memoria MUY por debajo del viejo factor N*windowSize (que reventaba
// a cientos de MB por pocos MB de fuente). Con fingerprints por índice (sin
// copiar ventanas) la memoria es proporcional al fuente.
//
// Método: se mide el delta de runtime.MemStats.TotalAlloc a través de scan — los
// bytes que el escaneo pidió al heap en total. Es la métrica correcta para lo que
// el test protege: copiar la ventana en cada fingerprint multiplicaría el total
// por windowSize, se libere o no después.
//
// TotalAlloc es acumulativo y monótono, así que la medición NO depende de cuándo
// corrió el recolector. HeapAlloc sí dependía: medido sin GC previo daba entre
// 458 y 552 MB para el mismo código según se corriera el test aislado, con la
// suite completa o con -v, y cruzaba el techo de 512 MB de forma intermitente.
// Medirlo *con* GC previo tampoco sirve: da 0 MB, porque todo lo que asigna scan
// es local y ya se liberó cuando el test observa.
//
// Es -short-aware para no lentificar el ciclo.
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
	allocated := after.TotalAlloc - before.TotalAlloc

	// ~1.4 GB medidos de forma estable para 24 MB de fuente (≈57x). El techo deja
	// margen holgado y sigue detectando de sobra el regreso del factor N*windowSize,
	// que multiplicaría el total por el tamaño de ventana (25 por defecto).
	const ceiling = 2048 << 20
	if allocated > ceiling {
		t.Errorf("el scan asignó %d MB, excede el techo de %d MB (scanned=%d)",
			allocated>>20, ceiling>>20, scanned)
	}
	// Sanidad: se escaneó realmente el volumen esperado.
	if scanned < 20 {
		t.Errorf("scanned=%d; se esperaban ≥20 archivos de ~1 MB", scanned)
	}
	t.Logf("scanned=%d, matches=%d, asignados=%d MB (techo %d MB)",
		scanned, len(matches), allocated>>20, ceiling>>20)
}
