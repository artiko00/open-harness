package tomlmin

import "strings"

// scanStringLiteral mide el literal que empieza en s[0] (una comilla simple o
// doble) y devuelve cuántos bytes ocupa y cuántos saltos de línea contiene.
// Ante un literal sin cerrar devuelve el tramo hasta el fin de línea —o hasta
// el fin del documento si era multilínea—, sin error: quien parsee el valor
// dará el diagnóstico.
func scanStringLiteral(s string) (int, int) {
	q := s[0]
	if isTripleQuote(s, q) {
		return scanMultiline(s, q)
	}
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\n':
			return i, 0
		case '\\':
			if q == '"' {
				i++
			}
		case q:
			return i + 1, 0
		}
	}
	return len(s), 0
}

func isTripleQuote(s string, q byte) bool {
	return len(s) >= 3 && s[1] == q && s[2] == q
}

func scanMultiline(s string, q byte) (int, int) {
	delim := strings.Repeat(string(q), 3)
	for at := 3; at < len(s); {
		idx := strings.Index(s[at:], delim)
		if idx < 0 {
			break
		}
		end := at + idx
		if q == '"' && escapedAt(s, end) {
			at = end + 1
			continue
		}
		return end + 3, strings.Count(s[:end+3], "\n")
	}
	return len(s), strings.Count(s, "\n")
}

// escapedAt indica si la posición i viene precedida por un número impar de
// barras invertidas, o sea si el carácter en i está escapado.
func escapedAt(s string, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}

// indexEqOutsideStrings devuelve la posición del primer '=' que no esté dentro
// de un literal, o -1.
func indexEqOutsideStrings(s string) int {
	for i := 0; i < len(s); {
		switch s[i] {
		case '=':
			return i
		case '"', '\'':
			n, _ := scanStringLiteral(s[i:])
			i += n
		default:
			i++
		}
	}
	return -1
}
