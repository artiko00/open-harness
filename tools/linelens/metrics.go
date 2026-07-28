package main

import (
	"bytes"
	"strings"
)

// physicalLines cuenta las líneas físicas al estilo de un editor: la última
// línea sin salto final también cuenta (por eso difiere de `wc -l`, que cuenta
// saltos de línea). Un archivo vacío tiene 0 líneas.
func physicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	n := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		n++
	}
	return n
}

// nonBlankLines cuenta las líneas con contenido tras recortar espacios; opera
// sobre el stream ya despojado de comentarios/strings, así que comentarios y
// blancos no suman.
func nonBlankLines(s string) int {
	n := 0
	for _, ln := range strings.Split(s, "\n") {
		if strings.TrimSpace(ln) != "" {
			n++
		}
	}
	return n
}

// maxLineLen devuelve la longitud en bytes de la línea más larga.
func maxLineLen(data []byte) int {
	max, cur := 0, 0
	for _, b := range data {
		if b == '\n' {
			if cur > max {
				max = cur
			}
			cur = 0
			continue
		}
		cur++
	}
	if cur > max {
		max = cur
	}
	return max
}

// maxNesting calcula la profundidad máxima de anidamiento por balance de llaves
// sobre el stream ya despojado de comentarios/strings. NO es complejidad
// ciclomática: solo mide cuán hondo llegan los bloques `{ }`.
func maxNesting(s string) int {
	depth, max := 0, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
			if depth > max {
				max = depth
			}
		case '}':
			if depth > 0 {
				depth--
			}
		}
	}
	return max
}
