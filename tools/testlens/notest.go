package main

import "path"

// defaultNoTest son archivos que por convención no llevan test propio: se
// excluyen del reporte para evitar falsos positivos (8.17). Configurable vía la
// clave "notest" del config.
var defaultNoTest = []string{
	"__init__.py", "conftest.py", "settings.py",
	"*_pb2.py", "*.pb.go", "*_gen.go", "*.g.dart",
	"main.go", "doc.go",
}

// noRequiereTest indica si relPath (con separadores "/") es un archivo que no
// debe reportarse como sin test: coincide con algún patrón (glob sobre el nombre
// base) o vive bajo un directorio de migraciones.
func noRequiereTest(relPath string, patterns []string) bool {
	for _, seg := range splitSegments(relPath) {
		if seg == "migrations" {
			return true
		}
	}
	base := path.Base(relPath)
	for _, pat := range patterns {
		if ok, _ := path.Match(pat, base); ok {
			return true
		}
	}
	return false
}

// notestEfectivo devuelve los patrones del config, o los por defecto si no hay.
func notestEfectivo(cfg config) []string {
	if len(cfg.notest) == 0 {
		return defaultNoTest
	}
	return cfg.notest
}
