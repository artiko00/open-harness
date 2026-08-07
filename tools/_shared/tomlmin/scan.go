package tomlmin

import (
	"fmt"
	"strings"
)

// logicalLine es una asignación o encabezado completo: puede abarcar varias
// líneas físicas cuando lleva un array o una inline table multilínea.
type logicalLine struct {
	text string // sin comentarios; conserva los saltos de línea internos
	line int    // línea física donde empieza, 1-based
}

// splitLogicalLines agrupa las líneas físicas de src en líneas lógicas.
// Los comentarios se descartan y los strings se copian tal cual, de modo que
// un '#', un ']' o un '[tabla]' dentro de un literal no alteran la
// segmentación.
func splitLogicalLines(src string) ([]logicalLine, error) {
	s := &lineScanner{src: src, lineNo: 1, start: 1}
	for s.i < len(src) {
		switch c := src[s.i]; {
		case c == '#':
			s.skipComment()
		case c == '\n':
			s.newline()
		case c == '"' || c == '\'':
			s.copyString()
		case c == '[' || c == '{':
			s.depth++
			s.keep()
		case c == ']' || c == '}':
			if s.depth > 0 {
				s.depth--
			}
			s.keep()
		default:
			s.keep()
		}
	}
	if s.depth != 0 {
		return nil, fmt.Errorf("tomlmin: line %d: unclosed delimiter", s.start)
	}
	s.flush()
	return s.out, nil
}

type lineScanner struct {
	src    string
	i      int
	lineNo int
	start  int
	depth  int
	buf    strings.Builder
	out    []logicalLine
}

func (s *lineScanner) keep() {
	s.buf.WriteByte(s.src[s.i])
	s.i++
}

func (s *lineScanner) skipComment() {
	for s.i < len(s.src) && s.src[s.i] != '\n' {
		s.i++
	}
}

func (s *lineScanner) newline() {
	if s.depth == 0 {
		s.flush()
		s.lineNo++
		s.start = s.lineNo
	} else {
		s.buf.WriteByte('\n')
		s.lineNo++
	}
	s.i++
}

// copyString traslada un literal completo al buffer. Un literal sin cerrar no
// aborta el escaneo: se copia lo que haya y el parseo del valor lo reportará
// como error, que así queda sujeto a la tolerancia por sección.
func (s *lineScanner) copyString() {
	n, newlines := scanStringLiteral(s.src[s.i:])
	s.buf.WriteString(s.src[s.i : s.i+n])
	s.i += n
	s.lineNo += newlines
}

func (s *lineScanner) flush() {
	if t := strings.TrimSpace(s.buf.String()); t != "" {
		s.out = append(s.out, logicalLine{text: t, line: s.start})
	}
	s.buf.Reset()
}
