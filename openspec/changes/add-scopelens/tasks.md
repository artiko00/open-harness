# Tasks: add-scopelens

Cada fase sigue rojo → verde → refactor (ADR-013). Ninguna tarea se marca `[x]` con tests en rojo o
cobertura por debajo del 100% (ADR-011). Los repos de prueba se construyen con un script que arma
árboles git temporales; los tests que requieren `git` real hacen `t.Skip` si no está en `PATH`.

## 0. Andamiaje

- [x] 0.1 Registrar F-019 en `.agent/feature-list.json` con los pasos de verificación
- [x] 0.2 Crear `tools/scopelens/go.mod` (module `github.com/artiko00/open-harness/tools/scopelens`) y sumarlo a `go.work`
- [x] 0.3 [red] Test: `run([]string) int` con `version` devuelve 0 e imprime `scopelens <version>`
- [x] 0.4 [green] `main.go` con `const version = "0.1.0"`, `osExit`, `run([]string) int`, `flag.ContinueOnError`
- [x] 0.5 [red] Test: flag desconocido devuelve 2 y no entra en pánico
- [x] 0.6 [green] Ruteo de subcomandos `check` / `version` / `init` / `--help`

## 1. Adaptador de git

- [x] 1.1 [red] Test: `gitRunner` fake con salida fija devuelve las rutas parseadas
- [x] 1.2 [red] Test: salida vacía de git → 0 archivos, sin error
- [x] 1.3 [red] Test: ruta con quoting octal `"caf\303\251.ts"` se decodifica a `café.ts`
- [x] 1.4 [red] Test: ruta con espacios se parsea como un solo archivo
- [x] 1.5 [green] `gitcmd.go`: `gitRunner` sobre `exec.CommandContext` con `--no-optional-locks` y timeout
- [x] 1.6 [green] `parseNameOnly` con decodificación de quoting octal
- [x] 1.7 [red] Test: `git` ausente en PATH → exit 2 con mensaje accionable
- [x] 1.8 [red] Test: exit code 128 de git → exit 2 incluyendo el stderr de git
- [x] 1.9 [red] Test: timeout → exit 2 mencionando el timeout
- [x] 1.10 [green] Errores tipados de git con `fmt.Errorf("...: %w", err)` y mapeo a exit 2

## 2. Resolución de la base

- [x] 2.1 [red] Test: precedencia `--base` > config > `origin/HEAD` > `main` > `master`
- [x] 2.2 [red] Test: repo sin `main` ni `master` ni `origin/HEAD` → exit 2 sugiriendo `--base`
- [x] 2.3 [red] Test: clon shallow → exit 2 mencionando `fetch-depth: 0`
- [x] 2.4 [red] Test: directorio sin `.git` → exit 2
- [x] 2.5 [green] `base.go`: detección de repo, de shallow (`rev-parse --is-shallow-repository`) y cadena de resolución
- [x] 2.6 [green] `merge-base <base> HEAD` con su hash abreviado para el encabezado
- [x] 2.7 [red] Test: `HEAD` sobre la propia base → cuenta sólo staged, exit 0
- [x] 2.8 [red] Test: repo recién inicializado sin commits → cuenta sólo staged, exit 0
- [x] 2.9 [green] Caso base-igual-a-HEAD y caso sin commits sin emitir error

## 3. Conteo acumulado

- [x] 3.1 [red] Test: unión de `<merge-base>...HEAD` con `--cached`, archivo repetido cuenta 1
- [x] 3.2 [red] Test: 5 commits de 4 archivos → conteo 20
- [x] 3.3 [red] Test: rename con `-M` → cuenta 1
- [x] 3.4 [red] Test: borrado con `--diff-filter=ACMRD` → cuenta y aparece en el reporte
- [x] 3.5 [red] Test: `--staged-only` ignora los commits previos de la rama
- [x] 3.6 [green] `scanner.go`: unión de conjuntos, orden lexicográfico explícito (sin iterar maps)
- [x] 3.7 [red] Test de determinismo: 20 corridas con salida byte-idéntica

## 4. Clasificación multi-ecosistema

