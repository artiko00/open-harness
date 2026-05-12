package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const packageJSONKey = "dupelens"

func loadConfigFromPackageJSON(dir string) (Config, bool, error) {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig, false, nil
		}
		return defaultConfig, false, fmt.Errorf("reading %q: %w", path, err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return defaultConfig, false, fmt.Errorf("parsing %q: %w", path, err)
	}

	section, ok := raw[packageJSONKey]
	if !ok {
		return defaultConfig, false, nil
	}

	var cfg Config
	if err := json.Unmarshal(section, &cfg); err != nil {
		return defaultConfig, false, fmt.Errorf("parsing %q.%s: %w", path, packageJSONKey, err)
	}

	applyConfigDefaults(&cfg)
	return cfg, true, nil
}
