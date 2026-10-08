package main

import (
	"cmp"
	"slices"
)

// rec es una ventana candidata: su hash y su posición (archivo, inicio).
type rec struct {
	h    uint64
	f, i uint32
}

// candidates devuelve, ordenadas por (hash, archivo, inicio), las ventanas cuyo
// hash aparece más de una vez. Una primera pasada marca cada hash en dos
// bitsets ("visto" y "visto dos veces"); la segunda materializa solo las
// repetidas y les aplica los filtros de ruido. La mayoría de las ventanas de un
// proyecto son únicas y nunca ocupan memoria. Un falso positivo del bitset solo
// agrega una ventana que formará un grupo de un miembro, sin semillas.
func candidates(files []fileData, v *vocab, w int, norm bool, g memGuard) ([]rec, error) {
	total := 0
	for _, f := range files {
		total += len(f.ids)
	}
	words := 1
	for words*64 < total*8 {
		words *= 2
	}
	if err := g.reserve(uint64(words) * 16); err != nil {
		return nil, err
	}
	seen, twice := make([]uint64, words), make([]uint64, words)
	mask := uint64(words*64 - 1)
	for _, f := range files {
		eachWindow(v, f, w, norm, func(h uint64, _ int, _ bool) {
			b := h & mask
			twice[b>>6] |= seen[b>>6] & (1 << (b & 63))
			seen[b>>6] |= 1 << (b & 63)
		})
	}
	var recs []rec
	var err error
	for fi, f := range files {
		eachWindow(v, f, w, norm, func(h uint64, i int, mono bool) {
			b := h & mask
			if err != nil || mono || twice[b>>6]&(1<<(b&63)) == 0 || (norm && lowEntropyWindow(f, i, w)) {
				return
			}
			if len(recs) == cap(recs) {
				if recs, err = grow(recs, g); err != nil {
					return
				}
			}
			recs = append(recs, rec{h, uint32(fi), uint32(i)})
		})
		if err == nil {
			err = g.check()
		}
		if err != nil {
			return nil, err
		}
	}
	slices.SortFunc(recs, func(a, b rec) int {
		return cmp.Or(cmp.Compare(a.h, b.h), cmp.Compare(a.f, b.f), cmp.Compare(a.i, b.i))
	})
	return recs, nil
}
