package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// motivoOmision devuelve el motivo por el que un archivo fuente no se puede
// analizar, o "" si es analizable. Descarta lo no regular antes de abrir nada,
// para que un FIFO nunca bloquee el proceso.
func motivoOmision(path string, d fs.DirEntry) string {
	if !pathmatch.IsRegular(d) {
		return pathmatch.ReasonNotRegular
	}
	f, err := os.Open(path)
	if err != nil {
		return pathmatch.ReasonReadError
	}
	f.Close()
	if pathmatch.IsBinaryPath(path) || pathmatch.IsBinaryContent(path) {
		return pathmatch.ReasonBinary
	}
	return ""
}

// printSkipped imprime la sección SKIPPED del contrato de salida, antes de la
// línea final. No imprime nada si no hubo omisiones.
func printSkipped(skips []pathmatch.Skip) {
	if len(skips) == 0 {
		return
	}
	fmt.Printf("\nSKIPPED (%d file(s) not analyzed):\n\n", len(skips))
	for _, s := range skips {
		fmt.Printf("  %s    %s\n", s.Path, s.Reason)
	}
}

// excludeEfectivo devuelve los patrones de exclusión de la config, o los por
// defecto cuando la config no define ninguno.
func excludeEfectivo(cfg config) []string {
	if len(cfg.exclude) == 0 {
		return defaultConfig.Exclude
	}
	return cfg.exclude
}
