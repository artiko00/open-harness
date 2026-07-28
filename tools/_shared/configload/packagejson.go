// Package configload factoriza los cargadores de configuracion por seccion
// (package.json, composer.json y pyproject.toml) compartidos por los tools
// de open-harness. Cada funcion lee un archivo del directorio dado, extrae
// la seccion del tool y la deserializa en el tipo T. Cuando el archivo o la
// seccion no existen devuelve (zero, false, nil).
package configload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PackageJSON lee dir/package.json y deserializa la seccion key en T.
// Devuelve (zero, false, nil) si el archivo o la seccion no existen.
func PackageJSON[T any](dir, key string) (T, bool, error) {
	var zero T
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("leyendo %q: %w", path, err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return zero, false, fmt.Errorf("parseando %q: %w", path, err)
	}

	section, ok := raw[key]
	if !ok {
		return zero, false, nil
	}

	var cfg T
	if err := json.Unmarshal(section, &cfg); err != nil {
		return zero, false, fmt.Errorf("parseando %q.%s: %w", path, key, err)
	}
	return cfg, true, nil
}
