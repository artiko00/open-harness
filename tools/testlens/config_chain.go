package main

func loadConfigFallbackChain(dir string) (Config, error) {
	if cfg, found, err := loadConfigFromPyprojectToml(dir); err != nil {
		return defaultConfig, err
	} else if found {
		return cfg, nil
	}
	if cfg, found, err := loadConfigFromPackageJSON(dir); err != nil {
		return defaultConfig, err
	} else if found {
		return cfg, nil
	}
	if cfg, found, err := loadConfigFromComposerJSON(dir); err != nil {
		return defaultConfig, err
	} else if found {
		return cfg, nil
	}
	return defaultConfig, nil
}
