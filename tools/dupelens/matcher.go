package main

import (
	"path/filepath"
	"strings"
)

// Match representa un par de bloques duplicados detectados.
type Match struct {
	FileA      string
	StartLineA int
	EndLineA   int
	FileB      string
	StartLineB int
	EndLineB   int
	Tokens     int
}

// findDuplicates agrupa Fingerprints por hash y emite Matches cuando
// dos archivos comparten una ventana idéntica.
//
// IMPLEMENTACIÓN PENDIENTE (scaffold):
// La detección efectiva queda como TODO. El scaffold compila y retorna
// lista vacía — la próxima iteración llena el algoritmo de matching.
func findDuplicates(perFile map[string][]Fingerprint, minTokens int) []Match {
	// TODO: agrupar fingerprints por hash y emitir matches verificados
	return nil
}

// matchGlob — semántica gitignore-style copiada del helper de linelens.
// Un patrón sin '/' coincide con el filename en cualquier directorio.
func matchGlob(pattern, relPath string) bool {
	relPath = filepath.ToSlash(relPath)
	pattern = filepath.ToSlash(pattern)

	if !strings.Contains(pattern, "**") {
		matched, _ := filepath.Match(pattern, relPath)
		if matched {
			return true
		}
		matched, _ = filepath.Match(pattern, filepath.Base(relPath))
		return matched
	}

	parts := strings.SplitN(pattern, "**/", 2)
	if len(parts) != 2 {
		matched, _ := filepath.Match(pattern, relPath)
		return matched
	}

	prefix := parts[0]
	suffix := parts[1]

	if prefix != "" && !strings.HasPrefix(relPath, prefix) {
		return false
	}

	base := filepath.Base(relPath)
	if matched, _ := filepath.Match(suffix, base); matched {
		return true
	}

	pathParts := strings.Split(relPath, "/")
	for i := range pathParts {
		sub := strings.Join(pathParts[i:], "/")
		if matched, _ := filepath.Match(suffix, sub); matched {
			return true
		}
	}
	return false
}

// isExcluded retorna true si el path debe omitirse del scan.
func isExcluded(relPath string, excludes []string) bool {
	relPath = filepath.ToSlash(relPath)
	parts := strings.Split(relPath, "/")

	for _, excl := range excludes {
		excl = filepath.ToSlash(excl)
		for _, part := range parts {
			if matched, _ := filepath.Match(excl, part); matched {
				return true
			}
		}
		if matchGlob(excl, relPath) {
			return true
		}
	}
	return false
}

// ruleForFile retorna la regla aplicable al path (primera coincidencia).
func ruleForFile(relPath string, rules []Rule) *Rule {
	for i := range rules {
		if matchGlob(rules[i].Pattern, relPath) {
			return &rules[i]
		}
	}
	return nil
}
