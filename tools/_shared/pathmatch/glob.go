// Package pathmatch centraliza el emparejamiento de globs y la detección de
// binarios compartida por los tools (linelens, dupelens, secretlens, testlens).
package pathmatch

import (
	"path/filepath"
	"strings"
)

// MatchGlob indica si relPath coincide con pattern. Soporta "*", "?" y "**".
// El "**" actúa como comodín de cero o más segmentos completos (cruza "/"),
// mientras que "*" y "?" nunca cruzan "/". Un patrón de un solo segmento
// (sin "/") también coincide contra el nombre base a cualquier profundidad.
func MatchGlob(pattern, relPath string) bool {
	pattern = filepath.ToSlash(pattern)
	relPath = filepath.ToSlash(relPath)

	if matchSegments(strings.Split(pattern, "/"), strings.Split(relPath, "/")) {
		return true
	}
	if !strings.Contains(pattern, "/") {
		matched, _ := filepath.Match(pattern, filepath.Base(relPath))
		return matched
	}
	return false
}

// matchSegments empareja segmento a segmento; "**" consume cero o más.
func matchSegments(pat, path []string) bool {
	if len(pat) == 0 {
		return len(path) == 0
	}
	if pat[0] == "**" {
		for i := 0; i <= len(path); i++ {
			if matchSegments(pat[1:], path[i:]) {
				return true
			}
		}
		return false
	}
	if len(path) == 0 {
		return false
	}
	if matched, _ := filepath.Match(pat[0], path[0]); matched {
		return matchSegments(pat[1:], path[1:])
	}
	return false
}

// IsExcluded indica si relPath debe omitirse según los patrones de exclusión.
// Compara cada patrón contra cada segmento del path y contra el path completo
// vía MatchGlob; nunca usa strings.Contains.
func IsExcluded(relPath string, excludes []string) bool {
	relPath = filepath.ToSlash(relPath)
	parts := strings.Split(relPath, "/")

	for _, excl := range excludes {
		excl = filepath.ToSlash(excl)
		for _, part := range parts {
			if matched, _ := filepath.Match(excl, part); matched {
				return true
			}
		}
		if MatchGlob(excl, relPath) {
			return true
		}
	}
	return false
}
