package main

import "slices"

// seed empareja la ventana ib del archivo b con la ventana ia del archivo a, la
// primera ocurrencia (canónica) de su clase de contenido.
type seed struct{ a, b, ia, ib uint32 }

// anchorSeeds recorre los grupos de igual hash (recs viene ordenado) y separa
// cada uno en clases de contenido idéntico, verificadas literalmente contra los
// tokens. Dentro de cada clase, cada ocurrencia se empareja solo con la primera
// en orden canónico (archivo, posición): k−1 semillas en lugar de k(k−1)/2
// pares. Como cada ventana aporta como mucho una semilla, el total queda
// acotado por el número de ventanas, y k copias cuestan k, no k².
func anchorSeeds(files []fileData, v *vocab, recs []rec, w int, norm bool, g memGuard) ([]seed, error) {
	var seeds []seed
	var used []bool
	var err error
	for lo := 0; lo < len(recs); {
		hi := lo + 1
		for hi < len(recs) && recs[hi].h == recs[lo].h {
			hi++
		}
		grp := recs[lo:hi]
		used = slices.Grow(used[:0], len(grp))[:len(grp)]
		clear(used)
		for i := range grp {
			for j := i + 1; j < len(grp) && !used[i]; j++ {
				if used[j] || !sameWindow(v, files, grp[i], grp[j], w, norm) {
					continue
				}
				used[j] = true
				if len(seeds) == cap(seeds) {
					if seeds, err = grow(seeds, g); err != nil {
						return nil, err
					}
				}
				seeds = append(seeds, seed{grp[i].f, grp[j].f, grp[i].i, grp[j].i})
			}
		}
		lo = hi
	}
	return seeds, nil
}

// sameWindow verifica literalmente que las ventanas a y b tienen los mismos w
// tokens en la vista pedida: descarta las colisiones del hash.
func sameWindow(v *vocab, files []fileData, a, b rec, w int, norm bool) bool {
	fa, fb := files[a.f], files[b.f]
	for k := 0; k < w; k++ {
		if v.at(fa, int(a.i)+k, norm) != v.at(fb, int(b.i)+k, norm) {
			return false
		}
	}
	return true
}
