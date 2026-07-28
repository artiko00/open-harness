package configload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// composerExtra refleja el anidamiento "extra" → "open-harness" → "<tool>"
// dentro de composer.json, la misma forma que usan phpstan y afines.
type composerExtra struct {
	Extra struct {
		OpenHarness map[string]json.RawMessage `json:"open-harness"`
	} `json:"extra"`
}

// Composer lee dir/composer.json y deserializa la seccion extra.open-harness[key]
// en T. Devuelve (zero, false, nil) si el archivo o la seccion no existen.
func Composer[T any](dir, key string) (T, bool, error) {
	var zero T
	path := filepath.Join(dir, "composer.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("leyendo %q: %w", path, err)
	}

	var doc composerExtra
	if err := json.Unmarshal(data, &doc); err != nil {
		return zero, false, fmt.Errorf("parseando %q: %w", path, err)
	}

	section, ok := doc.Extra.OpenHarness[key]
	if !ok {
		return zero, false, nil
	}

	var cfg T
	if err := json.Unmarshal(section, &cfg); err != nil {
		return zero, false, fmt.Errorf("parseando %q extra.open-harness.%s: %w", path, key, err)
	}
	return cfg, true, nil
}
