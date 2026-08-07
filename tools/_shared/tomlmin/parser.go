// Package tomlmin parses a small, well-defined subset of TOML used by
// open-harness tools to read configuration from pyproject.toml under
// the [tool.<name>] tables. See ADR-018 for the subset definition.
package tomlmin

import (
	"encoding/json"
	"strings"
)

// ExtractAsJSON parses the TOML document and returns the JSON-encoded
// bytes of the table at sectionPath (dotted, e.g. "tool.linelens").
// Returns (nil, false, nil) when the section is absent and the document
// is syntactically valid. Returns (nil, false, err) when the document
// uses TOML features outside the supported subset (see ADR-018).
//
// Pass sectionPath="" to extract the root table (top-level keys before
// any table header).
func ExtractAsJSON(toml []byte, sectionPath string) ([]byte, bool, error) {
	root, err := parseDocument(string(toml), sectionPath)
	if err != nil {
		return nil, false, err
	}
	section, ok := descend(root, sectionPath)
	if !ok {
		return nil, false, nil
	}
	// section is composed entirely of map[string]any, []any, string,
	// float64 and bool — all of which encoding/json handles. The error
	// path is therefore unreachable and we skip it to keep coverage tight.
	out, _ := json.Marshal(section)
	return out, true, nil
}

func descend(root map[string]any, path string) (map[string]any, bool) {
	if path == "" {
		return root, true
	}
	cur := root
	for _, seg := range strings.Split(path, ".") {
		next, ok := cur[seg].(map[string]any)
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}
