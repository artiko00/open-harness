package tomlmin

import (
	"fmt"
	"strings"
)

// parseDocument walks the TOML document line by line, materializing
// tables and array-of-tables into a nested map[string]any tree.
func parseDocument(src string) (map[string]any, error) {
	root := map[string]any{}
	cur := root
	curPath := ""

	for i, rawLine := range strings.Split(src, "\n") {
		line := stripTrailingComment(rawLine)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if isTableHeader(line) {
			next, path, err := openTable(root, line)
			if err != nil {
				return nil, fmt.Errorf("tomlmin: line %d: %w", i+1, err)
			}
			cur, curPath = next, path
			continue
		}
		if err := applyAssignment(cur, line); err != nil {
			return nil, fmt.Errorf("tomlmin: line %d (in %q): %w", i+1, curPath, err)
		}
	}
	return root, nil
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
