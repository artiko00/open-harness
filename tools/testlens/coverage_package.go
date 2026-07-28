package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

func checkCoveragePackage(cfg config, lang languageMapping) (coverage, error) {
	exclude := excludeEfectivo(cfg)
	notest := notestEfectivo(cfg)
	seen := map[string]bool{}
	cov := coverage{pkg: true}
	err := filepath.WalkDir(cfg.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(cfg.root, path)
		relSlash := filepath.ToSlash(rel)
		if d.IsDir() {
			if pathmatch.IsExcluded(relSlash, exclude) {
				return filepath.SkipDir
			}
			return nil
		}
		if !isSourceFile(filepath.Base(path), lang) || noRequiereTest(relSlash, notest) {
			return nil
		}
		cov.scanned++
		if motivo := motivoOmision(path, d); motivo != "" {
			cov.skips = append(cov.skips, pathmatch.Skip{Path: relSlash, Reason: motivo})
			return nil
		}
		seen[filepath.Dir(path)] = true
		return nil
	})
	if err != nil {
		return cov, err
	}

	dirs := make([]string, 0, len(seen))
	for d := range seen {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	for _, dir := range dirs {
		if !packageHasTests(dir, lang) {
			relPath, _ := filepath.Rel(cfg.root, dir)
			cov.untested = append(cov.untested, filepath.ToSlash(relPath)+"/")
		}
	}
	return cov, nil
}

// packageHasTests indica si dir tiene al menos un archivo de test con un marcador
// real del lenguaje. Un *_test vacío no cuenta (8.15, 8.9).
func packageHasTests(dir string, lang languageMapping) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() || !isTestFile(e.Name(), lang.extensions, lang.testSuffixes) {
			continue
		}
		if contieneMarcador(filepath.Join(dir, e.Name()), lang.testMarkers) {
			return true
		}
	}
	return false
}
