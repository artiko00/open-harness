package main

// mergeConfigChain fusiona los configs de la cadena de fallback campo por campo:
// para cada campo gana el valor del primer config que lo defina (no-cero/no-vacio).
// Los arrays son atomicos: gana entero el del primer config que los defina.
func mergeConfigChain(chain []Config) Config {
	var out Config
	for _, cfg := range chain {
		if out.Language == "" && cfg.Language != "" {
			out.Language = cfg.Language
		}
		if len(out.Exclude) == 0 && len(cfg.Exclude) > 0 {
			out.Exclude = cfg.Exclude
		}
	}
	return out
}
