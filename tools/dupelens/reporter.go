package main

import (
	"io"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// ReportOpts agrupa todas las opciones de salida del reporter.
// Format: "console" (default) o "json". NoColor desactiva ANSI.
// Snippet: callback opcional que devuelve un fragmento de N líneas
// formateado para mostrar contexto. Si nil, no se imprimen snippets.
// Skips: archivos no analizados, con su motivo.
type ReportOpts struct {
	Format       string
	NoColor      bool
	ScannedCount int
	Snippet      func(file string, startLine, count int) string
	Skips        []pathmatch.Skip
}

// report despacha al backend correspondiente según opts.Format.
// Retorna el número de violations (matches) para que el caller
// decida exit code en CI / git hooks.
func report(matches []Match, opts ReportOpts, w io.Writer) int {
	if opts.Format == "json" {
		return reportJSON(matches, opts, w)
	}
	return reportConsole(matches, opts, w)
}

// Códigos ANSI para output coloreado en consola.
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
)

func boldOpt(no bool) string {
	if no {
		return ""
	}
	return colorBold
}
func redOpt(no bool) string {
	if no {
		return ""
	}
	return colorRed
}
func greenOpt(no bool) string {
	if no {
		return ""
	}
	return colorGreen
}
func grayOpt(no bool) string {
	if no {
		return ""
	}
	return colorGray
}
func resetOpt(no bool) string {
	if no {
		return ""
	}
	return colorReset
}