- [x] 4.1 [red] Test: `*.test.ts`, `*.spec.js`, `__tests__/**` clasifican como `test`
- [x] 4.2 [red] Test: `test_*.py`, `*_test.py`, `tests/**` clasifican como `test`
- [x] 4.3 [red] Test: `*_test.go` clasifica como `test`
- [x] 4.4 [red] Test: `pnpm-lock.yaml`, `poetry.lock`, `go.sum` clasifican como `excluded` con motivo `lockfile`
- [x] 4.5 [red] Test: `**/*.pb.go` y `**/zz_generated*.go` clasifican como `excluded` con motivo `generated`
- [x] 4.6 [green] `classify.go` delegando todo el matching en `_shared/pathmatch`
- [x] 4.7 [red] Test: `--exclude-tests` descuenta la categoría `test` del conteo contable
- [x] 4.8 [green] Aplicación de `excludeTests` al conteo contable
- [x] 4.9 [refactor] Verificar con `dupelens check --dir tools` que no se duplicó lógica de globs

## 5. Configuración multi-ecosistema

- [x] 5.1 [red] Test: `scopelens.json` con `maxFiles`, `base`, `excludeTests`, `exclude`
- [x] 5.2 [red] Test: `pyproject.toml` → `[tool.scopelens]` (vía `tomlmin`)
- [x] 5.3 [red] Test: `package.json` → clave `"scopelens"`
- [x] 5.4 [red] Test: `composer.json` → `extra.open-harness.scopelens`
- [x] 5.5 [red] Test: precedencia completa flags > json > toml > package > composer > defaults
- [x] 5.6 [green] `config.go` + `config_chain.go` siguiendo el patrón de linelens (ADR-018)
- [x] 5.7 [red] Test: JSON malformado → exit 2 nombrando el archivo
- [x] 5.8 [red] Test: `maxFiles` negativo o no numérico → exit 2 nombrando la clave
- [x] 5.9 [green] Validación de config que falla ruidosa en vez de caer a defaults
- [x] 5.10 [red] Test: `exclude` del usuario reemplaza los defaults, no se suma
- [x] 5.11 [green] Defaults de exclusión de los tres ecosistemas + `defaultConfigJSON()` para `init`

## 6. Reporte y umbral

- [x] 6.1 [red] Test: conteo == `maxFiles` → exit 0; conteo == `maxFiles + 1` con `--fail` → exit 1
- [x] 6.2 [red] Test: sin `--fail`, un exceso reporta `FAIL:` pero devuelve exit 0
- [x] 6.3 [green] `reporter.go`: encabezado, desglose por categoría, motivo de exclusión, `SUMMARY`
- [x] 6.4 [red] Test: `--no-color` no emite ninguna secuencia ANSI
- [x] 6.5 [green] Colores con el mismo esquema que linelens
- [x] 6.6 [red] Test: `init` crea `scopelens.json` con los defaults
- [x] 6.7 [green] Subcomando `init`

## 7. Integración con el monorepo

- [x] 7.1 Escribir `docs/adr-020-scopelens-dependencia-git.md` (git como binario externo; alternativa `.git/` descartada)
- [x] 7.2 Agregar `scopelens` a `lefthook.yml` en `pre-commit` y en `pre-push` (go test)
- [x] 7.3 Agregar `scopelens` a `open-harness.json` y crear `scopelens.json` en la raíz
- [x] 7.4 Test de integración: repo temporal que excede el presupuesto aborta el `git commit`
- [x] 7.5 Verificar que el propio repo pasa `scopelens check --fail` (ADR-009)

## 8. Distribución

- [x] 8.1 Extender `scripts/build-npm.sh` para `scopelens` (4 plataformas + meta)
- [x] 8.2 Crear `npm/scopelens/` con los 5 `package.json`
- [skip] 8.3 Agregar `scopelens` al empaquetado PyPI (F-012)
- [skip] 8.4 Agregar `scopelens` al empaquetado Packagist (F-013)
- [skip] 8.5 Agregar `scopelens` al meta-paquete `@open_harness/open-harness`

## 9. Documentación y quality gates

- [x] 9.1 Sección de `scopelens` en `README.md` con la comparación contra pr-size-labeler y Danger JS
- [x] 9.2 Documentar en `AGENTS.md` el nuevo exit code 2 y su semántica
- [x] 9.3 `cd tools/scopelens && go test ./... -cover` → 100% de statements
- [x] 9.4 `linelens check --fail --no-color` sobre el repo completo
- [x] 9.5 `dupelens check --fail --no-color` sobre el repo completo
- [x] 9.6 `secretlens check --fail --no-color` sobre el repo completo
- [x] 9.7 `testlens check --lang go --dir tools/ --fail`
- [x] 9.8 `scopelens check --fail --no-color` sobre el repo completo
- [ ] 9.9 Actualizar `.agent/claude-progress.txt`
