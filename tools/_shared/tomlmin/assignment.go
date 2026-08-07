package tomlmin

import (
	"fmt"
	"strings"
)

// applyAssignment parsea "key = value" y guarda el valor en target. Devuelve
// error para líneas malformadas o para valores fuera del subset soportado.
func applyAssignment(target map[string]any, line string) error {
	eq := indexEqOutsideStrings(line)
	if eq < 0 {
		return fmt.Errorf("missing '=' in: %q", line)
	}
	key := strings.TrimSpace(line[:eq])
	rest := strings.TrimSpace(line[eq+1:])
	if key == "" {
		return fmt.Errorf("missing key before '='")
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
	if err := assignKey(target, key, val); err != nil {
		return fmt.Errorf("for key %q: %w", key, err)
	}
	return nil
}
