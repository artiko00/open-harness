package main

import "sort"

// mergeOnePair fusiona matches solapantes/contiguos del mismo par de archivos.
// Tokens del merged = windowSize + (numFingerprintsContiguos - 1).
func mergeOnePair(matches []Match, windowSize int) []Match {
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
	count := 1
	for i := 1; i < len(matches); i++ {
		n := matches[i]
		overlapA := n.StartLineA <= cur.EndLineA+1
		overlapB := n.StartLineB <= cur.EndLineB+1
		if overlapA && overlapB {
			if n.EndLineA > cur.EndLineA {
				cur.EndLineA = n.EndLineA
			}
			if n.EndLineB > cur.EndLineB {
				cur.EndLineB = n.EndLineB
			}
			count++
			cur.Tokens = windowSize + count - 1
		} else {
			out = append(out, cur)
			cur = n
			count = 1
		}
	}
	out = append(out, cur)
	return out
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
