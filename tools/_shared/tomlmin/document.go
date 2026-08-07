package tomlmin

import (
	"fmt"
	"strings"
)

// parseDocument recorre el documento TOML materializando tablas y arrays de
// tablas en un árbol de map[string]any.
//
// target es la sección que el llamador quiere extraer. Dentro de ella (y de sus
// sub-tablas) una asignación no parseable es un error: la config del usuario
// está mal y debe enterarse. Fuera de ella se descarta en silencio, porque el
// resto del pyproject.toml no es asunto nuestro y no debe tumbar la carga.
func parseDocument(src, target string) (map[string]any, error) {
	lines, err := splitLogicalLines(src)
	if err != nil {
		return nil, err
	}
	root := map[string]any{}
	cur := root
	curPath := ""

	for _, ll := range lines {
		if isTableHeader(ll.text) {
			next, path, err := openTable(root, ll.text)
			if err != nil {
				return nil, fmt.Errorf("tomlmin: line %d: %w", ll.line, err)
			}
			cur, curPath = next, path
			continue
		}
		if err := applyAssignment(cur, ll.text); err != nil {
			if !withinTarget(curPath, target) {
				continue
			}
			return nil, fmt.Errorf("tomlmin: line %d (in %q): %w", ll.line, curPath, err)
		}
	}
	return root, nil
}

// withinTarget indica si la tabla actual es la sección buscada o una sub-tabla
// suya. Con target vacío (tabla raíz) solo cuentan las claves previas al primer
// encabezado.
func withinTarget(curPath, target string) bool {
	if target == "" {
		return curPath == ""
	}
	return curPath == target || strings.HasPrefix(curPath, target+".")
}

func isTableHeader(line string) bool {
	return strings.HasPrefix(line, "[")
}

// openTable processes "[a.b.c]" or "[[a.b.c]]" and returns the table
// where subsequent assignments should go, plus the dotted path.
func openTable(root map[string]any, line string) (map[string]any, string, error) {
	isArray := strings.HasPrefix(line, "[[")
	open, close := "[", "]"
	if isArray {
		open, close = "[[", "]]"
	}
	if !strings.HasSuffix(line, close) {
		return nil, "", fmt.Errorf("unclosed table header: %s", line)
	}
	path := strings.TrimSpace(line[len(open) : len(line)-len(close)])
	if path == "" {
		return nil, "", fmt.Errorf("empty table name")
	}
	for _, seg := range strings.Split(path, ".") {
		if strings.TrimSpace(seg) == "" {
			return nil, "", fmt.Errorf("empty table segment in %q", path)
		}
	}
	if isArray {
		return mountArrayOfTables(root, path), path, nil
	}
	return mountTable(root, path), path, nil
}

func mountTable(root map[string]any, path string) map[string]any {
	cur := root
	for _, seg := range strings.Split(path, ".") {
		next, ok := cur[seg].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[seg] = next
		}
		cur = next
	}
	return cur
}

func mountArrayOfTables(root map[string]any, path string) map[string]any {
	parts := strings.Split(path, ".")
	parent := mountTableSlice(root, parts[:len(parts)-1])
	last := parts[len(parts)-1]
	arr, _ := parent[last].([]any)
	entry := map[string]any{}
	parent[last] = append(arr, entry)
	return entry
}

func mountTableSlice(root map[string]any, parts []string) map[string]any {
	cur := root
	for _, seg := range parts {
		next, ok := cur[seg].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[seg] = next
		}
		cur = next
	}
	return cur
}
