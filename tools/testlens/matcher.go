package main

import (
	"path/filepath"
	"strings"
)

// findTestCandidates returns possible test filenames for a given source file
func findTestCandidates(filename string, lang languageMapping) []string {
	// Get base name without extension
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	// If the file is already a test file, return empty (tests don't need tests)
	if isTestFile(filename, lang.extensions, lang.testSuffixes) {
		return nil
	}

	var candidates []string

	// Generate candidates based on suffix patterns
	for _, suffix := range lang.testSuffixes {
		candidates = append(candidates, base+suffix+ext)
	}

	// Generate candidates based on prefix patterns
	for _, prefix := range lang.testPrefixes {
		candidates = append(candidates, prefix+filename)
	}

	return candidates
}

// isTestFile checks if a file is a test file based on extension and naming patterns
func isTestFile(filename string, extensions []string, testIndicators []string) bool {
	ext := filepath.Ext(filename)

	// Must have a source extension
	hasSourceExt := false
	for _, e := range extensions {
		if ext == e {
			hasSourceExt = true
			break
		}
	}
	if !hasSourceExt {
		return false
	}

	// Check test indicators (suffixes/prefixes in the filename without extension)
	base := strings.TrimSuffix(filename, ext)
	for _, indicator := range testIndicators {
		if strings.HasSuffix(base, indicator) || strings.HasPrefix(base, indicator) {
			return true
		}
	}

	return false
}

// isSourceFile checks if a file is a source file (not a test file)
func isSourceFile(filename string, extensions []string) bool {
	ext := filepath.Ext(filename)

	// Must have a source extension
	hasSourceExt := false
	for _, e := range extensions {
		if ext == e {
			hasSourceExt = true
			break
		}
	}
	if !hasSourceExt {
		return false
	}

	// Check if it's actually a test file
	base := strings.TrimSuffix(filename, ext)
	testSuffixes := []string{"test", "spec", "Test", "Spec"}
	for _, suffix := range testSuffixes {
		if strings.HasSuffix(base, suffix) {
			return false
		}
	}

	return true
}

// testExists probes every layout-aware search dir (ADR-019, F-016).
func testExists(sourcePath, root string, candidates []string, lang languageMapping) string {
	for _, dir := range testSearchDirs(sourcePath, root, lang) {
		for _, candidate := range candidates {
			if fileExists(filepath.Join(dir, candidate)) {
				return candidate
			}
		}
	}
	return ""
}

