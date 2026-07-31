package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// defaultWindow es el tamaño de ventana de detección cuando la config no lo fija.
// Independiente del umbral de reporte (minTokens): granularidad estable y barata.
const defaultWindow = 25

// deriveWindow resuelve el tamaño de ventana: el configurado, o defaultWindow.
func deriveWindow(configured int) int {
	if configured <= 0 {
		return defaultWindow
	}
	return configured
}

// scanOpts agrupa lo que la config decide sobre cómo tokenizar y fingerprintear
// cada archivo, para no propagar parámetros sueltos hasta addFile.
type scanOpts struct {
	windowSize   int
	stripImports bool
}

// scan recorre el árbol de archivos, tokeniza los de código (según
// pathmatch.CodeExtensions), calcula fingerprints crudos y normalizados, y
// retorna los matches, la cuenta de escaneados y los omitidos. windowSize
// (detección) y minTokens (reporte) son independientes.
func scan(root string, cfg Config, minOverride int) ([]Match, int, []pathmatch.Skip, error) {
	var files []fileData
	var rawFps, normFps []Fingerprint
	var skips []pathmatch.Skip
	scanned := 0
	minTokens := cfg.Default.MinTokens
	if minOverride > 0 {
		minTokens = minOverride
	}
	opts := scanOpts{windowSize: deriveWindow(cfg.Default.WindowSize), stripImports: cfg.ignoreImports()}
	codeExts := pathmatch.CodeExtensions()

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return err
			}
			return nil
		}
		relPath, _ := filepath.Rel(root, path)
		relPath = filepath.ToSlash(relPath)

		if pathmatch.IsExcluded(relPath, cfg.Exclude) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !pathmatch.IsRegular(d) {
			skips = append(skips, pathmatch.Skip{Path: relPath, Reason: pathmatch.ReasonNotRegular})
			return nil
		}
		if pathmatch.IsBinaryPath(path) || pathmatch.IsBinaryContent(path) {
			skips = append(skips, pathmatch.Skip{Path: relPath, Reason: pathmatch.ReasonBinary})
			return nil
		}
		if rule := ruleForFile(relPath, cfg.Rules); rule != nil && rule.Skip {
			return nil
		}
		// Datos, fixtures y lockfiles no son código: no compiten en la detección.
		if !codeExts[strings.ToLower(filepath.Ext(relPath))] {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			skips = append(skips, pathmatch.Skip{Path: relPath, Reason: pathmatch.ReasonReadError})
			return nil
		}
		scanned++
		addFile(&files, &rawFps, &normFps, relPath, string(data), opts)
		return nil
	})

	if err != nil {
		return nil, scanned, skips, err
	}
	return findDuplicates(files, rawFps, normFps, opts.windowSize, cfg.Default.MinLines, minTokens), scanned, skips, nil
}
