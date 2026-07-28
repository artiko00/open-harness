package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/artiko00/open-harness/tools/_shared/configload"
)

type Config struct {
	Default DefaultConfig `json:"default"`
	Rules   []Rule        `json:"rules"`
	Exclude []string      `json:"exclude"`
}

type DefaultConfig struct {
	MaxLines   int `json:"maxLines"`
	MaxNesting int `json:"maxNesting"`
}

type Rule struct {
	Pattern  string `json:"pattern"`
	MaxLines int    `json:"maxLines"`
	Skip     bool   `json:"skip"`
}

var defaultConfig = Config{
	Default: DefaultConfig{MaxLines: 100},
	Rules:   []Rule{},
	Exclude: []string{
		"node_modules", "vendor", ".git", "dist", "build",
		"coverage", "__pycache__", "target", ".next", ".nuxt",
		"out", ".cache",
		// Generados: no son código escrito a mano.
		"*.pb.go", "*_gen.go", "*.g.dart", "*-lock.json",
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
		return defaultConfig, fmt.Errorf("config %q: %w", path, err)
	}
	if unknown != "" {
		fmt.Fprintf(os.Stderr, "warning: config %q: clave desconocida %q (ignorada)\n", path, unknown)
	}
	applyConfigDefaults(&cfg)
	return cfg, nil
}


func applyConfigDefaults(cfg *Config) {
	if cfg.Default.MaxLines == 0 {
		cfg.Default.MaxLines = defaultConfig.Default.MaxLines
	}
	if len(cfg.Exclude) == 0 {
		cfg.Exclude = defaultConfig.Exclude
	}
}

func defaultConfigJSON() string {
	return `{
  "default": {
    "maxLines": 100,
    "maxNesting": 0
  },
  "rules": [
    { "pattern": "**/*.spec.*",  "maxLines": 300 },
    { "pattern": "**/*.test.*",  "maxLines": 300 },
    { "pattern": "**/*_test.go", "maxLines": 300 },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": [
    "node_modules",
    "vendor",
    ".git",
    "dist",
    "build",
    "coverage",
    "__pycache__",
    "target",
    ".next",
    "out",
    "*.pb.go",
    "*_gen.go",
    "*.g.dart",
    "*-lock.json"
  ]
}
`
}
