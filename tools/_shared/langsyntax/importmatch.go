package langsyntax

import "strings"

// startsImport indica si la línea (ya sin indentación) abre una declaración de
// import para la familia dada. Las exclusiones se evalúan primero.
func startsImport(line string, fam family) bool {
	for _, p := range importNegPairs[fam] {
		if startsWithWord(line, p[0]) && strings.Contains(line, p[1]) {
			return false
		}
	}
	for _, p := range importPrefixes[fam] {
		if startsWithWord(line, p) {
			return true
		}
	}
	for _, p := range importPairs[fam] {
		if startsWithWord(line, p[0]) && strings.Contains(line, p[1]) {
			return true
		}
	}
	return false
}

// startsWithWord indica si line empieza con w como palabra completa: lo que
// sigue no puede continuar el identificador (`importedValue` no es `import`).
func startsWithWord(line, w string) bool {
	if !strings.HasPrefix(line, w) {
		return false
	}
	rest := line[len(w):]
	return rest == "" || !isWordByte(rest[0])
}

// isWordByte indica si el byte puede formar parte de un identificador.
func isWordByte(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
