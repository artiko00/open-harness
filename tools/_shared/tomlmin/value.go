package tomlmin

import (
	"fmt"
	"strconv"
	"strings"
)

// parseValue reads a single scalar / array / inline-table from the front of
// src. Caller guarantees src has at least one non-blank char.
func parseValue(src string) (any, string, error) {
	s := strings.TrimLeft(src, " \t")
	switch s[0] {
	case '"':
		return parseString(s)
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
	var b strings.Builder
	i := 1
	for i < len(s) {
		c := s[i]
		if c == '\\' {
			if i+1 >= len(s) {
				return nil, "", fmt.Errorf("unterminated string escape")
			}
			esc, ok := unescape(s[i+1])
			if !ok {
				return nil, "", fmt.Errorf("invalid escape \\%c", s[i+1])
			}
			b.WriteByte(esc)
			i += 2
			continue
		}
		if c == '"' {
			return b.String(), s[i+1:], nil
		}
		b.WriteByte(c)
		i++
	}
	return nil, "", fmt.Errorf("unterminated string")
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

func parseNumber(s string) (any, string, error) {
	end := 0
	for end < len(s) && isNumberChar(s[end]) {
		end++
	}
	if end == 0 {
		return nil, "", fmt.Errorf("unrecognized value: %q", s)
	}
	tok := s[:end]
	n, err := strconv.ParseFloat(tok, 64)
	if err != nil {
		return nil, "", fmt.Errorf("invalid number %q: %w", tok, err)
	}
	return n, s[end:], nil
}

func isNumberChar(c byte) bool {
	return (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E'
}
