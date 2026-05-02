package main

import (
	"bufio"
	"os"
	"path/filepath"
)

type FileResult struct {
	RelPath  string
	Lines    int
	MaxLines int
	Skipped  bool
}

func (r FileResult) IsViolation() bool {
	return !r.Skipped && r.Lines > r.MaxLines
}

func scan(root string, cfg Config, maxOverride int) ([]FileResult, error) {
	var results []FileResult

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

		maxLines := cfg.Default.MaxLines
		if maxOverride > 0 {
			maxLines = maxOverride
		} else if rule != nil && rule.MaxLines > 0 {
			maxLines = rule.MaxLines
		}

		lines, err := countLines(path)
		if err != nil {
			return nil
		}

		results = append(results, FileResult{
			RelPath:  relPath,
			Lines:    lines,
			MaxLines: maxLines,
		})
		return nil
	})

	return results, err
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}
