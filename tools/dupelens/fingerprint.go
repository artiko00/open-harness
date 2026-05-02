package main

import "sync"

// Fingerprint representa una ventana de tokens con su hash Rabin-Karp.
type Fingerprint struct {
	Hash      uint64
	StartLine int
	EndLine   int
	Window    []string
}

// rkBase y rkMod son la base y módulo del rolling hash.
// Mod elegido para evitar overflow al multiplicar dos valores ≤ rkMod
// dentro de uint64 (rkMod^2 ≈ 10^18 < uint64 max ≈ 1.8·10^19).
const (
	rkBase uint64 = 257
	rkMod  uint64 = 1_000_000_007
)

// fingerprint genera todas las ventanas de tamaño windowSize sobre los tokens
// con su hash Rabin-Karp asociado. Vacío si tokens<windowSize o windowSize≤0.
//
// Algoritmo:
//  1. Pre-hashear cada token a uint64 con hashToken
//  2. Calcular hash inicial de la primera ventana (suma con base/mod)
//  3. Deslizar: h = (h - tokens[i-1]*pow)*base + tokens[i+w-1]
//  4. Emitir Fingerprint con hash, líneas y copia del window
func fingerprint(tokens []Token, windowSize int) []Fingerprint {
	if windowSize <= 0 || len(tokens) < windowSize {
		return nil
	}

	hashes := make([]uint64, len(tokens))
	for i := range tokens {
		hashes[i] = hashToken(tokens[i].Value)
	}

	// pow = rkBase^(windowSize-1) mod rkMod — usado al desplazar
	var pow uint64 = 1
	for i := 0; i < windowSize-1; i++ {
		pow = (pow * rkBase) % rkMod
	}

	out := make([]Fingerprint, 0, len(tokens)-windowSize+1)

	// Hash de la primera ventana
	var h uint64
	for i := 0; i < windowSize; i++ {
		h = (h*rkBase + hashes[i]) % rkMod
	}
	out = append(out, makeFP(tokens, 0, windowSize, h))

	// Sliding window
	for i := 1; i <= len(tokens)-windowSize; i++ {
		// Restar el token saliente (multiplicado por pow para reposicionarlo)
		// Sumar rkMod previo a restar evita underflow en uint64.
		h = (h + rkMod - (hashes[i-1]*pow)%rkMod) % rkMod
		h = (h*rkBase + hashes[i+windowSize-1]) % rkMod
		out = append(out, makeFP(tokens, i, windowSize, h))
	}

	return out
}

// makeFP construye un Fingerprint copiando los token values y rangos de líneas.
func makeFP(tokens []Token, start, size int, h uint64) Fingerprint {
	w := make([]string, size)
	for j := 0; j < size; j++ {
		w[j] = tokens[start+j].Value
	}
	return Fingerprint{
		Hash:      h,
		StartLine: tokens[start].Line,
		EndLine:   tokens[start+size-1].Line,
		Window:    w,
	}
}

// tokenHashCache memoiza hashToken — los mismos identifiers aparecen
// repetidamente en codebases reales, evitar re-hashear ahorra trabajo.
var tokenHashCache sync.Map

// hashToken aplica un hash determinístico a un token individual.
// Polinomial sobre runes con base/mod del Rabin-Karp, con cache thread-safe.
func hashToken(token string) uint64 {
	if v, ok := tokenHashCache.Load(token); ok {
		return v.(uint64)
	}
	var h uint64
	for _, r := range token {
		h = (h*rkBase + uint64(r)) % rkMod
	}
	tokenHashCache.Store(token, h)
	return h
}
