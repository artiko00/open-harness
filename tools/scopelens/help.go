package main

import "fmt"

func printUsage() {
	fmt.Print(`scopelens — presupuesto de archivos por PR (diff acumulado rama vs base)

USO:
  scopelens <comando> [opciones]

COMANDOS:
  check     mide el alcance del cambio y compara contra el presupuesto
  init      crea un scopelens.json con los defaults
  version   muestra la versión

OPCIONES DE CHECK:
  --max-files <int>   presupuesto de archivos       (default: config o 15)
  --base <ref>        rama base a comparar           (default: origin/HEAD > main > master)
  --dir <path>        directorio del repositorio     (default: .)
  --staged-only       contar sólo el índice (staged)
  --exclude-tests     descontar los archivos de test del presupuesto
  --fail              exit 1 si excede el presupuesto (usar en git hooks)
  --no-color          salida sin secuencias ANSI

EXIT CODES:
  0  medido dentro del presupuesto (o fuera, sin --fail)
  1  medido y excede el presupuesto (con --fail)
  2  no se pudo medir (git ausente, no es repo, shallow, base no resoluble, config inválida)

EJEMPLOS:
  scopelens check
  scopelens check --fail --no-color
  scopelens check --base develop --max-files 20
`)
}
