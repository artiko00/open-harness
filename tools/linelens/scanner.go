package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// scan recorre el árbol y devuelve los archivos medidos junto con los que no
// se pudieron analizar (omisiones), cada uno con su motivo canónico. Solo mide
// archivos con extensión de código (pathmatch.CodeExtensions); los datos
// (i18n.json, schema.sql, ...) se ignoran en silencio.
func scan(root string, cfg Config, maxOverride int) ([]FileResult, []pathmatch.Skip, error) {
	var results []FileResult
	var skips []pathmatch.Skip

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

		// Antes de cualquier apertura: un FIFO o socket colgaría el proceso.
		if !pathmatch.IsRegular(d) {
			skips = append(skips, pathmatch.Skip{Path: relPath, Reason: pathmatch.ReasonNotRegular})
			return nil
		}

		if pathmatch.IsBinaryPath(path) || pathmatch.IsBinaryContent(path) {
			skips = append(skips, pathmatch.Skip{Path: relPath, Reason: pathmatch.ReasonBinary})
			return nil
		}

		ext := strings.ToLower(filepath.Ext(relPath))
		if !pathmatch.CodeExtensions()[ext] {
			return nil
		}

		rule := ruleForFile(relPath, cfg.Rules)
		if rule != nil && rule.Skip {
			return nil
		}

		maxLines := cfg.Default.MaxLines
		if maxOverride > 0 {
			maxLines = maxOverride
		} else if rule != nil && rule.MaxLines > 0 {
			maxLines = rule.MaxLines
		}

		code, physical, nesting, err := countFile(path, ext)
		if err != nil {
			skips = append(skips, pathmatch.Skip{Path: relPath, Reason: motivoLectura(err)})
			return nil
		}

		results = append(results, FileResult{
			RelPath:    relPath,
			Lines:      code,
			Physical:   physical,
			Nesting:    nesting,
			MaxLines:   maxLines,
			MaxNesting: cfg.Default.MaxNesting,
		})
		return nil
	})

	return results, skips, err
}
