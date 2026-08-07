package tomlmin

import (
	"fmt"
	"strings"
)

// assignKey guarda val en target bajo rawKey, que puede ser una clave simple,
// una dotted key (a.b) o llevar segmentos citados ("a.b", 'a.b'), en cuyo caso
// el punto interior es parte del nombre y no anida.
func assignKey(target map[string]any, rawKey string, val any) error {
	parts, err := splitKey(rawKey)
	if err != nil {
		return err
	}
	for _, seg := range parts[:len(parts)-1] {
		next, ok := target[seg]
		if !ok {
			next = map[string]any{}
			target[seg] = next
		}
		tbl, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("key %q is not a table", seg)
		}
		target = tbl
	}
	target[parts[len(parts)-1]] = val
	return nil
}

// splitKey descompone la clave en segmentos, respetando las comillas.
func splitKey(raw string) ([]string, error) {
	var parts []string
	s := strings.TrimSpace(raw)
	for {
		seg, rest, err := readKeySegment(s)
		if err != nil {
			return nil, err
		}
		parts = append(parts, seg)
		if rest == "" {
			return parts, nil
		}
		if rest[0] != '.' {
			return nil, fmt.Errorf("unexpected %q in key %q", rest, raw)
		}
		s = strings.TrimSpace(rest[1:])
	}
}

func readKeySegment(s string) (string, string, error) {
	if s == "" {
		return "", "", fmt.Errorf("empty key segment")
	}
	if s[0] == '"' || s[0] == '\'' {
		v, rest, err := parseStringValue(s)
		if err != nil {
			return "", "", err
		}
		return v.(string), strings.TrimSpace(rest), nil
	}
	end := strings.IndexByte(s, '.')
	if end < 0 {
		return strings.TrimSpace(s), "", nil
	}
	seg := strings.TrimSpace(s[:end])
	if seg == "" {
		return "", "", fmt.Errorf("empty key segment")
	}
	return seg, s[end:], nil
}
