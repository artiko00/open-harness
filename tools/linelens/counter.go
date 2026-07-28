package main

import (
	"errors"
	"os"

	"github.com/artiko00/open-harness/tools/_shared/langsyntax"
	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// errLineTooLong señala una línea que supera pathmatch.BufferLimit; se traduce
// al motivo canónico ReasonLineTooLong para mantener el contrato de skips.
var errLineTooLong = errors.New("line exceeds buffer limit")

// countFile mide un archivo y devuelve: líneas de CÓDIGO (sin comentarios ni
// blancos), líneas FÍSICAS totales y la profundidad máxima de anidamiento.
// Lee el archivo completo (io.ReadFile) y rechaza cualquier línea que exceda
// BufferLimit con errLineTooLong, preservando el gate de la fase 2.
func countFile(path, ext string) (code, physical, nesting int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, 0, err
	}
	if maxLineLen(data) > pathmatch.BufferLimit {
		return 0, 0, 0, errLineTooLong
	}
	physical = physicalLines(data)
	stripped := langsyntax.StripComments(string(data), ext)
	code = nonBlankLines(stripped)
	nesting = maxNesting(stripped)
	return code, physical, nesting, nil
}

// motivoLectura traduce el error de lectura al motivo canónico de omisión.
func motivoLectura(err error) string {
	if errors.Is(err, errLineTooLong) {
		return pathmatch.ReasonLineTooLong
	}
	return pathmatch.ReasonReadError
}
