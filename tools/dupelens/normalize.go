package main

// normalizeTokens produce una copia de tokens con los identificadores
// reemplazados por "ID" y los literales numéricos por "NUM", preservando
// keywords y operadores. Habilita la detección de clones Type-2 (misma
// estructura, identificadores distintos). Conserva la línea de cada token.
func normalizeTokens(tokens []Token) []Token {
	out := make([]Token, len(tokens))
	for i, t := range tokens {
		out[i] = Token{Value: normalizeValue(t.Value), Line: t.Line}
	}
	return out
}

// normalizeValue mapea un token a su forma normalizada.
func normalizeValue(v string) string {
	if isKeyword(v) {
		return v
	}
	if isNumberLiteral(v) {
		return "NUM"
	}
	if isIdentifier(v) {
		return "ID"
	}
	return v
}

// isNumberLiteral: heurística — el token empieza con dígito.
func isNumberLiteral(v string) bool {
	return len(v) > 0 && v[0] >= '0' && v[0] <= '9'
}

// isIdentifier: empieza con letra o '_' y el resto es alfanumérico o '_'.
func isIdentifier(v string) bool {
	for i, r := range v {
		letter := r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		digit := r >= '0' && r <= '9'
		if i == 0 && !letter {
			return false
		}
		if i > 0 && !letter && !digit {
			return false
		}
	}
	return len(v) > 0
}

// fingerprintCode genera fingerprints descartando ventanas monótonas (todos los
// tokens iguales), que no aportan señal estructural y disparan falsos positivos
// en contenido repetitivo: literales de datos, tablas, y —tras stripear los
// contenidos de string— mapas de keywords que quedan como una tira de "true".
// Se aplica tanto a la pasada exacta (tokens crudos) como a la renamed
// (normalizados, p. ej. "ID ID ID …").
func fingerprintCode(tokens []Token, fileID, windowSize int) []Fingerprint {
	all := fingerprint(tokens, fileID, windowSize)
	out := all[:0]
	for _, fp := range all {
		if !monotoneWindow(tokens, fp.StartIdx, windowSize) {
			out = append(out, fp)
		}
	}
	return out
}

// fingerprintNormalized fingerprintea los tokens normalizados (clones Type-2),
// filtrando ventanas monótonas igual que la pasada exacta.
func fingerprintNormalized(norm []Token, fileID, windowSize int) []Fingerprint {
	return fingerprintCode(norm, fileID, windowSize)
}

// monotoneWindow indica si todos los tokens de la ventana tienen el mismo valor.
func monotoneWindow(t []Token, start, w int) bool {
	for k := 1; k < w; k++ {
		if t[start+k].Value != t[start].Value {
			return false
		}
	}
	return true
}
