package main

import "math/bits"

// Rolling hash Rabin-Karp módulo el primo de Mersenne 2^61−1. Con cientos de
// millones de ventanas, un módulo de ~30 bits produce millones de colisiones
// falsas que inflan los grupos de hash y el prefiltro; con 61 bits son
// despreciables. La verificación literal se conserva igual (ver sameWindow).
const (
	rkBase uint64 = 257
	m61    uint64 = 1<<61 - 1
)

// mulMod multiplica a·b módulo 2^61−1 (a, b < 2^61) con aritmética de 128
// bits: como 2^61 ≡ 1, el producto se reduce sumando sus tramos de 61 bits.
func mulMod(a, b uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	return reduce((hi<<3 | lo>>61) + lo&m61)
}

// addMod suma a+b módulo 2^61−1 (a, b < 2^61).
func addMod(a, b uint64) uint64 { return reduce(a + b) }

// reduce lleva x < 2^63 al rango [0, 2^61−1).
func reduce(x uint64) uint64 {
	x = x&m61 + x>>61
	if x >= m61 {
		x -= m61
	}
	return x
}

// hashToken aplica un hash polinomial sobre los runes del token. Se calcula una
// sola vez por token distinto (ver vocab), no por posición.
func hashToken(token string) uint64 {
	var h uint64
	for _, r := range token {
		h = addMod(mulMod(h, rkBase), uint64(r))
	}
	return h
}

// eachWindow recorre las ventanas de w tokens de f en la vista pedida (cruda o
// normalizada) y entrega a fn el hash de cada una, su inicio y si es monótona
// (todos sus tokens iguales). La monotonía se lleva en O(1) con el largo de la
// racha de tokens iguales que termina en cada posición. No asigna memoria.
func eachWindow(v *vocab, f fileData, w int, norm bool, fn func(h uint64, start int, mono bool)) {
	if w <= 0 || len(f.ids) < w {
		return
	}
	pow := uint64(1)
	for i := 1; i < w; i++ {
		pow = mulMod(pow, rkBase)
	}
	var h uint64
	run := 0
	for i := range f.ids {
		t := v.at(f, i, norm)
		if i >= w {
			h = addMod(h, m61-mulMod(v.hash[v.at(f, i-w, norm)], pow))
		}
		h = addMod(mulMod(h, rkBase), v.hash[t])
		run++
		if i == 0 || t != v.at(f, i-1, norm) {
			run = 1
		}
		if i >= w-1 {
			fn(h, i-w+1, run >= w)
		}
	}
}
