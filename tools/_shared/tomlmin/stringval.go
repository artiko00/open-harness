package tomlmin

import (
	"fmt"
	"strings"
)

// parseStringValue despacha entre las cuatro formas de string de TOML: básica,
// literal y sus variantes multilínea.
func parseStringValue(s string) (any, string, error) {
	q := s[0]
	if isTripleQuote(s, q) {
		return parseMultilineString(s, q)
	}
	if q == '\'' {
		return parseLiteralString(s)
	}
	return parseString(s)
}

// parseLiteralString lee '…': sin escapes, tal cual, y sin cruzar la línea.
func parseLiteralString(s string) (any, string, error) {
	end := strings.IndexAny(s[1:], "'\n")
	if end < 0 || s[1+end] == '\n' {
		return nil, "", fmt.Errorf("unterminated literal string")
	}
	return s[1 : 1+end], s[end+2:], nil
}

// parseMultilineString lee """…""" o '''…''', descartando el salto de línea
// que sigue de inmediato al delimitador de apertura.
func parseMultilineString(s string, q byte) (any, string, error) {
	n, _ := scanStringLiteral(s)
	delim := strings.Repeat(string(q), 3)
	if n < 6 || !strings.HasSuffix(s[:n], delim) {
		return nil, "", fmt.Errorf("unterminated multiline string")
	}
	body := strings.TrimPrefix(s[3:n-3], "\n")
	if q == '\'' {
		return body, s[n:], nil
	}
	out, err := unescapeBody(body)
	if err != nil {
		return nil, "", err
	}
	return out, s[n:], nil
}
