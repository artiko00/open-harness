package langsyntax

import "strings"

// StripComments elimina comentarios y el contenido de strings literales del
// código fuente, preservando los saltos de línea para no alterar la numeración
// de líneas del análisis posterior. La extensión `ext` decide la sintaxis de
// comentarios de línea: '#' es comentario en Python/Ruby/shell/YAML pero NO en
// Rust/CSS/JS/Go/C (ver hashCommentExts).
//
// Reconoce:
//   - Strings: "...", '...', `...` (con escape \" y \' salvo backticks)
//   - Comentarios de línea: //, y # según la extensión
//   - Comentarios de bloque: /* ... */
//
// State machine sobre bytes (ASCII-safe): el contenido multi-byte dentro de
// strings/comentarios se descarta junto al resto.
func StripComments(src, ext string) string {
	hash := hashStartsComment(ext)
	var b strings.Builder
	b.Grow(len(src))
	i := 0
	for i < len(src) {
		c := src[i]
		if c == '/' && i+1 < len(src) && src[i+1] == '*' {
			i = skipBlock(src, i+2, &b)
			continue
		}
		if c == '/' && i+1 < len(src) && src[i+1] == '/' {
			i = skipLine(src, i+2)
			continue
		}
		if c == '#' && hash {
			i = skipLine(src, i+1)
			continue
		}
		if c == '"' || c == '\'' || c == '`' {
			i = skipStr(src, i+1, c, &b)
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

// skipBlock salta hasta */ preservando los newlines internos en el output.
func skipBlock(src string, start int, b *strings.Builder) int {
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

// skipLine salta hasta el siguiente newline sin consumirlo.
func skipLine(src string, start int) int {
	for i := start; i < len(src); i++ {
		if src[i] == '\n' {
			return i
		}
	}
	return len(src)
}

// skipStr salta hasta el delimitador de cierre. Maneja escapes para "..." y
// '...' pero no para backticks (raw strings). Preserva newlines internos.
func skipStr(src string, start int, quote byte, b *strings.Builder) int {
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
