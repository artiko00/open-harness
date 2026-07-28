package main

import (
	"io"
	"os"
	"strings"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// contieneMarcador devuelve true si el archivo contiene al menos uno de los
// marcadores de test del lenguaje (func Test, it(, def test_, #[test]...). Un
// archivo de test vacío o sin marcadores NO cubre (8.15).
func contieneMarcador(path string, markers []string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, pathmatch.BufferLimit)
	n, _ := io.ReadFull(f, buf)
	contenido := string(buf[:n])
	for _, m := range markers {
		if strings.Contains(contenido, m) {
			return true
		}
	}
	return false
}
