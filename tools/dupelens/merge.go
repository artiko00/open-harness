package main

import "sort"

// mergeOnePair fusiona matches solapantes/contiguos (en líneas) del mismo par
// de archivos: une corridas de diagonales vecinas, p. ej. cuando una copia
// tiene un token de más. Tokens del fusionado = tramo de tokens que abarca en A.
func mergeOnePair(matches []Match) []Match {
	if len(matches) <= 1 {
		return matches
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].StartLineA != matches[j].StartLineA {
			return matches[i].StartLineA < matches[j].StartLineA
		}
		return matches[i].StartLineB < matches[j].StartLineB
	})

	var out []Match
	cur := matches[0]
	for _, n := range matches[1:] {
		if n.StartLineA <= cur.EndLineA+1 && n.StartLineB <= cur.EndLineB+1 {
			cur.EndLineA = max(cur.EndLineA, n.EndLineA)
			cur.EndLineB = max(cur.EndLineB, n.EndLineB)
			cur.startIdxA = min(cur.startIdxA, n.startIdxA)
			cur.endIdxA = max(cur.endIdxA, n.endIdxA)
			cur.Tokens = cur.endIdxA - cur.startIdxA
			continue
		}
		out = append(out, cur)
		cur = n
	}
	return append(out, cur)
}

// filterByMinLines descarta matches cuyo rango (en cualquiera de los archivos)
// es menor al mínimo configurado. minLines≤1 actúa como passthrough.
func filterByMinLines(matches []Match, minLines int) []Match {
	if minLines <= 1 {
		return matches
	}
	var out []Match
	for _, m := range matches {
		linesA := m.EndLineA - m.StartLineA + 1
		linesB := m.EndLineB - m.StartLineB + 1
		if linesA < minLines || linesB < minLines {
			continue
		}
		out = append(out, m)
	}
	return out
}

// filterByMinTokens descarta matches cuyo bloque fusionado tiene menos tokens
// que el umbral de reporte. Separa el ruido de reporte (minTokens) del tamaño
// de ventana de detección (windowSize): un umbral alto no obliga a agrandar la
// ventana. minTokens≤0 actúa como passthrough.
func filterByMinTokens(matches []Match, minTokens int) []Match {
	if minTokens <= 0 {
		return matches
	}
	var out []Match
	for _, m := range matches {
		if m.Tokens >= minTokens {
			out = append(out, m)
		}
	}
	return out
}

// sortMatches ordena por (FileA|FileB, StartLineA) para output reproducible.
func sortMatches(matches []Match) {
	sort.Slice(matches, func(i, j int) bool {
		ki := matches[i].FileA + "|" + matches[i].FileB
		kj := matches[j].FileA + "|" + matches[j].FileB
		if ki != kj {
			return ki < kj
		}
		return matches[i].StartLineA < matches[j].StartLineA
	})
}
