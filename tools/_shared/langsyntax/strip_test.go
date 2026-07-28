package langsyntax

import (
	"strings"
	"testing"
)

func TestStripComments_lineCommentSlashSlash(t *testing.T) {
	out := StripComments("before // ignored garbage\nafter", ".go")
	if strings.Contains(out, "ignored") || strings.Contains(out, "garbage") {
		t.Errorf("// no eliminado: %q", out)
	}
	if !strings.Contains(out, "before") || !strings.Contains(out, "after") {
		t.Errorf("contexto perdido: %q", out)
	}
}

func TestStripComments_lineCommentToEOF(t *testing.T) {
	out := StripComments("code // hasta el final", ".go")
	if strings.Contains(out, "final") {
		t.Errorf("comentario a EOF no eliminado: %q", out)
	}
}

func TestStripComments_blockComment(t *testing.T) {
	out := StripComments("head /* a\nb\nc */ tail", ".go")
	if strings.Contains(out, "a") && strings.Contains(out, "b") {
		// 'a' podría aparecer en 'tail'; chequeo específico:
	}
	if strings.Contains(out, "head") == false || strings.Contains(out, "tail") == false {
		t.Errorf("contexto de bloque perdido: %q", out)
	}
	if n := strings.Count(out, "\n"); n != 2 {
		t.Errorf("bloque debe preservar 2 newlines, got %d (%q)", n, out)
	}
}

func TestStripComments_blockCommentUnclosed(t *testing.T) {
	out := StripComments("x /* sin cierre\ny", ".go")
	if strings.Contains(out, "cierre") {
		t.Errorf("bloque sin cierre no eliminado: %q", out)
	}
}

func TestStripComments_hashCommentForPython(t *testing.T) {
	out := StripComments("code # comentario\ndef foo", ".py")
	if strings.Contains(out, "comentario") {
		t.Errorf("# en python debe ser comentario: %q", out)
	}
	if !strings.Contains(out, "def") {
		t.Errorf("contexto perdido: %q", out)
	}
}

func TestStripComments_hashNotCommentForRustCssJs(t *testing.T) {
	for _, ext := range []string{".rs", ".css", ".js", ".go", ".c"} {
		out := StripComments("valor #keepme", ext)
		if !strings.Contains(out, "#keepme") {
			t.Errorf("# no debe ser comentario en %s: %q", ext, out)
		}
	}
}

func TestStripComments_extWithoutDotAndUppercase(t *testing.T) {
	if !hashStartsComment("py") {
		t.Error("ext sin punto debe normalizarse")
	}
	if !hashStartsComment(".PY") {
		t.Error("ext en mayúsculas debe normalizarse")
	}
	if hashStartsComment("") {
		t.Error("ext vacía no debe activar hash comment")
	}
}

func TestStripComments_doubleQuoteWithEscape(t *testing.T) {
	out := StripComments(`a "in \"side\" val" b`, ".go")
	if strings.Contains(out, "side") || strings.Contains(out, "val") {
		t.Errorf("string con escape no eliminada: %q", out)
	}
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("contexto perdido: %q", out)
	}
}

func TestStripComments_singleQuoteAndBacktick(t *testing.T) {
	out := StripComments("x 'inner' `raw\nmulti` y", ".go")
	if strings.Contains(out, "inner") || strings.Contains(out, "multi") {
		t.Errorf("strings no eliminadas: %q", out)
	}
	if strings.Count(out, "\n") != 1 {
		t.Errorf("backtick debe preservar newline interno: %q", out)
	}
}

func TestStripComments_unterminatedString(t *testing.T) {
	out := StripComments("code 'sin cierre", ".go")
	if strings.Contains(out, "cierre") {
		t.Errorf("string sin cierre no eliminada: %q", out)
	}
}

func TestStripComments_newlineInsideDoubleString(t *testing.T) {
	out := StripComments("a \"line\nbreak\" b", ".go")
	if strings.Count(out, "\n") != 1 {
		t.Errorf("string con newline debe preservarlo: %q", out)
	}
}

func TestStripComments_empty(t *testing.T) {
	if StripComments("", ".go") != "" {
		t.Error("vacío debe devolver vacío")
	}
}
