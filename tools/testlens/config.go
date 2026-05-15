package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Language string   `json:"language"`
	Exclude  []string `json:"exclude"`
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
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig, err
	}
	applyConfigDefaults(&cfg)
	return cfg, nil
}

func applyConfigDefaults(cfg *Config) {
	if cfg.Language == "" {
		cfg.Language = defaultConfig.Language
	}
	if len(cfg.Exclude) == 0 {
		cfg.Exclude = defaultConfig.Exclude
	}
}
