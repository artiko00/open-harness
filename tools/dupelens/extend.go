package main

// extend lleva la semilla s a la corrida máxima de ventanas coincidentes sobre
// su diagonal, hacia atrás y hacia adelante, mientras las ventanas de ambos
// archivos tengan los mismos tokens y sigan pasando los filtros de ruido. Así
// un par se reporta completo aunque el ancla de su clase cambie a mitad del
// bloque. Devuelve la primera y la última ventana de la corrida, en A.
func extend(files []fileData, v *vocab, s seed, w int, norm bool) (first, last int) {
	fa, fb, d := files[s.a], files[s.b], int(diag(s))
	match := func(p int) bool {
		q := p + d
		return p >= 0 && q >= 0 && p+w <= len(fa.ids) && q+w <= len(fb.ids) &&
			sameWindow(v, files, rec{f: s.a, i: uint32(p)}, rec{f: s.b, i: uint32(q)}, w, norm) &&
			keepWindow(v, fa, p, w, norm) && keepWindow(v, fb, q, w, norm)
	}
	first, last = int(s.ia), int(s.ia)
	for match(first - 1) {
		first--
	}
	for match(last + 1) {
		last++
	}
	return first, last
}

// keepWindow aplica a una ventana los mismos filtros de ruido que el
// prefiltro: descarta las monótonas y, en la vista normalizada, las de baja
// entropía en el fuente crudo.
func keepWindow(v *vocab, f fileData, p, w int, norm bool) bool {
	mono := true
	for k := 1; k < w && mono; k++ {
		mono = v.at(f, p+k, norm) == v.at(f, p, norm)
	}
	return !mono && !(norm && lowEntropyWindow(f, p, w))
}
