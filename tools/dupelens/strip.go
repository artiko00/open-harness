package main

import "strings"

// stripCommentsAndStrings preprocesa el código fuente eliminando contenido
// de strings literales y comentarios, preservando saltos de línea para que
// el seguimiento de líneas de tokenize() siga siendo correcto.
//
// Reconoce:
//   - Strings: "...", '...', `...` (con escape \" y \')
//   - Comentarios single-line: //, #
//   - Comentarios multi-line: /* ... */
//
// Approach: state machine simple sobre bytes. ASCII-safe; los caracteres
// multi-byte (UTF-8 dentro de strings/comments) se descartan junto al resto.
func stripCommentsAndStrings(src string) string {
	var b strings.Builder
	b.Grow(len(src))
	i := 0
	for i < len(src) {
		c := src[i]
		if c == '/' && i+1 < len(src) && src[i+1] == '*' {
			i = skipBlockComment(src, i+2, &b)
			continue
		}
		if c == '/' && i+1 < len(src) && src[i+1] == '/' {
			i = skipLineComment(src, i+2)
			continue
		}
		if c == '#' {
			i = skipLineComment(src, i+1)
			continue
		}
		if c == '"' || c == '\'' || c == '`' {
			i = skipString(src, i+1, c, &b)
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

// skipBlockComment salta hasta */ preservando newlines en el output.
func skipBlockComment(src string, start int, b *strings.Builder) int {
	for i := start; i < len(src); i++ {
		if i+1 < len(src) && src[i] == '*' && src[i+1] == '/' {
			return i + 2
		}
		if src[i] == '\n' {
			b.WriteByte('\n')
		}
	}
	return len(src)
}

// skipLineComment salta hasta el siguiente newline (sin consumirlo).
func skipLineComment(src string, start int) int {
	for i := start; i < len(src); i++ {
		if src[i] == '\n' {
			return i
		}
	}
	return len(src)
}

// skipString salta hasta el delimitador de cierre, manejando escapes para
// "..." y '...' (no para backticks — Go raw strings no escapan).
// Preserva newlines internos para conservar tracking de líneas.
func skipString(src string, start int, quote byte, b *strings.Builder) int {
	canEscape := quote != '`'
	for i := start; i < len(src); i++ {
		c := src[i]
		if canEscape && c == '\\' && i+1 < len(src) {
			i++
			continue
		}
		if c == quote {
			return i + 1
		}
		if c == '\n' {
			b.WriteByte('\n')
		}
	}
	return len(src)
}
