# Tasks: add-scopelens

Cada fase sigue rojo → verde → refactor (ADR-013). Ninguna tarea se marca `[x]` con tests en rojo o
cobertura por debajo del 100% (ADR-011). Los repos de prueba se construyen con un script que arma
árboles git temporales; los tests que requieren `git` real hacen `t.Skip` si no está en `PATH`.

## 0. Andamiaje

- [ ] 0.1 Registrar F-019 en `.agent/feature-list.json` con los pasos de verificación
- [ ] 0.2 Crear `tools/scopelens/go.mod` (module `github.com/artiko00/open-harness/tools/scopelens`) y sumarlo a `go.work`
- [ ] 0.3 [red] Test: `run([]string) int` con `version` devuelve 0 e imprime `scopelens <version>`
- [ ] 0.4 [green] `main.go` con `const version = "0.1.0"`, `osExit`, `run([]string) int`, `flag.ContinueOnError`
- [ ] 0.5 [red] Test: flag desconocido devuelve 2 y no entra en pánico
- [ ] 0.6 [green] Ruteo de subcomandos `check` / `version` / `init` / `--help`

## 1. Adaptador de git

- [ ] 1.1 [red] Test: `gitRunner` fake con salida fija devuelve las rutas parseadas
- [ ] 1.2 [red] Test: salida vacía de git → 0 archivos, sin error
- [ ] 1.3 [red] Test: ruta con quoting octal `"caf\303\251.ts"` se decodifica a `café.ts`
- [ ] 1.4 [red] Test: ruta con espacios se parsea como un solo archivo
- [ ] 1.5 [green] `gitcmd.go`: `gitRunner` sobre `exec.CommandContext` con `--no-optional-locks` y timeout
- [ ] 1.6 [green] `parseNameOnly` con decodificación de quoting octal
- [ ] 1.7 [red] Test: `git` ausente en PATH → exit 2 con mensaje accionable
- [ ] 1.8 [red] Test: exit code 128 de git → exit 2 incluyendo el stderr de git
- [ ] 1.9 [red] Test: timeout → exit 2 mencionando el timeout
- [ ] 1.10 [green] Errores tipados de git con `fmt.Errorf("...: %w", err)` y mapeo a exit 2

## 2. Resolución de la base

- [ ] 2.1 [red] Test: precedencia `--base` > config > `origin/HEAD` > `main` > `master`
- [ ] 2.2 [red] Test: repo sin `main` ni `master` ni `origin/HEAD` → exit 2 sugiriendo `--base`
- [ ] 2.3 [red] Test: clon shallow → exit 2 mencionando `fetch-depth: 0`
- [ ] 2.4 [red] Test: directorio sin `.git` → exit 2
- [ ] 2.5 [green] `base.go`: detección de repo, de shallow (`rev-parse --is-shallow-repository`) y cadena de resolución
- [ ] 2.6 [green] `merge-base <base> HEAD` con su hash abreviado para el encabezado
- [ ] 2.7 [red] Test: `HEAD` sobre la propia base → cuenta sólo staged, exit 0
- [ ] 2.8 [red] Test: repo recién inicializado sin commits → cuenta sólo staged, exit 0
- [ ] 2.9 [green] Caso base-igual-a-HEAD y caso sin commits sin emitir error

## 3. Conteo acumulado

- [ ] 3.1 [red] Test: unión de `<merge-base>...HEAD` con `--cached`, archivo repetido cuenta 1
- [ ] 3.2 [red] Test: 5 commits de 4 archivos → conteo 20
- [ ] 3.3 [red] Test: rename con `-M` → cuenta 1
- [ ] 3.4 [red] Test: borrado con `--diff-filter=ACMRD` → cuenta y aparece en el reporte
- [ ] 3.5 [red] Test: `--staged-only` ignora los commits previos de la rama
- [ ] 3.6 [green] `scanner.go`: unión de conjuntos, orden lexicográfico explícito (sin iterar maps)
- [ ] 3.7 [red] Test de determinismo: 20 corridas con salida byte-idéntica

## 4. Clasificación multi-ecosistema

