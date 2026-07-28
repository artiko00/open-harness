package main

import (
	"regexp"
	"strings"
)

// danglingKey detecta una clave secreta seguida de ':' o '=' sin valor en la
// misma línea (JSON pretty-printed con el valor en la línea siguiente).
var danglingKey = regexp.MustCompile(
	`(?i)(?:secret|password|passwd|api_key|apikey|auth_token|private_key|token|access_token|bearer)["']?\s*[:=]\s*["']?\s*$`)

func isDanglingKey(line string) bool {
	return danglingKey.MatchString(line)
}

// matchLine evalúa las reglas sobre candidate y devuelve los hallazgos. El
// allowlist y la entropía se aplican al VALOR capturado, no a la línea entera.
func matchLine(candidate string, lineNum int, relPath string, compiled []compiledRule, cfg Config) []Finding {
	var out []Finding
	for _, c := range compiled {
		m := c.re.FindStringSubmatch(candidate)
		if m == nil {
			continue
		}
		value := extractValue(m)
		if isAllowed(value, cfg.Allowlist) {
			continue
		}
		if c.rule.EntropyGate && shannonEntropy(value) < cfg.MinEntropy {
			continue
		}
		out = append(out, Finding{
			RelPath:  relPath,
			Line:     lineNum,
			Content:  truncate(strings.TrimSpace(candidate), 80),
			RuleName: c.rule.Name,
			Severity: c.rule.Severity,
		})
	}
	return out
}

// extractValue toma el último grupo de captura no vacío como valor; si la regla
// no tiene grupos, usa la coincidencia completa.
func extractValue(m []string) string {
	for g := len(m) - 1; g >= 1; g-- {
		if m[g] != "" {
			return m[g]
		}
	}
	return m[0]
}
