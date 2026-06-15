package main

// languageMapping defines how a language maps source files to test files.
// See ADR-019 for the testDirs / mirrors layout strategy.
type languageMapping struct {
	extensions   []string
	testSuffixes []string
	testPrefixes []string
	testDirs     []string    // adjacent subdirs to also probe (e.g. "__tests__", "tests")
	mirrors      [][2]string // dir-prefix substitutions (e.g. ["src/main/java","src/test/java"])
	packageBased bool        // true when one test file covers all files in the directory (e.g. Go)
}

// mapLanguageExtensions returns the mapping of languages to their file patterns
func mapLanguageExtensions() map[string]languageMapping {
	return map[string]languageMapping{
		"go": {
			extensions:   []string{".go"},
			testSuffixes: []string{"_test"},
			testPrefixes: []string{},
			packageBased: true,
		},
		"typescript": {
			extensions:   []string{".ts", ".tsx"},
			testSuffixes: []string{".test", ".spec"},
			testPrefixes: []string{"test_"},
			testDirs:     []string{"__tests__", "tests"},
		},
		"javascript": {
			extensions:   []string{".js", ".jsx"},
			testSuffixes: []string{".test", ".spec"},
			testPrefixes: []string{"test_"},
			testDirs:     []string{"__tests__", "tests"},
		},
		"python": {
			extensions:   []string{".py"},
			testSuffixes: []string{"_test"},
			testPrefixes: []string{"test_"},
			testDirs:     []string{"tests"},
			mirrors:      [][2]string{{"src", "tests"}},
		},
		"ruby": {
			extensions:   []string{".rb"},
			testSuffixes: []string{"_spec", "_test"},
			testPrefixes: []string{},
			testDirs:     []string{"spec"},
			mirrors:      [][2]string{{"lib", "spec"}},
		},
		"rust": {
			extensions:   []string{".rs"},
			testSuffixes: []string{"_test"},
			testPrefixes: []string{},
			testDirs:     []string{"tests"},
		},
		"java": {
			extensions:   []string{".java"},
			testSuffixes: []string{"Test"},
			testPrefixes: []string{},
			mirrors:      [][2]string{{"src/main/java", "src/test/java"}},
		},
		"kotlin": {
			extensions:   []string{".kt", ".kts"},
			testSuffixes: []string{"Test"},
			testPrefixes: []string{},
			mirrors:      [][2]string{{"src/main/kotlin", "src/test/kotlin"}},
		},
		"csharp": {
			extensions:   []string{".cs"},
			testSuffixes:  []string{"Tests"},
			testPrefixes:  []string{},
		},
		"dart": {
			extensions:   []string{".dart"},
			testSuffixes: []string{"_test"},
			testPrefixes: []string{},
			testDirs:     []string{"test"},
			mirrors:      [][2]string{{"lib", "test"}},
		},
	}
}

// extensionsForLanguage returns source extensions for a given language
func extensionsForLanguage(lang string) []string {
	if m, ok := mapLanguageExtensions()[lang]; ok {
		return m.extensions
	}
	return nil
}

// supportedLanguages returns list of all supported language keys
func supportedLanguages() []string {
	return []string{
		"go", "typescript", "javascript", "python",
		"ruby", "rust", "java", "kotlin", "csharp", "dart",
	}
}