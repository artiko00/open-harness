// Package langsyntax centraliza el conocimiento de sintaxis por lenguaje que
// comparten las herramientas del harness (dupelens, linelens): principalmente
// cómo reconocer comentarios y strings para descartarlos del análisis.
package langsyntax

import "strings"

// hashCommentExts son las extensiones donde '#' inicia un comentario de línea
// (Python, Ruby, shell, YAML, TOML, Perl, R, Makefile, Terraform). En Rust,
// CSS, JS, Go y C el '#' es significativo (#[derive], #fff, #private) y NO
// debe tratarse como comentario.
var hashCommentExts = map[string]bool{
	".py": true, ".pyw": true, ".rb": true, ".sh": true, ".bash": true,
	".zsh": true, ".yml": true, ".yaml": true, ".toml": true, ".pl": true,
	".r": true, ".mk": true, ".tf": true,
}

// hashStartsComment indica si '#' inicia un comentario de línea para la
// extensión dada. Acepta la extensión con o sin punto inicial y en cualquier
// caja (".PY", "py", ".py" son equivalentes).
func hashStartsComment(ext string) bool {
	ext = strings.ToLower(ext)
	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}
	return hashCommentExts[ext]
}
