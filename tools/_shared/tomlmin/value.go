package tomlmin

import (
	"fmt"
	"strings"
)

// wsChars son los blancos que separan tokens. El salto de línea cuenta porque
// una línea lógica puede abarcar varias físicas (arrays multilínea).
const wsChars = " \t\r\n"

func trimWS(s string) string { return strings.TrimLeft(s, wsChars) }

// parseValue reads a single scalar / array / inline-table from the front of
// src. Caller guarantees src has at least one non-blank char.
func parseValue(src string) (any, string, error) {
	s := trimWS(src)
	switch s[0] {
	case '"', '\'':
		return parseStringValue(s)
	case '[':
		return parseArray(s)
	case '{':
		return parseInlineTable(s)
	case 't', 'f':
		return parseBool(s)
	}
	return parseNumber(s)
}

func parseString(s string) (any, string, error) {
	if s[0] != '"' {
		return nil, "", fmt.Errorf("expected '\"' at start of string")
	}
	i := 1
	for i < len(s) {
		switch s[i] {
		case '\\':
			if i+1 >= len(s) {
				return nil, "", fmt.Errorf("unterminated string escape")
			}
			i += 2
		case '"':
			out, err := unescapeBody(s[1:i])
			return out, s[i+1:], err
		default:
			i++
		}
	}
	return nil, "", fmt.Errorf("unterminated string")
}

func unescapeBody(s string) (string, error) {
	if !strings.ContainsRune(s, '\\') {
		return s, nil
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			i++
			continue
		}
		if i+1 >= len(s) {
			return "", fmt.Errorf("unterminated string escape")
		}
		esc, ok := unescape(s[i+1])
		if !ok {
			return "", fmt.Errorf("invalid escape \\%c", s[i+1])
		}
		b.WriteByte(esc)
		i += 2
	}
	return b.String(), nil
}

func unescape(c byte) (byte, bool) {
	switch c {
	case '"':
		return '"', true
	case '\\':
		return '\\', true
	case 'n':
		return '\n', true
	case 't':
		return '\t', true
	case 'r':
		return '\r', true
	}
	return 0, false
}

func parseBool(s string) (any, string, error) {
	if strings.HasPrefix(s, "true") {
		return true, s[4:], nil
	}
	if strings.HasPrefix(s, "false") {
		return false, s[5:], nil
	}
	return nil, "", fmt.Errorf("invalid literal: %q", s)
}
