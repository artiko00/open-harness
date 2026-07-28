package main

import "strings"

// parseNameOnly convierte la salida de `git diff --name-only` en rutas. Cada
// línea es una ruta; git entrecomilla y aplica quoting octal a las que tienen
// caracteres no ASCII o de control. Una salida vacía da cero rutas.
func parseNameOnly(out []byte) []string {
	var paths []string
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		if len(line) >= 2 && line[0] == '"' && line[len(line)-1] == '"' {
			line = unquotePath(line[1 : len(line)-1])
		}
		paths = append(paths, line)
	}
	return paths
}

// unquotePath decodifica el quoting estilo C de git: escapes octales \NNN y
// escapes de un carácter (\t, \n, \", \\, …).
func unquotePath(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' || i+1 >= len(s) {
			b.WriteByte(c)
			continue
		}
		n := s[i+1]
		if n >= '0' && n <= '7' {
			val, j := 0, i+1
			for k := 0; k < 3 && j < len(s) && s[j] >= '0' && s[j] <= '7'; k++ {
				val = val*8 + int(s[j]-'0')
				j++
			}
			b.WriteByte(byte(val))
			i = j - 1
			continue
		}
		b.WriteByte(unescape(n))
		i++
	}
	return b.String()
}

func unescape(n byte) byte {
	switch n {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	default:
		return n
	}
}
