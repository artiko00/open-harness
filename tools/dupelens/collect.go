package main

import (
	"path/filepath"
	"strings"
)

// addFile tokeniza el contenido y registra el archivo como ids internados y
// líneas: 8 bytes por token. Un archivo con menos tokens que la ventana no
// puede aportar ninguna ventana y no se registra.
func addFile(files *[]fileData, v *vocab, relPath, content string, opts scanOpts) {
	toks := tokenize(content, strings.ToLower(filepath.Ext(relPath)), opts.stripImports)
	if len(toks) < opts.windowSize {
		return
	}
	f := fileData{name: relPath, ids: make([]uint32, len(toks)), lines: make([]uint32, len(toks))}
	for i, t := range toks {
		f.ids[i], f.lines[i] = v.id(t.Value), uint32(t.Line)
	}
	*files = append(*files, f)
}
