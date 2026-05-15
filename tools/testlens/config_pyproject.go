package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/artiko00/open-harness/tools/_shared/tomlmin"
)

const pyprojectSection = "tool.testlens"

func loadConfigFromPyprojectToml(dir string) (Config, bool, error) {
	data, err := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig, false, nil
		}
		return defaultConfig, false, err
	}

	section, found, err := tomlmin.ExtractAsJSON(data, pyprojectSection)
	if err != nil {
		return defaultConfig, false, fmt.Errorf("pyproject.toml: %w", err)
	}
	if !found {
		return defaultConfig, false, nil
	}

	var cfg Config
	if err := json.Unmarshal(section, &cfg); err != nil {
		return defaultConfig, false, err
	}
	applyConfigDefaults(&cfg)
	return cfg, true, nil
}
