package tomlmin

import (
	"fmt"
	"strings"
)

// parseArray reads "[v1, v2, …]" returning a []any. Caller guarantees s[0]=='['.
// Admite coma final antes del cierre y saltos de línea entre elementos.
func parseArray(s string) (any, string, error) {
	rest := trimWS(s[1:])
	out := []any{}
	if strings.HasPrefix(rest, "]") {
		return out, rest[1:], nil
	}
	for {
		if rest == "" || rest[0] == ',' || rest[0] == ']' {
			return nil, "", fmt.Errorf("unexpected token in array near %q", rest)
		}
		v, after, err := parseValue(rest)
		if err != nil {
			return nil, "", err
		}
		out = append(out, v)
		rest = trimWS(after)
		if strings.HasPrefix(rest, ",") {
			rest = trimWS(rest[1:])
			if strings.HasPrefix(rest, "]") { // coma final: TOML la permite
				return out, rest[1:], nil
			}
			continue
		}
		if strings.HasPrefix(rest, "]") {
			return out, rest[1:], nil
		}
		return nil, "", fmt.Errorf("expected ',' or ']' in array, got %q", rest)
	}
}

// parseInlineTable reads "{ k1 = v1, k2 = v2 }" returning a map[string]any.
// Caller guarantees s[0]=='{'. A diferencia de los arrays, TOML no admite coma
// final aquí y se sigue rechazando.
func parseInlineTable(s string) (any, string, error) {
	rest := trimWS(s[1:])
	out := map[string]any{}
	if strings.HasPrefix(rest, "}") {
		return out, rest[1:], nil
	}
	for {
		k, after, err := readInlineKey(rest)
		if err != nil {
			return nil, "", err
		}
		v, after2, err := parseValue(after)
		if err != nil {
			return nil, "", err
		}
		if err := assignKey(out, k, v); err != nil {
			return nil, "", err
		}
		rest = trimWS(after2)
		if strings.HasPrefix(rest, ",") {
			rest = trimWS(rest[1:])
			continue
		}
		if strings.HasPrefix(rest, "}") {
			return out, rest[1:], nil
		}
		return nil, "", fmt.Errorf("expected ',' or '}' in inline table, got %q", rest)
	}
}

func readInlineKey(s string) (string, string, error) {
	eq := indexEqOutsideStrings(s)
	if eq < 0 {
		return "", "", fmt.Errorf("missing '=' in inline table near %q", s)
	}
	key := strings.TrimSpace(s[:eq])
	if key == "" {
		return "", "", fmt.Errorf("empty key in inline table")
	}
	return key, trimWS(s[eq+1:]), nil
}
