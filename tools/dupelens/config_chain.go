package main

import "github.com/artiko00/open-harness/tools/_shared/configload"

// loadConfigFallbackChain acumula los archivos presentes en orden de
// prioridad (pyproject.toml, package.json, composer.json) y hace merge por
// campo: gana el valor del PRIMER archivo de la cadena que lo defina. Los
// arrays son atómicos. Al agotar la cadena, applyConfigDefaults rellena el
// resto con los defaults compilados.
func loadConfigFallbackChain(dir string) (Config, error) {
	loaders := []func(string) (Config, bool, error){
		func(d string) (Config, bool, error) { return configload.Pyproject[Config](d, "tool.dupelens") },
		func(d string) (Config, bool, error) { return configload.PackageJSON[Config](d, "dupelens") },
		func(d string) (Config, bool, error) { return configload.Composer[Config](d, "dupelens") },
	}
	var acc Config
	for _, load := range loaders {
		cfg, found, err := load(dir)
		if err != nil {
			return defaultConfig, err
		}
		if found {
			mergeConfig(&acc, cfg)
		}
	}
	applyConfigDefaults(&acc)
	return acc, nil
}

// mergeConfig completa los campos aún sin definir en dst (cero/vacío) con los
// de src, sin sobrescribir lo ya presente. Los arrays se toman enteros.
func mergeConfig(dst *Config, src Config) {
	if dst.Default.MinTokens == 0 {
		dst.Default.MinTokens = src.Default.MinTokens
	}
	if dst.Default.MinLines == 0 {
		dst.Default.MinLines = src.Default.MinLines
	}
	if len(dst.Rules) == 0 {
		dst.Rules = src.Rules
	}
	if len(dst.Exclude) == 0 {
		dst.Exclude = src.Exclude
	}
}
