package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// coverage es el resultado del análisis: archivos escaneados, rutas sin test,
// omisiones y si el lenguaje analiza por paquete (afecta el sufijo de consola).
type coverage struct {
	scanned  int
	untested []string
	skips    []pathmatch.Skip
	pkg      bool
}

func checkCoverage(cfg config) (coverage, error) {
	exclude := excludeEfectivo(cfg)
	lang := resolverLenguaje(cfg, exclude)

	if lang.packageBased {
		return checkCoveragePackage(cfg, lang)
	}

	notest := notestEfectivo(cfg)
	cov := coverage{}
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
		if testExists(path, cfg.root, findTestCandidates(filepath.Base(path), lang), lang) == "" {
			cov.untested = append(cov.untested, relSlash)
		}
		return nil
	})
	return cov, err
}

// resolverLenguaje devuelve el mapping REAL del lenguaje: el pedido por --lang o,
// en modo "auto", el detectado por conteo de archivos. Si la detección falla
// (ningún archivo soportado), cae a todas las extensiones sin patrones de test.
func resolverLenguaje(cfg config, exclude []string) languageMapping {
	mappings := mapLanguageExtensions()
	if cfg.language != "auto" {
		return mappings[cfg.language]
	}
	if key := detectLanguage(cfg.root, exclude); key != "" {
		return mappings[key]
	}
	fmt.Fprintln(os.Stderr, "Could not detect language, using all extensions")
	return languageMapping{extensions: allExtensions(mappings)}
}
