package main

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

type Finding struct {
	RelPath  string
	Line     int
	Content  string
	RuleName string
	Severity string
}

type compiledRule struct {
	rule PatternRule
	re   *regexp.Regexp
}

func scan(root string, cfg Config) ([]Finding, []pathmatch.Skip, int, error) {
	compiled, err := compilePatterns(cfg.Patterns)
	if err != nil {
		return nil, nil, 0, err
	}

	var findings []Finding
	var skips []pathmatch.Skip
	scanned := 0

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
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

		// El chequeo de archivo regular va antes de mirar el contenido: abrir un
		// FIFO para detectar binarios colgaría el proceso.
		if !pathmatch.IsRegular(d) {
			skips = append(skips, pathmatch.Skip{Path: relPath, Reason: pathmatch.ReasonNotRegular})
			return nil
		}

		if pathmatch.IsBinaryPath(path) || pathmatch.IsBinaryContent(path) {
			skips = append(skips, pathmatch.Skip{Path: relPath, Reason: pathmatch.ReasonBinary})
			return nil
		}

		ff, reason := scanFile(path, relPath, compiled, cfg)
		if reason != "" {
			skips = append(skips, pathmatch.Skip{Path: relPath, Reason: reason})
			return nil
		}

		scanned++
		findings = append(findings, ff...)
		return nil
	})

	return findings, skips, scanned, err
}
