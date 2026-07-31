package langsyntax

import "strings"

// StripImports elimina las declaraciones de import, include y re-export del
// código fuente, preservando los saltos de línea para no alterar la numeración
// del análisis posterior. Los imports son sintaxis obligatoria de acceso a otro
// módulo: no expresan lógica, y con los identificadores normalizados vuelven
// idénticas las cabeceras de archivos sin relación entre sí.
//
// DEBE aplicarse DESPUÉS de StripComments: con los strings y comentarios ya
// vacíos, ningún delimitador dentro de una ruta o de un comentario altera el
// balance que cierra las declaraciones multilínea.
//
// Reconoce declaraciones de una línea y multilínea (`import { … } from`,
// `import ( … )` de Go, `from x import ( … )` de Python) por balance de
// delimitadores. Una extensión sin familia conocida devuelve el fuente intacto.
func StripImports(src, ext string) string {
	fam, ok := extFamily[normalizeExt(ext)]
	if !ok {
		return src
	}
	var b strings.Builder
	b.Grow(len(src))
	depth := 0
	for i, line := range strings.Split(src, "\n") {
		if i > 0 {
			b.WriteByte('\n')
		}
		trimmed := strings.TrimLeft(line, " \t")
		if depth > 0 {
			depth = clampDepth(depth + delimiterDelta(trimmed))
			continue
		}
		if startsImport(trimmed, fam) {
			depth = clampDepth(delimiterDelta(trimmed))
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}

// startsImport indica si la línea (ya sin indentación) abre una declaración de
// import para la familia dada. Las exclusiones se evalúan primero.
func startsImport(line string, fam family) bool {
	for _, p := range importNegPairs[fam] {
		if startsWithWord(line, p[0]) && strings.Contains(line, p[1]) {
			return false
		}
	}
	for _, p := range importPrefixes[fam] {
		if startsWithWord(line, p) {
			return true
		}
	}
	for _, p := range importPairs[fam] {
		if startsWithWord(line, p[0]) && strings.Contains(line, p[1]) {
			return true
		}
	}
	return false
}

// startsWithWord indica si line empieza con w como palabra completa: lo que
// sigue no puede continuar el identificador (`importedValue` no es `import`).
func startsWithWord(line, w string) bool {
	if !strings.HasPrefix(line, w) {
		return false
	}
	rest := line[len(w):]
	return rest == "" || !isWordByte(rest[0])
}

// isWordByte indica si el byte puede formar parte de un identificador.
func isWordByte(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// delimiterDelta es el balance de delimitadores de apertura menos los de cierre.
func delimiterDelta(line string) int {
	d := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '(', '[', '{':
			d++
		case ')', ']', '}':
			d--
		}
	}
	return d
}

// clampDepth evita que un desbalance deje la profundidad en negativo, lo que
// haría que el stripper nunca volviera a emitir código.
func clampDepth(d int) int {
	if d < 0 {
		return 0
	}
	return d
}
