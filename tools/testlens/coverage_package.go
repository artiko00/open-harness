package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func checkCoveragePackage(cfg config, lang languageMapping) (int, error) {
	exclude := cfg.exclude
	if len(exclude) == 0 {
		exclude = defaultConfig.Exclude
	}
	seen := map[string]bool{}
	err := filepath.Walk(cfg.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if shouldSkipDir(path, exclude) {
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
			fmt.Println(violationLine(relPath+"/", "- no tests found", cfg.noColor))
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
