package main

// Match representa un par de bloques duplicados detectados. Los campos fileID*
// y startIdx* son internos: ubican el bloque en los slices de tokens para la
// clasificación exact/renamed. Kind es "exact" (tokens idénticos) o "renamed"
// (misma estructura, identificadores distintos).
type Match struct {
	FileA      string
	StartLineA int
	EndLineA   int
	FileB      string
	StartLineB int
	EndLineB   int
	Tokens     int
	Kind       string

	fileIDA   int
	startIdxA int
	fileIDB   int
	startIdxB int
}

// fileData agrupa, por archivo, sus tokens crudos y normalizados. Se guardan
// una sola vez; los fingerprints referencian posiciones dentro de estos slices.
type fileData struct {
	name string
	raw  []Token
	norm []Token
}

// findDuplicates corre dos pasadas: exact (sobre tokens crudos) y renamed
// (sobre tokens normalizados). Los matches renamed que ya cubre un match exact
// se descartan para no reportarlos dos veces. Output ordenado deterministicamente.
func findDuplicates(files []fileData, rawFps, normFps []Fingerprint, windowSize, minLines, minTokens int) []Match {
	exact := detect(files, rawFps, false, windowSize, minLines, minTokens, "exact")
	renamed := detect(files, normFps, true, windowSize, minLines, minTokens, "renamed")
	renamed = dropCoveredByExact(renamed, exact)
	out := append(exact, renamed...)
	sortMatches(out)
	return out
}

// detect agrupa fingerprints por hash, verifica colisiones por índice (contra
// tokens crudos o normalizados según useNorm), fusiona solapamientos y filtra
// por líneas y tokens mínimos. Etiqueta cada match con kind.
func detect(files []fileData, fps []Fingerprint, useNorm bool, windowSize, minLines, minTokens int, kind string) []Match {
	byHash := groupByHash(fps)
	type pairKey struct{ A, B string }
	groups := make(map[pairKey][]Match)
	for _, ents := range byHash {
		for i := 0; i < len(ents); i++ {
			for j := i + 1; j < len(ents); j++ {
				a, b := ents[i], ents[j]
				if a.FileID == b.FileID && a.StartIdx == b.StartIdx {
					continue
				}
				ta := pick(files[a.FileID], useNorm)
				tb := pick(files[b.FileID], useNorm)
				if !sameTokens(ta, a.StartIdx, tb, b.StartIdx, windowSize) {
					continue
				}
				m := canonicalMatch(files, a, b, windowSize)
				groups[pairKey{m.FileA, m.FileB}] = append(groups[pairKey{m.FileA, m.FileB}], m)
			}
		}
	}
	var out []Match
	for _, ms := range groups {
		out = append(out, mergeOnePair(ms, windowSize)...)
	}
	out = filterByMinLines(out, minLines)
	out = filterByMinTokens(out, minTokens)
	for i := range out {
		out[i].Kind = kind
	}
	return out
}
