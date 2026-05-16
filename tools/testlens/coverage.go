package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func checkCoverage(cfg config) (int, error) {
	mappings := mapLanguageExtensions()
	extensions := getExtensions(cfg.language, cfg.root, mappings)
	lang := getLanguageMapping(cfg.language, mappings, extensions)

	if lang.packageBased {
		return checkCoveragePackage(cfg, lang)
	}

	exclude := cfg.exclude
	if len(exclude) == 0 {
		exclude = defaultConfig.Exclude
	}
	violations := 0
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
		if !isSourceFile(path, extensions) {
			return nil
		}
		filename := filepath.Base(path)
		if testExists(path, cfg.root, findTestCandidates(filename, lang), lang) == "" {
			relPath, _ := filepath.Rel(cfg.root, path)
			fmt.Printf("  %s - no test found\n", relPath)
			violations++
		}
		return nil
	})
	return violations, err
}

func getExtensions(language, root string, mappings map[string]languageMapping) []string {
	if language == "auto" {
		extensions := detectLanguageFromFiles(root)
		if extensions == nil {
			fmt.Println("Could not detect language, using all extensions")
			return allExtensions(mappings)
		}
		return extensions
	}
	if m, ok := mappings[language]; ok {
		return m.extensions
	}
	return nil
}

func getLanguageMapping(language string, mappings map[string]languageMapping, extensions []string) languageMapping {
	if language != "auto" {
		return mappings[language]
	}
	return languageMapping{extensions: extensions, testSuffixes: []string{"test", "spec", "Test", "Spec"}}
}
