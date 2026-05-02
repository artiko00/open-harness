package main

import (
	"os"
	"path/filepath"
	"strings"
)

// shouldSkipDir checks if a directory should be skipped based on name patterns
func shouldSkipDir(path string, skipDirs []string) bool {
	for _, skip := range skipDirs {
		if strings.Contains(path, skip) {
			return true
		}
	}
	return false
}

// scanDirectory recursively scans a directory for files with given extensions
func scanDirectory(root string, skipDirs []string, extensions []string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDir(path, skipDirs) {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		for _, e := range extensions {
			if ext == e {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

// filterSourceFiles returns only source files (not test files) from a list
func filterSourceFiles(files []string, extensions []string) []string {
	var sourceFiles []string
	for _, f := range files {
		if isSourceFile(f, extensions) {
			sourceFiles = append(sourceFiles, f)
		}
	}
	return sourceFiles
}