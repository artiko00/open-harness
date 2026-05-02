package main

import (
	"path/filepath"
	"strings"
)

// matchGlob soporta patrones con **, *, ? para paths relativos.
// Ejemplos: "**/*.spec.ts", "*.go", "node_modules", "src/**"
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

	// El sufijo puede contener * y ? pero no más **
	// Intentar contra el nombre del archivo y contra segmentos del path
	base := filepath.Base(relPath)
	if matched, _ := filepath.Match(suffix, base); matched {
		return true
	}

	// También intentar contra sub-paths (ej: "**/*.spec.ts" contra "a/b/c.spec.ts")
	pathParts := strings.Split(relPath, "/")
	for i := range pathParts {
		sub := strings.Join(pathParts[i:], "/")
		if matched, _ := filepath.Match(suffix, sub); matched {
			return true
		}
	}
	return false
}

// isExcluded retorna true si el path relativo debe omitirse.
// Compara contra cada componente del path y contra el path completo.
func isExcluded(relPath string, excludes []string) bool {
	relPath = filepath.ToSlash(relPath)
	parts := strings.Split(relPath, "/")

	for _, excl := range excludes {
		excl = filepath.ToSlash(excl)

		// Coincidencia exacta de segmento (ej: "node_modules")
		for _, part := range parts {
			if matched, _ := filepath.Match(excl, part); matched {
				return true
			}
		}

		// Glob contra el path completo o con **
		if matchGlob(excl, relPath) {
			return true
		}
	}
	return false
}

// ruleForFile retorna la regla que aplica a este path (primera coincidencia).
// Si ninguna regla aplica, retorna nil.
func ruleForFile(relPath string, rules []Rule) *Rule {
	for i := range rules {
		if matchGlob(rules[i].Pattern, relPath) {
			return &rules[i]
		}
	}
	return nil
}
