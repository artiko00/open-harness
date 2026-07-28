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
	testMarkers  []string    // in-file markers que confirman que un archivo es un test real
}

// mapLanguageExtensions returns the mapping of languages to their file patterns
func mapLanguageExtensions() map[string]languageMapping {
	return map[string]languageMapping{
		"go": {
			extensions:   []string{".go"},
			testSuffixes: []string{"_test"},
			packageBased: true,
			testMarkers:  []string{"func Test", "func Benchmark", "func Example"},
		},
		"typescript": {
			extensions:   []string{".ts", ".tsx"},
			testSuffixes: []string{".test", ".spec"},
			testPrefixes: []string{"test_"},
			testDirs:     []string{"__tests__", "tests"},
			testMarkers:  []string{"it(", "test(", "describe("},
		},
		"javascript": {
			extensions:   []string{".js", ".jsx"},
			testSuffixes: []string{".test", ".spec"},
			testPrefixes: []string{"test_"},
			testDirs:     []string{"__tests__", "tests"},
			testMarkers:  []string{"it(", "test(", "describe("},
		},
		"python": {
			extensions:   []string{".py"},
			testSuffixes: []string{"_test"},
			testPrefixes: []string{"test_"},
			testDirs:     []string{"tests"},
			mirrors:      [][2]string{{"src", "tests"}, {"app", "tests"}},
			testMarkers:  []string{"def test"},
		},
		"ruby": {
			extensions:   []string{".rb"},
			testSuffixes: []string{"_spec", "_test"},
			testDirs:     []string{"spec"},
			mirrors:      [][2]string{{"lib", "spec"}},
			testMarkers:  []string{"describe ", "it ", "def test"},
		},
		"rust": {
			extensions:   []string{".rs"},
			testSuffixes: []string{"_test"},
			testDirs:     []string{"tests"},
			testMarkers:  []string{"#[test]"},
		},
		"java": {
			extensions:   []string{".java"},
			testSuffixes: []string{"Test"},
			mirrors:      [][2]string{{"src/main/java", "src/test/java"}},
			testMarkers:  []string{"@Test"},
		},
		"kotlin": {
			extensions:   []string{".kt", ".kts"},
			testSuffixes: []string{"Test"},
			mirrors:      [][2]string{{"src/main/kotlin", "src/test/kotlin"}},
			testMarkers:  []string{"@Test"},
		},
		"csharp": {
			extensions:   []string{".cs"},
			testSuffixes: []string{"Tests"},
			testMarkers:  []string{"[Test", "[Fact"},
		},
		"dart": {
			extensions:   []string{".dart"},
			testSuffixes: []string{"_test"},
			testDirs:     []string{"test"},
			mirrors:      [][2]string{{"lib", "test"}},
			testMarkers:  []string{"test(", "group("},
		},
	}
}
