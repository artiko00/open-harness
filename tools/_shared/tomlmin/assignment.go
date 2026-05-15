package tomlmin

import (
	"fmt"
	"strings"
)

// applyAssignment parses "key = value" and stores the value into target.
// Returns an error for malformed lines or unsupported features
// (dotted keys, missing value, …).
func applyAssignment(target map[string]any, line string) error {
	eq := strings.IndexByte(line, '=')
	if eq < 0 {
		return fmt.Errorf("missing '=' in: %q", line)
	}
	key := strings.TrimSpace(line[:eq])
	rest := strings.TrimSpace(line[eq+1:])
	if key == "" {
		return fmt.Errorf("missing key before '='")
	}
	if strings.Contains(key, ".") {
		return fmt.Errorf("dotted keys inside tables are not supported: %q", key)
	}
	if rest == "" {
		return fmt.Errorf("missing value after '=' for key %q", key)
	}
	val, leftover, err := parseValue(rest)
	if err != nil {
		return fmt.Errorf("for key %q: %w", key, err)
	}
	if strings.TrimSpace(leftover) != "" {
		return fmt.Errorf("trailing garbage after value for %q: %q", key, leftover)
	}
	target[key] = val
	return nil
}

// stripTrailingComment removes "# …" from the end of a line, respecting
// quoted strings (so a "#" inside a string literal is preserved).
func stripTrailingComment(line string) string {
	inStr := false
	escaped := false
	for i, r := range line {
		if escaped {
			escaped = false
			continue
		}
		switch r {
		case '\\':
			if inStr {
				escaped = true
			}
		case '"':
			inStr = !inStr
		case '#':
			if !inStr {
				return line[:i]
			}
		}
	}
	return line
}
