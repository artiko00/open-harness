package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/artiko00/open-harness/tools/_shared/configload"
)

type Config struct {
	MaxFiles     int      `json:"maxFiles"`
	Base         string   `json:"base"`
	ExcludeTests bool     `json:"excludeTests"`
	Exclude      []string `json:"exclude"`
}

const defaultMaxFiles = 15

var defaultExclude = []string{
	// Comunes
	".git/**", "node_modules/**", "vendor/**", "dist/**", "build/**", "coverage/**",
	// JS/TS
	"package-lock.json", "pnpm-lock.yaml", "yarn.lock", ".next/**", ".nuxt/**",
	"out/**", "**/__snapshots__/**",
	// Python
	"poetry.lock", "Pipfile.lock", "uv.lock", "**/__pycache__/**", "*.egg-info/**", ".venv/**",
	// Go
	"go.sum", "**/*.pb.go", "**/zz_generated*.go",
}

// loadConfig resuelve scopelens.json en dir; si no existe, recorre la cadena
// (pyproject > package.json > composer). Un archivo presente pero inválido, o
// un maxFiles negativo, devuelven error (mapeado a exit 2 por el llamador).
func loadConfig(dir string) (Config, error) {
	path := filepath.Join(dir, "scopelens.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return loadConfigChain(dir)
		}
		return Config{}, fmt.Errorf("leyendo %q: %w", path, err)
	}

	var cfg Config
	unknown, err := configload.Strict(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("config %q: %w", path, err)
	}
	if unknown != "" {
		fmt.Fprintf(os.Stderr, "warning: config %q: clave desconocida %q (ignorada)\n", path, unknown)
	}
	return validateAndDefault(cfg, path)
}

// validateAndDefault falla ruidoso ante maxFiles negativo y rellena los
// campos ausentes con los defaults compilados.
func validateAndDefault(cfg Config, src string) (Config, error) {
	if cfg.MaxFiles < 0 {
		return Config{}, fmt.Errorf("config %q: maxFiles debe ser >= 0, no %d", src, cfg.MaxFiles)
	}
	if cfg.MaxFiles == 0 {
		cfg.MaxFiles = defaultMaxFiles
	}
	if len(cfg.Exclude) == 0 {
		cfg.Exclude = defaultExclude
	}
	return cfg, nil
}
