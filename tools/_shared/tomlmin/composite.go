package tomlmin

import (
	"fmt"
	"strings"
)

// parseArray reads "[v1, v2, …]" returning a []any. Caller guarantees s[0]=='['.
func parseArray(s string) (any, string, error) {
	rest := strings.TrimLeft(s[1:], " \t")
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
		rest = strings.TrimLeft(after, " \t")
		if strings.HasPrefix(rest, ",") {
			rest = strings.TrimLeft(rest[1:], " \t")
			continue
		}
		if strings.HasPrefix(rest, "]") {
			return out, rest[1:], nil
		}
		return nil, "", fmt.Errorf("expected ',' or ']' in array, got %q", rest)
	}
}

// parseInlineTable reads "{ k1 = v1, k2 = v2 }" returning a map[string]any.
// Caller guarantees s[0]=='{'.
func parseInlineTable(s string) (any, string, error) {
	rest := strings.TrimLeft(s[1:], " \t")
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
		out[k] = v
		rest = strings.TrimLeft(after2, " \t")
		if strings.HasPrefix(rest, ",") {
			rest = strings.TrimLeft(rest[1:], " \t")
			continue
		}
		if strings.HasPrefix(rest, "}") {
			return out, rest[1:], nil
		}
		return nil, "", fmt.Errorf("expected ',' or '}' in inline table, got %q", rest)
	}
}

func readInlineKey(s string) (string, string, error) {
	eq := strings.IndexByte(s, '=')
	if eq < 0 {
		return "", "", fmt.Errorf("missing '=' in inline table near %q", s)
	}
	key := strings.TrimSpace(s[:eq])
	if key == "" {
		return "", "", fmt.Errorf("empty key in inline table")
	}
	return key, strings.TrimLeft(s[eq+1:], " \t"), nil
}
