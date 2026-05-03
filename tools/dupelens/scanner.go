package main

import (
	"os"
	"path/filepath"
)

// scan recorre el árbol de archivos, tokeniza, calcula fingerprints,
// y retorna los matches encontrados + cuenta de archivos escaneados.
// Usa un único windowSize global — fingerprints con ventanas distintas
// no son comparables entre archivos.
func scan(root string, cfg Config, minOverride int) ([]Match, int, error) {
	perFile := make(map[string][]Fingerprint)
	scanned := 0
	windowSize := cfg.Default.MinTokens
	if minOverride > 0 {
		windowSize = minOverride
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return err
			}
			return nil
		}
		relPath, _ := filepath.Rel(root, path)
		relPath = filepath.ToSlash(relPath)

		if isExcluded(relPath, cfg.Exclude) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if isBinaryPath(path) || isBinaryContent(path) {
			return nil
		}
		rule := ruleForFile(relPath, cfg.Rules)
		if rule != nil && rule.Skip {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		scanned++
		tokens := tokenize(string(data))
		fps := fingerprint(tokens, windowSize)
		if len(fps) > 0 {
			perFile[relPath] = fps
		}
		return nil
	})

	if err != nil {
		return nil, scanned, err
	}
	return findDuplicates(perFile, windowSize, cfg.Default.MinLines), scanned, nil
}
