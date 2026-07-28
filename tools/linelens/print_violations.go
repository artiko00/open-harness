package main

import "fmt"

// printViolations imprime cada archivo en falta mostrando AMBAS métricas:
// líneas de código (contra maxLines) y, si aplica, la profundidad de
// anidamiento (contra maxNesting). Entre paréntesis va siempre el total físico.
func printViolations(v []FileResult, noColor bool) {
	if noColor {
		fmt.Printf("VIOLATIONS (%d files exceed limits):\n\n", len(v))
	} else {
		fmt.Printf("%s%sVIOLATIONS%s (%d files exceed limits):\n\n",
			colorBold, colorRed, colorReset, len(v))
	}

	ancho := anchoPaths(v)
	for _, r := range v {
		linea := detalleViolacion(r, ancho)
		if noColor {
			fmt.Println("  " + linea)
			continue
		}
		fmt.Printf("  %s%s%s\n", colorBold, linea, colorReset)
	}
}

// anchoPaths calcula el ancho de columna para alinear los paths. recortaPath ya
// topa cada path en 60, así que el ancho nunca lo excede.
func anchoPaths(v []FileResult) int {
	ancho := 0
	for _, r := range v {
		n := len(recortaPath(r.RelPath))
		if n > ancho {
			ancho = n
		}
	}
	return ancho
}

// recortaPath acorta con "..." los paths de más de 60 caracteres.
func recortaPath(p string) string {
	if len(p) > 60 {
		return "..." + p[len(p)-57:]
	}
	return p
}

// detalleViolacion arma la línea "<path>  N code / M lines (max: X)" y agrega
// el detalle de anidamiento solo cuando ese es el umbral excedido.
func detalleViolacion(r FileResult, ancho int) string {
	s := fmt.Sprintf("%-*s  %4d code / %d lines (max: %d)",
		ancho, recortaPath(r.RelPath), r.Lines, r.Physical, r.MaxLines)
	if r.MaxNesting > 0 && r.Nesting > r.MaxNesting {
		s += fmt.Sprintf("  nesting %d (max: %d)", r.Nesting, r.MaxNesting)
	}
	return s
}
