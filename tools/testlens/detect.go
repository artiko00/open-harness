package main

import (
	"io/fs"
	"path/filepath"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

func allExtensions(mappings map[string]languageMapping) []string {
	var exts []string
	seen := make(map[string]bool)
	for _, lang := range supportedLanguages() {
		for _, e := range mappings[lang].extensions {
			if !seen[e] {
				exts = append(exts, e)
				seen[e] = true
			}
		}
	}
	return exts
}

// detectLanguage recorre supportedLanguages() en ORDEN FIJO, cuenta los archivos
// de cada lenguaje bajo root (podando los excludes) y elige el de MAYOR conteo.
// Sin umbral: cualquier archivo alcanza. Determinista: ante empate gana el
// primero de supportedLanguages(). Devuelve "" si no hay archivos soportados.
func detectLanguage(root string, exclude []string) string {
	extCounts := contarExtensiones(root, exclude)
	mappings := mapLanguageExtensions()
	best := ""
	bestN := 0
	for _, lang := range supportedLanguages() {
		n := 0
		for _, ext := range mappings[lang].extensions {
			n += extCounts[ext]
		}
		if n > bestN {
			bestN, best = n, lang
		}
	}
	return best
}

// contarExtensiones acumula el conteo por extensión bajo root, saltando los
// directorios excluidos (sin eso node_modules/ decidiría el lenguaje).
func contarExtensiones(root string, exclude []string) map[string]int {
	extCounts := make(map[string]int)
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if d.IsDir() {
			if pathmatch.IsExcluded(filepath.ToSlash(rel), exclude) {
				return filepath.SkipDir
			}
			return nil
		}
		if ext := filepath.Ext(path); ext != "" {
			extCounts[ext]++
		}
		return nil
	})
	return extCounts
}
