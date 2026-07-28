package configload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/artiko00/open-harness/tools/_shared/tomlmin"
)

// Pyproject lee dir/pyproject.toml y deserializa la tabla [section] en T
// usando tomlmin. Devuelve (zero, false, nil) si el archivo o la seccion no
// existen.
func Pyproject[T any](dir, section string) (T, bool, error) {
	var zero T
	path := filepath.Join(dir, "pyproject.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("leyendo %q: %w", path, err)
	}

	raw, found, err := tomlmin.ExtractAsJSON(data, section)
	if err != nil {
		return zero, false, fmt.Errorf("pyproject.toml: %w", err)
	}
	if !found {
		return zero, false, nil
	}

	var cfg T
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return zero, false, fmt.Errorf("parseando pyproject.toml [%s]: %w", section, err)
	}
	return cfg, true, nil
}
