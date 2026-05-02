package main

// Fingerprint representa una ventana de tokens con su hash Rabin-Karp.
type Fingerprint struct {
	Hash      uint64
	StartLine int
	EndLine   int
	Window    []string
}

// rkBase y rkMod son la base y módulo del rolling hash.
// Valores estándar para Rabin-Karp con baja probabilidad de colisión.
const (
	rkBase uint64 = 257
	rkMod  uint64 = 1_000_000_007
)

// fingerprint genera todas las ventanas de tamaño windowSize sobre los tokens
// con su hash Rabin-Karp asociado. Vacío si tokens < windowSize.
//
// IMPLEMENTACIÓN PENDIENTE (scaffold):
// La implementación completa del rolling hash queda como TODO para la
// próxima iteración. Por ahora retorna nil para que el binario compile.
//
// Algoritmo planeado:
//  1. h = sum(tokens[0..windowSize-1]) usando base y mod
//  2. para cada ventana i+1: h = (h - tokens[i]*pow) * base + tokens[i+windowSize]
//  3. emitir Fingerprint con hash y rango de líneas
func fingerprint(tokens []Token, windowSize int) []Fingerprint {
	if len(tokens) < windowSize || windowSize <= 0 {
		return nil
	}
	// TODO: implementar Rabin-Karp rolling hash
	return nil
}

// hashToken aplica un hash determinístico a un token individual.
// Usado por fingerprint() para construir el hash de ventana.
func hashToken(token string) uint64 {
	var h uint64
	for _, r := range token {
		h = (h*rkBase + uint64(r)) % rkMod
	}
	return h
}
