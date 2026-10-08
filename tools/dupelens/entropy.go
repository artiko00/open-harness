package main

// Umbrales del filtro de baja entropía. Una ventana se considera repetitiva
// cuando abarca al menos minEntropyLines líneas y la proporción de ellas que
// empieza con el mismo token que la primera alcanza maxHeadShare. Son fijos y
// documentados, como monotoneWindow: exponerlos como config multiplicaría las
// combinaciones sin que haya un caso de uso que lo pida.
const (
	minEntropyLines = 3
	maxHeadShare    = 0.75
)

// lowEntropyWindow indica si la ventana [start, start+w) proviene de un bloque
// repetitivo: arrays de datos, tablas literales, listas de constantes. En esos
// bloques cada línea repite la forma de la anterior, así que colisionan con
// cualquier otro bloque de la misma forma aunque no compartan origen.
//
// El criterio se ancla en el primer token de la primera línea de la ventana en
// lugar de la moda de todos los primeros tokens: es más conservador (descarta
// menos), se computa en una pasada sin asignar memoria y cubre igual el caso que motiva el filtro, donde
// todas las líneas del bloque empiezan igual.
//
// Se evalúa sobre los tokens CRUDOS aunque el fingerprint sea el normalizado:
// normalizado, casi todo código real es de baja entropía por construcción
// (`const ID = ID ID`) y el filtro descartaría duplicados legítimos.
func lowEntropyWindow(f fileData, start, w int) bool {
	end := start + w
	i := firstWholeLine(f, start, end)
	if i >= end {
		return false
	}
	head := f.ids[i]
	lines, same, prevLine := 0, 0, uint32(0)
	for ; i < end; i++ {
		if f.lines[i] == prevLine {
			continue
		}
		prevLine = f.lines[i]
		lines++
		if f.ids[i] == head {
			same++
		}
	}
	if lines < minEntropyLines {
		return false
	}
	return float64(same)/float64(lines) >= maxHeadShare
}

// firstWholeLine devuelve el índice del primer token que abre una línea dentro
// de la ventana. Las ventanas deslizantes suelen empezar a mitad de una línea, y
// ahí el token inicial no es el comienzo real de esa línea: tomarlo como ancla
// haría parecer variado un bloque uniforme.
func firstWholeLine(f fileData, start, end int) int {
	if start == 0 || f.lines[start-1] != f.lines[start] {
		return start
	}
	i := start
	for i < end && f.lines[i] == f.lines[start] {
		i++
	}
	return i
}
