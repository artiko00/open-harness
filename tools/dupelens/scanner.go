package main

import (
	"os"
	"path/filepath"
)

// scan recorre el árbol de archivos, tokeniza, calcula fingerprints,
// y retorna los matches encontrados.
func scan(root string, cfg Config, minOverride int) ([]Match, error) {
	perFile := make(map[string][]Fingerprint)
	minTokens := cfg.Default.MinTokens
	if minOverride > 0 {
		minTokens = minOverride
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
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

		threshold := minTokens
		if rule != nil && rule.MinTokens > 0 {
			threshold = rule.MinTokens
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		tokens := tokenize(string(data))
		fps := fingerprint(tokens, threshold)
		if len(fps) > 0 {
			perFile[relPath] = fps
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return findDuplicates(perFile, minTokens), nil
}
