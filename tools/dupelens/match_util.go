package main

// groupByHash agrupa fingerprints por su hash Rabin-Karp.
func groupByHash(fps []Fingerprint) map[uint64][]Fingerprint {
	byHash := make(map[uint64][]Fingerprint)
	for _, fp := range fps {
		byHash[fp.Hash] = append(byHash[fp.Hash], fp)
	}
	return byHash
}

// pick devuelve los tokens normalizados o crudos del archivo según useNorm.
func pick(f fileData, useNorm bool) []Token {
	if useNorm {
		return f.norm
	}
	return f.raw
}

// sameTokens compara n tokens por índice sobre dos slices. Verifica colisiones
// del rolling hash sin necesidad de haber copiado las ventanas. El chequeo de
// límites protege ante rangos fuera de slice (defensa; devuelve false).
func sameTokens(a []Token, ai int, b []Token, bi, n int) bool {
	if ai+n > len(a) || bi+n > len(b) {
		return false
	}
	for k := 0; k < n; k++ {
		if a[ai+k].Value != b[bi+k].Value {
			return false
		}
	}
	return true
}

// canonicalMatch ordena (FileA, FileB) alfabéticamente para evitar duplicados
// inversos y simplificar el merge por par, arrastrando fileID/startIdx.
func canonicalMatch(files []fileData, a, b Fingerprint, windowSize int) Match {
	fa, fb := files[a.FileID].name, files[b.FileID].name
	if fa > fb || (fa == fb && a.StartLine > b.StartLine) {
		a, b = b, a
		fa, fb = fb, fa
	}
	return Match{
		FileA: fa, StartLineA: a.StartLine, EndLineA: a.EndLine, fileIDA: a.FileID, startIdxA: a.StartIdx,
		FileB: fb, StartLineB: b.StartLine, EndLineB: b.EndLine, fileIDB: b.FileID, startIdxB: b.StartIdx,
		Tokens: windowSize,
	}
}
