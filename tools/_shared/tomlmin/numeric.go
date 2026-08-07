package tomlmin

import (
	"fmt"
	"strconv"
	"strings"
)

// parseNumber resuelve el token escalar que no es string, bool, array ni
// inline table: entero (decimal, hex, octal, binario, con '_'), float, o
// fecha/hora RFC 3339, que se devuelve como string sin interpretar.
func parseNumber(s string) (any, string, error) {
	tok, rest := splitToken(s)
	if tok == "" {
		return nil, "", fmt.Errorf("unrecognized value: %q", s)
	}
	if date, after, ok := readDateTime(tok, rest); ok {
		return date, after, nil
	}
	if n, ok := parseIntLiteral(tok); ok {
		return n, rest, nil
	}
	n, err := strconv.ParseFloat(strings.ReplaceAll(tok, "_", ""), 64)
	if err != nil {
		return nil, "", fmt.Errorf("invalid number %q: %w", tok, err)
	}
	return n, rest, nil
}

// splitToken corta el token escalar en el primer blanco o delimitador.
func splitToken(s string) (string, string) {
	end := strings.IndexAny(s, wsChars+",]}")
	if end < 0 {
		return s, ""
	}
	return s[:end], s[end:]
}

// parseIntLiteral acepta 1_000 y los prefijos 0x / 0o / 0b.
func parseIntLiteral(tok string) (float64, bool) {
	t := strings.ReplaceAll(tok, "_", "")
	sign, mag := 1.0, t
	if len(t) > 1 && (t[0] == '+' || t[0] == '-') {
		if t[0] == '-' {
			sign = -1
		}
		mag = t[1:]
	}
	base := baseOf(mag)
	if base == 0 {
		if t == tok { // sin '_' no hay nada que normalizar: lo resuelve ParseFloat
			return 0, false
		}
		n, err := strconv.ParseFloat(t, 64)
		return n, err == nil
	}
	n, err := strconv.ParseInt(mag[2:], base, 64)
	if err != nil {
		return 0, false
	}
	return sign * float64(n), true
}

func baseOf(mag string) int {
	switch {
	case strings.HasPrefix(mag, "0x"):
		return 16
	case strings.HasPrefix(mag, "0o"):
		return 8
	case strings.HasPrefix(mag, "0b"):
		return 2
	}
	return 0
}

// readDateTime detecta fechas y horas RFC 3339 y las devuelve como string. La
// variante local ("2026-01-01 10:00:00") lleva un espacio en el medio, así que
// puede continuar en el token siguiente.
func readDateTime(tok, rest string) (string, string, bool) {
	if !isDateOrTime(tok) {
		return "", "", false
	}
	if len(tok) == 10 && strings.HasPrefix(rest, " ") {
		if next, after := splitToken(rest[1:]); isDateOrTime(next) {
			return tok + " " + next, after, true
		}
	}
	return tok, rest, true
}

func isDateOrTime(tok string) bool {
	if len(tok) >= 10 && digits(tok[0:4]) && tok[4] == '-' && digits(tok[5:7]) && tok[7] == '-' && digits(tok[8:10]) {
		return true
	}
	return len(tok) >= 8 && digits(tok[0:2]) && tok[2] == ':' && digits(tok[3:5]) && tok[5] == ':' && digits(tok[6:8])
}

func digits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
