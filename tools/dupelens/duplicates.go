package main

// Match representa un par de bloques duplicados detectados.
type Match struct {
	FileA      string
	StartLineA int
	EndLineA   int
	FileB      string
	StartLineB int
	EndLineB   int
	Tokens     int
}

// findDuplicates agrupa Fingerprints por hash, verifica colisiones con
// comparación literal de Window, fusiona matches solapados (sliding) y
// filtra por número mínimo de líneas. Output ordenado deterministicamente.
func findDuplicates(perFile map[string][]Fingerprint, windowSize, minLines int) []Match {
	if len(perFile) == 0 {
		return nil
	}

	type entry struct {
		File string
		FP   Fingerprint
	}
	byHash := make(map[uint64][]entry)
	for file, fps := range perFile {
		for _, fp := range fps {
			byHash[fp.Hash] = append(byHash[fp.Hash], entry{file, fp})
		}
	}

	type pairKey struct{ A, B string }
	groups := make(map[pairKey][]Match)
	for _, ents := range byHash {
		for i := 0; i < len(ents); i++ {
			for j := i + 1; j < len(ents); j++ {
				a, b := ents[i], ents[j]
				if a.File == b.File && a.FP.StartLine == b.FP.StartLine {
					continue
				}
				if !sameWindow(a.FP.Window, b.FP.Window) {
					continue
				}
				m := canonicalMatch(a.File, a.FP, b.File, b.FP, windowSize)
				k := pairKey{m.FileA, m.FileB}
				groups[k] = append(groups[k], m)
			}
		}
	}

	var out []Match
	for _, ms := range groups {
		out = append(out, mergeOnePair(ms, windowSize)...)
	}
	out = filterByMinLines(out, minLines)
	sortMatches(out)
	return out
}

// canonicalMatch ordena (FileA, FileB) alfabéticamente para evitar duplicados
// inversos y simplificar el merge por par.
func canonicalMatch(fileA string, fpA Fingerprint, fileB string, fpB Fingerprint, windowSize int) Match {
	if fileA > fileB || (fileA == fileB && fpA.StartLine > fpB.StartLine) {
		fileA, fileB = fileB, fileA
		fpA, fpB = fpB, fpA
	}
	return Match{
		FileA: fileA, StartLineA: fpA.StartLine, EndLineA: fpA.EndLine,
		FileB: fileB, StartLineB: fpB.StartLine, EndLineB: fpB.EndLine,
		Tokens: windowSize,
	}
}

// sameWindow compara dos slices de tokens literalmente.
// Necesario para descartar colisiones del rolling hash.
func sameWindow(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
