package main

import (
	"fmt"
	"io"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// writeSkipped imprime la sección SKIPPED con una línea por archivo omitido.
// No imprime nada si no hubo omisiones.
func writeSkipped(w io.Writer, skips []pathmatch.Skip) {
	if len(skips) == 0 {
		return
	}
	fmt.Fprintf(w, "SKIPPED (%d file(s) not analyzed):\n\n", len(skips))
	for _, s := range skips {
		fmt.Fprintf(w, "  %s    %s\n", s.Path, s.Reason)
	}
	fmt.Fprintln(w)
}

// skippedSuffix aplica format (con un %d) al conteo de omitidos; "" si no hubo.
func skippedSuffix(skips []pathmatch.Skip, format string) string {
	if len(skips) == 0 {
		return ""
	}
	return fmt.Sprintf(format, len(skips))
}
