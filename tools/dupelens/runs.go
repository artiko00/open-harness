package main

import (
	"cmp"
	"slices"
)

// mergeRuns convierte las semillas en hallazgos. Ordenadas por (par de
// archivos, diagonal ib−ia, ia), cada semilla se extiende hasta su clon máximo
// (ver extend); las que ya cubrió una extensión de la misma diagonal se saltan,
// así que cada tramo se recorre una sola vez. Las corridas de un mismo par
// pasan por la fusión por líneas de mergeOnePair.
func mergeRuns(files []fileData, v *vocab, seeds []seed, w int, norm bool, g memGuard) ([]Match, error) {
	slices.SortFunc(seeds, func(x, y seed) int {
		return cmp.Or(cmp.Compare(x.a, y.a), cmp.Compare(x.b, y.b),
			cmp.Compare(diag(x), diag(y)), cmp.Compare(x.ia, y.ia))
	})
	var out, pair []Match
	var err error
	for lo := 0; lo < len(seeds); {
		hi, covered := lo, -1
		for ; hi < len(seeds) && seeds[hi].a == seeds[lo].a && seeds[hi].b == seeds[lo].b; hi++ {
			s := seeds[hi]
			if hi > lo && diag(s) == diag(seeds[hi-1]) && int(s.ia) <= covered {
				continue
			}
			first, last := extend(files, v, s, w, norm)
			covered = last
			pair = append(pair, runMatch(files, s, first, last, w))
		}
		merged := mergeOnePair(pair)
		for len(out)+len(merged) > cap(out) {
			if out, err = grow(out, g); err != nil {
				return nil, err
			}
		}
		out = append(out, merged...)
		pair = pair[:0]
		lo = hi
	}
	return out, nil
}

func diag(s seed) int64 { return int64(s.ib) - int64(s.ia) }

// runMatch arma el Match de la corrida de ventanas [first, last] de la
// diagonal de s: abarca desde la primera ventana hasta el último token de la
// última.
func runMatch(files []fileData, s seed, first, last, w int) Match {
	fa, fb, d := files[s.a], files[s.b], int(diag(s))
	endA := last + w - 1
	return Match{
		FileA: fa.name, StartLineA: int(fa.lines[first]), EndLineA: int(fa.lines[endA]),
		FileB: fb.name, StartLineB: int(fb.lines[first+d]), EndLineB: int(fb.lines[endA+d]),
		Tokens: endA - first + 1, startIdxA: first, endIdxA: endA + 1,
	}
}
