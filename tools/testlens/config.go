package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/artiko00/open-harness/tools/_shared/configload"
)

type Config struct {
	Language string   `json:"language"`
	Exclude  []string `json:"exclude"`
	NoTest   []string `json:"notest"`
}

var defaultConfig = Config{
	Language: "auto",
	Exclude: []string{
		"node_modules", ".git", "vendor", "dist", "build",
		"coverage", "__pycache__", "target", ".next", ".nuxt",
		"out", ".cache", "testdata",
	},
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return loadConfigFallbackChain(filepath.Dir(path))
		}
		return defaultConfig, err
	}

	var cfg Config
	unknown, err := configload.Strict(data, &cfg)
	if err != nil {
		return defaultConfig, fmt.Errorf("config: %w", err)
	}
	if unknown != "" {
		fmt.Fprintln(os.Stderr, unknownKeyWarning(path, unknown))
	}
	applyConfigDefaults(&cfg)
	return cfg, nil
}

// unknownKeyWarning arma el aviso de clave desconocida; sugiere "exclude" ante
// la clave legada "skip".
func unknownKeyWarning(path, key string) string {
	if key == "skip" {
		return fmt.Sprintf("warning: config %q: clave desconocida %q (ignorada; usá \"exclude\")", path, key)
	}
	return fmt.Sprintf("warning: config %q: clave desconocida %q (ignorada)", path, key)
}

// wrapConfigError envuelve el error del parseo con contexto y, para la clave
// legada "skip", sugiere el nombre correcto "exclude".
func applyConfigDefaults(cfg *Config) {
	if cfg.Language == "" {
		cfg.Language = defaultConfig.Language
	}
	if len(cfg.Exclude) == 0 {
		cfg.Exclude = defaultConfig.Exclude
	}
}
