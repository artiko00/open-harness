package main

// languageMapping defines how a language maps source files to test files
type languageMapping struct {
	extensions   []string
	testSuffixes []string
	testPrefixes []string
	packageBased bool // true when one test file covers all files in the directory (e.g. Go)
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
			testSuffixes:  []string{".test", ".spec"},
			testPrefixes:  []string{"test_"},
		},
		"javascript": {
			extensions:   []string{".js", ".jsx"},
			testSuffixes:  []string{".test", ".spec"},
			testPrefixes:  []string{"test_"},
		},
		"python": {
			extensions:   []string{".py"},
			testSuffixes:  []string{"_test"},
			testPrefixes:  []string{"test_"},
		},
		"ruby": {
			extensions:   []string{".rb"},
			testSuffixes:  []string{"_spec", "_test"},
			testPrefixes:  []string{},
		},
		"rust": {
			extensions:   []string{".rs"},
			testSuffixes:  []string{"_test"},
			testPrefixes:  []string{},
		},
		"java": {
			extensions:   []string{".java"},
			testSuffixes:  []string{"Test"},
			testPrefixes:  []string{},
		},
		"kotlin": {
			extensions:   []string{".kt", ".kts"},
			testSuffixes:  []string{"Test"},
			testPrefixes:  []string{},
		},
		"csharp": {
			extensions:   []string{".cs"},
			testSuffixes:  []string{"Tests"},
			testPrefixes:  []string{},
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
		"ruby", "rust", "java", "kotlin", "csharp",
	}
}