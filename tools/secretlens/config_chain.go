package main

import "github.com/artiko00/open-harness/tools/_shared/configload"

// loadConfigFallbackChain acumula los archivos presentes en orden de prioridad
// (pyproject.toml → package.json → composer.json) y hace merge por campo:
// para cada campo gana el valor del PRIMER archivo de la cadena que lo defina.
// Al agotar la cadena, applyConfigDefaults rellena lo que quede.
func loadConfigFallbackChain(dir string) (Config, error) {
	loaders := []func(string) (Config, bool, error){
		func(d string) (Config, bool, error) { return configload.Pyproject[Config](d, "tool.secretlens") },
		func(d string) (Config, bool, error) { return configload.PackageJSON[Config](d, "secretlens") },
		func(d string) (Config, bool, error) { return configload.Composer[Config](d, "secretlens") },
	}

	var chain []Config
	for _, load := range loaders {
		cfg, found, err := load(dir)
		if err != nil {
			return defaultConfig, err
		}
		if found {
			chain = append(chain, cfg)
		}
	}

	merged := mergeConfigChain(chain)
	applyConfigDefaults(&merged)
	return merged, nil
}

// mergeConfigChain combina la cadena por campo. Los arrays son atómicos:
// gana entero el del primer archivo de la cadena que los defina (no-vacío).
func mergeConfigChain(chain []Config) Config {
	var out Config
	for _, cfg := range chain {
		if len(out.Patterns) == 0 {
			out.Patterns = cfg.Patterns
		}
		if len(out.Allowlist) == 0 {
			out.Allowlist = cfg.Allowlist
		}
		if len(out.Exclude) == 0 {
			out.Exclude = cfg.Exclude
		}
	}
	return out
}
