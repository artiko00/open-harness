package main

import (
	"fmt"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// printSkips imprime la sección de archivos no analizados con su motivo.
func printSkips(skips []pathmatch.Skip, noColor bool) {
	if len(skips) == 0 {
		return
	}
	fmt.Println()
	if noColor {
		fmt.Printf("SKIPPED (%d file(s) not analyzed):\n\n", len(skips))
	} else {
		fmt.Printf("%s%sSKIPPED%s (%d file(s) not analyzed):\n\n",
			colorBold, colorYellow, colorReset, len(skips))
	}

	ancho := 0
	for _, s := range skips {
		if len(s.Path) > ancho {
			ancho = len(s.Path)
		}
	}
	for _, s := range skips {
		fmt.Printf("  %-*s    %s\n", ancho, s.Path, s.Reason)
	}
}

// printOK imprime la línea final de éxito, con el conteo de omitidos si hubo.
func printOK(total, skipped int, noColor bool) {
	sufijo := ""
	if skipped > 0 {
		sufijo = fmt.Sprintf(" (%d skipped)", skipped)
	}
	if noColor {
		fmt.Printf("OK: all %d files within limits%s\n", total, sufijo)
		return
	}
	fmt.Printf("%s%sOK%s: all %d files within limits%s\n",
		colorBold, colorGreen, colorReset, total, sufijo)
}

// sufijoSkipped devuelve el fragmento ", N skipped" para la línea SUMMARY.
func sufijoSkipped(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf(", %d skipped", n)
}
