package main

import (
	"fmt"
	"os"
	"strings"
)

// Marcadores «/» delimitan encabezados en negrita; se sustituyen por ANSI o
// por vacío según --no-color, de modo que el texto base sea único.
const tutorialText = `«linelens — guía de configuración (v0.3.0)»

Configura linelens con un archivo linelens.json (o scopelens.json) en la raíz
del proyecto. Genera uno base con: linelens init

«CLAVES DE CONFIG (linelens.json)»

  «default»     objeto con los límites por defecto aplicados a todo archivo.
  «maxLines»    (dentro de default) máximo de líneas por archivo. Default: 100.
                 Ejemplo: "default": { "maxLines": 100 }
  «maxNesting»  (dentro de default) máxima anidación permitida. Default: 0
                 (0 = desactivado). Ejemplo: "default": { "maxNesting": 4 }
  «rules»       lista de reglas por patrón glob; cada una acepta:
                 - «pattern»  glob del archivo, ej. "**/*_test.go".
                 - «maxLines» límite propio para ese patrón, ej. 300.
                 - «skip»     true para omitir por completo el patrón.
                 Ejemplo: "rules": [ { "pattern": "**/*_test.go", "maxLines": 300 } ]
  «exclude»     lista de rutas/globs a ignorar. Si se omite, usa el default
                 (node_modules, vendor, .git, dist, *.pb.go, *_gen.go, ...).
                 Ejemplo: "exclude": [ "node_modules", "vendor", "*.g.dart" ]

«FLAGS (linelens check)»

  «--config»    ruta al archivo de config      (default: linelens.json)
  «--max»       fuerza el máximo de líneas      (default: config o 100)
  «--dir»       directorio a escanear           (default: .)
  «--fail»      exit 1 si hay violaciones       (útil en git hooks)
  «--no-color»  desactiva el color ANSI
  «--format»    formato de salida: console | json (default: console)

Otros comandos: linelens init, linelens version, linelens --tutorial.

«CAMBIOS DE 0.3.0 RELEVANTES»

  - Config estricta: las claves desconocidas se ignoran con un warning en
    stderr en vez de fallar silenciosamente.
  - Los archivos omitidos (skip) se reportan y, con --fail, un skip que marca
    gate ya no "pasa en verde": deja de aprobar por omisión.
`

func printTutorial(noColor bool) {
	on, off := colorBold, colorReset
	if noColor {
		on, off = "", ""
	}
	text := strings.ReplaceAll(tutorialText, "«", on)
	text = strings.ReplaceAll(text, "»", off)
	fmt.Fprint(os.Stdout, text)
}
