package main

import "math"

// shannonEntropy calcula la entropía de Shannon (bits por carácter) sobre el
// valor capturado. Un valor uniforme (p. ej. "00000000") da 0; un secreto
// aleatorio se acerca a log2(alfabeto).
func shannonEntropy(s string) float64 {
	if s == "" {
		return 0
	}
	freq := make(map[rune]float64)
	var total float64
	for _, r := range s {
		freq[r]++
		total++
	}
	var e float64
	for _, c := range freq {
		p := c / total
		e -= p * math.Log2(p)
	}
	return e
}
