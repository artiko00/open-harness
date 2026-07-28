package main

import (
	"path/filepath"
	"strings"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// Categorías de clasificación de un archivo tocado.
const (
	catSource   = "source"
	catTest     = "test"
	catExcluded = "excluded"
)

// testPatterns son los layouts de test de los tres ecosistemas (mismo criterio
// que testlens tras F-015, para no divergir).
var testPatterns = []string{
	// JS/TS
	"*.test.*", "*.spec.*", "**/__tests__/**", "**/tests/**",
	// Python
	"test_*.py", "*_test.py",
	// Go
	"*_test.go",
}

// lockfiles son los basenames que se reportan con motivo "lockfile".
var lockfiles = map[string]bool{
	"package-lock.json": true, "pnpm-lock.yaml": true, "yarn.lock": true,
	"poetry.lock": true, "Pipfile.lock": true, "uv.lock": true, "go.sum": true,
}

// classify asigna a cada ruta exactamente una categoría, derivada sólo del
// path (nunca leyendo el archivo). excluded tiene prioridad sobre test.
func classify(path string, exclude []string) string {
	if pathmatch.IsExcluded(path, exclude) {
		return catExcluded
	}
	if isTest(path) {
		return catTest
	}
	return catSource
}

func isTest(path string) bool {
	for _, p := range testPatterns {
		if pathmatch.MatchGlob(p, path) {
			return true
		}
	}
	return false
}

// excludeReason describe por qué se excluyó una ruta: lockfile, generated o el
// motivo genérico.
func excludeReason(path string) string {
	base := filepath.Base(path)
	if lockfiles[base] {
		return "lockfile"
	}
	if strings.HasSuffix(base, ".pb.go") || (strings.HasPrefix(base, "zz_generated") && strings.HasSuffix(base, ".go")) {
		return "generated"
	}
	return "excluded"
}
