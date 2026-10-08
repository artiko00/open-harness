package main

// La normalización reemplaza identificadores por "ID" y literales numéricos por
// "NUM", preservando keywords y operadores: habilita la detección de clones
// Type-2 (misma estructura, identificadores distintos). Se aplica una vez por
// token distinto, en vocab.

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
