package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func checkCoveragePackage(cfg config, lang languageMapping) (int, error) {
	seen := map[string]bool{}
	err := filepath.Walk(cfg.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if shouldSkipDir(path, []string{"node_modules", ".git", "vendor", "dist", "build", "testdata"}) {
				return filepath.SkipDir
			}
			return nil
		}
		if isSourceFile(path, lang.extensions) {
			seen[filepath.Dir(path)] = true
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	dirs := make([]string, 0, len(seen))
	for d := range seen {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	violations := 0
	for _, dir := range dirs {
		if !packageHasTests(dir, lang.testSuffixes, lang.extensions) {
			relPath, _ := filepath.Rel(cfg.root, dir)
			fmt.Printf("  %s/ - no tests found\n", relPath)
			violations++
		}
	}
	return violations, nil
}

func packageHasTests(dir string, testSuffixes []string, extensions []string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && isTestFile(e.Name(), extensions, testSuffixes) {
			return true
		}
	}
	return false
}