- [ ] 4.1 [red] Test: `*.test.ts`, `*.spec.js`, `__tests__/**` clasifican como `test`
- [ ] 4.2 [red] Test: `test_*.py`, `*_test.py`, `tests/**` clasifican como `test`
- [ ] 4.3 [red] Test: `*_test.go` clasifica como `test`
- [ ] 4.4 [red] Test: `pnpm-lock.yaml`, `poetry.lock`, `go.sum` clasifican como `excluded` con motivo `lockfile`
- [ ] 4.5 [red] Test: `**/*.pb.go` y `**/zz_generated*.go` clasifican como `excluded` con motivo `generated`
- [ ] 4.6 [green] `classify.go` delegando todo el matching en `_shared/pathmatch`
- [ ] 4.7 [red] Test: `--exclude-tests` descuenta la categoría `test` del conteo contable
- [ ] 4.8 [green] Aplicación de `excludeTests` al conteo contable
- [ ] 4.9 [refactor] Verificar con `dupelens check --dir tools` que no se duplicó lógica de globs

## 5. Configuración multi-ecosistema

- [ ] 5.1 [red] Test: `scopelens.json` con `maxFiles`, `base`, `excludeTests`, `exclude`
- [ ] 5.2 [red] Test: `pyproject.toml` → `[tool.scopelens]` (vía `tomlmin`)
- [ ] 5.3 [red] Test: `package.json` → clave `"scopelens"`
- [ ] 5.4 [red] Test: `composer.json` → `extra.open-harness.scopelens`
- [ ] 5.5 [red] Test: precedencia completa flags > json > toml > package > composer > defaults
- [ ] 5.6 [green] `config.go` + `config_chain.go` siguiendo el patrón de linelens (ADR-018)
- [ ] 5.7 [red] Test: JSON malformado → exit 2 nombrando el archivo
- [ ] 5.8 [red] Test: `maxFiles` negativo o no numérico → exit 2 nombrando la clave
- [ ] 5.9 [green] Validación de config que falla ruidosa en vez de caer a defaults
- [ ] 5.10 [red] Test: `exclude` del usuario reemplaza los defaults, no se suma
- [ ] 5.11 [green] Defaults de exclusión de los tres ecosistemas + `defaultConfigJSON()` para `init`

## 6. Reporte y umbral

- [ ] 6.1 [red] Test: conteo == `maxFiles` → exit 0; conteo == `maxFiles + 1` con `--fail` → exit 1
- [ ] 6.2 [red] Test: sin `--fail`, un exceso reporta `FAIL:` pero devuelve exit 0
- [ ] 6.3 [green] `reporter.go`: encabezado, desglose por categoría, motivo de exclusión, `SUMMARY`
- [ ] 6.4 [red] Test: `--no-color` no emite ninguna secuencia ANSI
- [ ] 6.5 [green] Colores con el mismo esquema que linelens
- [ ] 6.6 [red] Test: `init` crea `scopelens.json` con los defaults
- [ ] 6.7 [green] Subcomando `init`

## 7. Integración con el monorepo

- [ ] 7.1 Escribir `docs/adr-020-scopelens-dependencia-git.md` (git como binario externo; alternativa `.git/` descartada)
- [ ] 7.2 Agregar `scopelens` a `lefthook.yml` en `pre-commit` y en `pre-push` (go test)
- [ ] 7.3 Agregar `scopelens` a `open-harness.json` y crear `scopelens.json` en la raíz
- [ ] 7.4 Test de integración: repo temporal que excede el presupuesto aborta el `git commit`
- [ ] 7.5 Verificar que el propio repo pasa `scopelens check --fail` (ADR-009)

## 8. Distribución

- [ ] 8.1 Extender `scripts/build-npm.sh` para `scopelens` (4 plataformas + meta)
- [ ] 8.2 Crear `npm/scopelens/` con los 5 `package.json`
- [ ] 8.3 Agregar `scopelens` al empaquetado PyPI (F-012)
- [ ] 8.4 Agregar `scopelens` al empaquetado Packagist (F-013)
- [ ] 8.5 Agregar `scopelens` al meta-paquete `@open_harness/open-harness`

## 9. Documentación y quality gates

- [ ] 9.1 Sección de `scopelens` en `README.md` con la comparación contra pr-size-labeler y Danger JS
- [ ] 9.2 Documentar en `AGENTS.md` el nuevo exit code 2 y su semántica
- [ ] 9.3 `cd tools/scopelens && go test ./... -cover` → 100% de statements
- [ ] 9.4 `linelens check --fail --no-color` sobre el repo completo
- [ ] 9.5 `dupelens check --fail --no-color` sobre el repo completo
- [ ] 9.6 `secretlens check --fail --no-color` sobre el repo completo
- [ ] 9.7 `testlens check --lang go --dir tools/ --fail`
- [ ] 9.8 `scopelens check --fail --no-color` sobre el repo completo
- [ ] 9.9 Actualizar `.agent/claude-progress.txt`
