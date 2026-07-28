package main

import (
	"path/filepath"
	"strings"
)

// findTestCandidates returns possible test filenames for a given source file.
// Un archivo que ya es test no necesita candidatos.
func findTestCandidates(filename string, lang languageMapping) []string {
	if !isSourceFile(filename, lang) {
		return nil
	}
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	var candidates []string
	for _, suffix := range lang.testSuffixes {
		candidates = append(candidates, base+suffix+ext)
	}
	for _, prefix := range lang.testPrefixes {
		candidates = append(candidates, prefix+filename)
	}
	return candidates
}

// tieneExt indica si filename termina en alguna de las extensiones dadas.
func tieneExt(filename string, extensions []string) bool {
	ext := filepath.Ext(filename)
	for _, e := range extensions {
		if ext == e {
			return true
		}
	}
	return false
}

// isTestFile checks if a file is a test file based on extension and naming
// patterns (sufijos O prefijos: reconoce tanto foo_test.py como test_foo.py).
func isTestFile(filename string, extensions []string, indicators []string) bool {
	if !tieneExt(filename, extensions) {
		return false
	}
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	for _, ind := range indicators {
		if strings.HasSuffix(base, ind) || strings.HasPrefix(base, ind) {
			return true
		}
	}
	return false
}

// isSourceFile indica si filename es fuente analizable: tiene extensión del
// lenguaje y NO es un archivo de test (por sufijo o por prefijo, p.ej. test_x.py).
func isSourceFile(filename string, lang languageMapping) bool {
	if !tieneExt(filename, lang.extensions) {
		return false
	}
	indicators := append(append([]string{}, lang.testSuffixes...), lang.testPrefixes...)
	return !isTestFile(filename, lang.extensions, indicators)
}

// testExists probes every layout-aware search dir (ADR-019, F-016). Un candidato
// solo cuenta si además contiene un marcador de test del lenguaje (8.15).
func testExists(sourcePath, root string, candidates []string, lang languageMapping) string {
	for _, dir := range testSearchDirs(sourcePath, root, lang) {
		for _, candidate := range candidates {
			p := filepath.Join(dir, candidate)
			if fileExists(p) && contieneMarcador(p, lang.testMarkers) {
				return candidate
			}
		}
	}
	return ""
}
