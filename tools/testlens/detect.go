package main

import (
	"os"
	"path/filepath"
)

func allExtensions(mappings map[string]languageMapping) []string {
	var exts []string
	seen := make(map[string]bool)
	for _, m := range mappings {
		for _, e := range m.extensions {
			if !seen[e] {
				exts = append(exts, e)
				seen[e] = true
			}
		}
	}
	return exts
}

func detectLanguageFromFiles(root string) []string {
	extCounts := make(map[string]int)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := filepath.Ext(path)
		if ext != "" {
			extCounts[ext]++
		}
		return nil
	})

	mappings := mapLanguageExtensions()
	for _, m := range mappings {
		for _, ext := range m.extensions {
			if extCounts[ext] > 5 {
				return m.extensions
			}
		}
	}
	return nil
}