package main

import "fmt"

const tutorialText = `secretlens — guia de configuracion

Coloca un secretlens.json (o scopelens.json) en la raiz del repo. Todas las
claves son opcionales; se combinan con los valores por defecto.

CLAVES DE CONFIG:

  patterns  ([]{name,pattern,severity,entropyGate}) — reglas custom de deteccion.
    Por defecto: [] (solo se usan los patterns integrados). Desde 0.3.0 los
    patterns son ADITIVOS: tus reglas se SUMAN a las integradas, no las
    reemplazan (ver disableDefaultPatterns).
    Ejemplo: "patterns": [{"name":"corp-token","pattern":"CORP-[A-Z0-9]{20}",
      "severity":"high","entropyGate":true}]

  allowlist  ([]string) — subcadenas que marcan un valor como falso positivo.
    Por defecto: ["example","placeholder","your_key_here","changeme","xxxx","****"].
    Ejemplo: "allowlist": ["example", "dummy", "TEST_ONLY"]

  exclude  ([]string) — globs/dirs que no se escanean.
    Por defecto: ["node_modules","vendor",".git","dist","build","coverage",
      "__pycache__","target",".next","out",".cache","*.lock","go.sum"].
    Ejemplo: "exclude": ["node_modules", "*.lock", "testdata"]

  minEntropy  (float64) — umbral de entropia de Shannon (bits/caracter) que debe
    superar el valor capturado por reglas genericas KEY=VALUE para reportarse.
    Por defecto: 3.0. Ejemplo: "minEntropy": 3.5

  disableDefaultPatterns  (bool) — si es true, NO se anteponen los patterns
    integrados y solo se usan los tuyos. Por defecto: false.
    Ejemplo: "disableDefaultPatterns": true

FLAGS (subcomando check):

  --config    ruta al archivo de config     (default: secretlens.json)
  --dir       directorio a escanear         (default: .)
  --format    formato de salida             (console | json, default: console)
  --fail      exit 1 si hay hallazgos       (usar en git hooks)
  --no-color  desactiva el color ANSI

Comandos: check, init, version, help, --tutorial.

CAMBIOS 0.3.0 RELEVANTES:
  - patterns es ADITIVO: tus reglas se suman a las integradas.
  - disableDefaultPatterns permite optar por SOLO tus reglas.
`

// printTutorial imprime la guia estatica de configuracion a stdout. Si noColor
// es false, colorea el titulo reutilizando el esquema del reporter.
func printTutorial(noColor bool) {
	if noColor {
		fmt.Print(tutorialText)
		return
	}
	fmt.Printf("%s%ssecretlens — guia de configuracion%s\n", colorBold, colorCyan, colorReset)
	fmt.Print(tutorialText[len("secretlens — guia de configuracion\n"):])
}
