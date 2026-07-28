package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

func compilePatterns(rules []PatternRule) ([]compiledRule, error) {
	compiled := make([]compiledRule, 0, len(rules))
	for _, r := range rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", r.Pattern, err)
		}
		compiled = append(compiled, compiledRule{r, re})
	}
	return compiled, nil
}

// scanFile devuelve los hallazgos y, si el archivo no pudo analizarse, el
// motivo de omisión. Motivo vacío significa que el archivo se leyó completo.
// Decodifica UTF-16 (por BOM) a UTF-8 antes de aplicar las reglas y usa una
// ventana de una línea de lookahead para asignaciones partidas.
func scanFile(path, relPath string, compiled []compiledRule, cfg Config) ([]Finding, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, pathmatch.ReasonReadError
	}

	lines, reason := splitLines(decodeContent(data))
	if reason != "" {
		return nil, reason
	}

	var findings []Finding
	for i, line := range lines {
		candidate := line
		if i+1 < len(lines) && isDanglingKey(line) {
			candidate = line + " " + strings.TrimSpace(lines[i+1])
		}
		findings = append(findings, matchLine(candidate, i+1, relPath, compiled, cfg)...)
	}
	return findings, ""
}

func splitLines(data []byte) ([]string, string) {
	s := bufio.NewScanner(bytes.NewReader(data))
	s.Buffer(make([]byte, 64*1024), pathmatch.BufferLimit)

	var lines []string
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	// El único error posible sobre un bytes.Reader es bufio.ErrTooLong (línea
	// que excede el buffer); io.EOF no cuenta como error para el Scanner.
	if s.Err() != nil {
		return nil, pathmatch.ReasonLineTooLong
	}
	return lines, ""
}

func isAllowed(value string, allowlist []string) bool {
	lower := strings.ToLower(value)
	for _, a := range allowlist {
		if strings.Contains(lower, strings.ToLower(a)) {
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
