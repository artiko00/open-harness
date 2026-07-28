package main

// Fingerprint identifica una ventana de tokens por su hash Rabin-Karp y por su
// posición (FileID + StartIdx) en el slice de tokens del archivo. NO copia los
// tokens: la verificación literal se hace por índice contra el slice del
// archivo (ver sameTokens), de modo que la memoria es proporcional al fuente y
// no a N*windowSize.
type Fingerprint struct {
	Hash      uint64
	FileID    int
	StartIdx  int
	StartLine int
	EndLine   int
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
// fileID identifica el archivo dueño de estos tokens.
func fingerprint(tokens []Token, fileID, windowSize int) []Fingerprint {
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
	out = append(out, makeFP(tokens, fileID, 0, windowSize, h))

	// Sliding window: restar saliente (sumando rkMod evita underflow) y sumar entrante.
	for i := 1; i <= len(tokens)-windowSize; i++ {
		h = (h + rkMod - (hashes[i-1]*pow)%rkMod) % rkMod
		h = (h*rkBase + hashes[i+windowSize-1]) % rkMod
		out = append(out, makeFP(tokens, fileID, i, windowSize, h))
	}

	return out
}

// makeFP construye un Fingerprint con su posición y rango de líneas (sin copiar tokens).
func makeFP(tokens []Token, fileID, start, size int, h uint64) Fingerprint {
	return Fingerprint{
		Hash:      h,
		FileID:    fileID,
		StartIdx:  start,
		StartLine: tokens[start].Line,
		EndLine:   tokens[start+size-1].Line,
	}
}

// hashToken aplica un hash determinístico a un token individual.
// Polinomial sobre runes con base/mod del Rabin-Karp. Pura y sin estado: no
// memoiza para mantener la memoria proporcional al fuente (una caché global
// crecería sin cota con identificadores únicos, rompiendo esa garantía).
func hashToken(token string) uint64 {
	var h uint64
	for _, r := range token {
		h = (h*rkBase + uint64(r)) % rkMod
	}
	return h
}
