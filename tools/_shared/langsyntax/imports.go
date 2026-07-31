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
//
// El fuente se recorre sin materializar sus líneas y el buffer de salida se
// asigna recién ante el primer descarte: un archivo sin imports —o de un
// lenguaje sin familia— se devuelve tal cual, sin copiarlo.
func StripImports(src, ext string) string {
	fam, ok := extFamily[normalizeExt(ext)]
	if !ok {
		return src
	}
	var b strings.Builder
	depth, pos, dropped := 0, 0, false
	for pos <= len(src) {
		nl := strings.IndexByte(src[pos:], '\n')
		end := len(src)
		if nl >= 0 {
			end = pos + nl
		}
		line := src[pos:end]
		if depth > 0 || startsImport(strings.TrimLeft(line, " \t"), fam) {
			if !dropped {
				b.Grow(len(src))
				b.WriteString(src[:pos])
				dropped = true
			}
			depth = clampDepth(depth + delimiterDelta(line))
		} else if dropped {
			b.WriteString(line)
		}
		if nl < 0 {
			break
		}
		if dropped {
			b.WriteByte('\n')
		}
		pos = end + 1
	}
	if !dropped {
		return src
	}
	return b.String()
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
