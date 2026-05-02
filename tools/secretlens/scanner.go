package main

import (
	"os"
	"path/filepath"
	"regexp"
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

func scan(root string, cfg Config) ([]Finding, error) {
	compiled, err := compilePatterns(cfg.Patterns)
	if err != nil {
		return nil, err
	}

	var findings []Finding

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
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

		if d.IsDir() || isBinaryPath(path) || isBinaryContent(path) {
			return nil
		}

		ff, err := scanFile(path, relPath, compiled, cfg.Allowlist)
		if err != nil {
			return nil
		}

		findings = append(findings, ff...)
		return nil
	})

	return findings, err
}
