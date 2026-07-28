package main

// FileResult es la medición de un archivo: líneas de código, líneas físicas y
// profundidad de anidamiento, con sus umbrales.
type FileResult struct {
	RelPath    string
	Lines      int // líneas de código (sin comentarios ni blancos)
	Physical   int // líneas físicas totales del archivo
	Nesting    int // profundidad máxima de anidamiento
	MaxLines   int
	MaxNesting int // umbral de anidamiento (0 = desactivado)
}

// IsViolation indica si el archivo excede el límite de líneas de código o, con
// el umbral activo (>0), la profundidad de anidamiento.
func (r FileResult) IsViolation() bool {
	if r.Lines > r.MaxLines {
		return true
	}
	return r.MaxNesting > 0 && r.Nesting > r.MaxNesting
}
