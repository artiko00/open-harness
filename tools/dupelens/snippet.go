package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// readSnippet retorna hasta `count` líneas del archivo `relPath` (relativo a `root`)
// empezando en `startLine` (1-indexed). Cada línea va prefijada con "  | " para
// distinguirla del output principal del reporter. Si el archivo no existe o
// startLine queda fuera de rango, retorna "".
func readSnippet(root, relPath string, startLine, count int) string {
	if startLine < 1 || count < 1 {
		return ""
	}
	full := filepath.Join(root, relPath)
	f, err := os.Open(full)
	if err != nil {
		return ""
	}
	defer f.Close()

	var b strings.Builder
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	line := 0
	taken := 0
	for scanner.Scan() {
		line++
		if line < startLine {
			continue
		}
		fmt.Fprintf(&b, "  | %s\n", scanner.Text())
		taken++
		if taken >= count {
			break
		}
	}
	if taken == 0 {
		return ""
	}
	return b.String()
}

// makeSnippetFunc retorna un closure que el reporter usa cuando opts.Snippet
// está configurado. Escapa la dependencia de filesystem del reporter para
// que sea testeable inyectando otros backends.
func makeSnippetFunc(root string) func(file string, startLine, count int) string {
	return func(file string, startLine, count int) string {
		return readSnippet(root, file, startLine, count)
	}
}
