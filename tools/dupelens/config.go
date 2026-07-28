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
	MinTokens int `json:"minTokens"`
	MinLines  int `json:"minLines"`
	// WindowSize es el tamaño de ventana de detección (independiente de
	// MinTokens, el umbral de reporte). 0 = usar defaultWindow. Bajar MinTokens
	// para reducir ruido no obliga a tocar la ventana.
	WindowSize int `json:"windowSize"`
}

type Rule struct {
	Pattern   string `json:"pattern"`
	MinTokens int    `json:"minTokens"`
	Skip      bool   `json:"skip"`
}

var defaultConfig = Config{
	Default: DefaultConfig{MinTokens: 50, MinLines: 5},
	Rules:   []Rule{},
	Exclude: []string{
		"node_modules", "vendor", ".git", "dist", "build",
		"coverage", "__pycache__", "target", ".next", ".nuxt",
		"out", ".cache",
	},
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return loadConfigFallbackChain(filepath.Dir(path))
		}
		return defaultConfig, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Config
	unknown, err := configload.Strict(data, &cfg)
	if err != nil {
		return defaultConfig, fmt.Errorf("parsing config %q: %w (run 'dupelens init' to regenerate)", path, err)
	}
	if unknown != "" {
		fmt.Fprintf(os.Stderr, "warning: config %q: clave desconocida %q (ignorada)\n", path, unknown)
	}
	applyConfigDefaults(&cfg)
	return cfg, nil
}

func applyConfigDefaults(cfg *Config) {
	if cfg.Default.MinTokens == 0 {
		cfg.Default.MinTokens = defaultConfig.Default.MinTokens
	}
	if cfg.Default.MinLines == 0 {
		cfg.Default.MinLines = defaultConfig.Default.MinLines
	}
	if len(cfg.Exclude) == 0 {
		cfg.Exclude = defaultConfig.Exclude
	}
}

func defaultConfigJSON() string {
	return `{
  "default": {
    "minTokens": 50,
    "minLines": 5
  },
  "rules": [
    { "pattern": "**/*.spec.*",  "skip": true },
    { "pattern": "**/*.test.*",  "skip": true },
    { "pattern": "**/*_test.go", "skip": true },
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
    "out"
  ]
}
`
}
