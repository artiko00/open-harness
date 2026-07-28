package main

import (
	"fmt"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// printSkipped lista los archivos que no se pudieron analizar, con su motivo.
func printSkipped(skips []pathmatch.Skip, noColor bool) {
	if len(skips) == 0 {
		return
	}

	if noColor {
		fmt.Printf("SKIPPED (%d file(s) not analyzed):\n\n", len(skips))
	} else {
		fmt.Printf("%s%sSKIPPED%s (%d file(s) not analyzed):\n\n",
			colorBold, colorYellow, colorReset, len(skips))
	}

	for _, s := range skips {
		fmt.Printf("  %s    %s\n", s.Path, s.Reason)
	}
	fmt.Println()
}

// printOK imprime la línea final limpia, con el conteo de omitidos si los hubo.
func printOK(n int, noColor bool) {
	if noColor {
		fmt.Printf("OK: no secrets detected%s\n", sufijoOK(n))
		return
	}
	fmt.Printf("%s%sOK%s: no secrets detected%s\n",
		colorBold, colorGreen, colorReset, sufijoOK(n))
}

// sufijoOK devuelve " (N skipped)" para la línea OK, o "" si no hay omitidos.
func sufijoOK(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf(" (%d skipped)", n)
}

// sufijoResumen devuelve ", N skipped" para la línea SUMMARY, o "" si no hay.
func sufijoResumen(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf(", %d skipped", n)
}
