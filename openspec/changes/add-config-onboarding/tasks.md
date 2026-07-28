# Tasks: add-config-onboarding

Cada fase rojo → verde → refactor (ADR-013). Ninguna tarea `[x]` con tests en rojo o cobertura bajo
100% (ADR-011). Los 5 tools son Go (100% cov); el launcher del meta es JS/Python (smoke test).

## 0. Andamiaje

- [x] 0.1 Registrar F-020 en `.agent/feature-list.json`

## 1. `--tutorial` en los 5 tools

- [x] 1.1 [red] Test (×5): `<tool> --tutorial` imprime a stdout y devuelve exit 0
- [x] 1.2 [red] Test (×5): el tutorial menciona cada clave JSON del struct `Config` del tool
- [x] 1.3 [red] Test (×5): `--tutorial --no-color` no emite ninguna secuencia ANSI
- [x] 1.4 [green] `tutorial.go` (×5): contenido estático + render, respetando `--no-color`
- [x] 1.5 [green] `case "--tutorial", "tutorial":` en `run()` de los 5 `main.go`
- [x] 1.6 [green] La guía lista los flags de cada tool (`--config`, `--fail`, `--format`, propios)
- [x] 1.7 [refactor] Verificar que ningún `tutorial.go` supera 100 líneas; separar el texto si hace falta

## 2. `open-harness init` (meta)

- [x] 2.1 [red] Smoke test: `open-harness init` en un temp dir crea los 5 `<tool>.json`
- [x] 2.2 [green] `bin/open-harness.js`: comando `init` que spawnea el `init` de cada tool
- [x] 2.3 [green] Entry point equivalente en `pypi/open_harness` (Python) para `open-harness init`
- [x] 2.4 [green] `package.json` del meta: agregar el bin `open-harness`
- [x] 2.5 [green] No sobrescribir; reportar creados vs ya existentes

## 3. Changelogs

- [x] 3.1 `CHANGELOG.md` en linelens/dupelens/secretlens/testlens: entrada 0.3.0 (fixes F-018 + 3 BREAKING)
- [x] 3.2 `CHANGELOG.md` en scopelens: entrada 0.1.0 (release inicial)
- [x] 3.3 `CHANGELOG.md` del meta open-harness: entrada 0.3.0
- [x] 3.4 Formato Keep a Changelog; el de secretlens marca el breaking de patterns aditivos con opt-out

## 4. Guía de configuración

- [x] 4.1 `docs/CONFIGURATION.md`: claves/defaults de los 5 tools, precedencia de la cadena, ejemplos por lenguaje
- [x] 4.2 Enlazar `docs/CONFIGURATION.md` desde el README
- [x] 4.3 Mencionar `--tutorial` y `open-harness init` en el README

## 5. Quality gates

- [x] 5.1 `cd tools/<tool> && go test ./... -cover` → 100% en los 5 tools
- [x] 5.2 `linelens check --fail` sobre el repo (ningún `tutorial.go` > 100 líneas)
- [x] 5.3 `dupelens check --fail` sobre el repo (los 5 tutoriales no son código duplicado)
- [x] 5.4 `secretlens check --fail` sobre el repo
- [x] 5.5 `testlens check --lang go --dir tools/ --fail`
- [x] 5.6 `scopelens check --fail` sobre el repo
- [x] 5.7 Smoke test del meta `open-harness init` verde
- [x] 5.8 Actualizar `.agent/claude-progress.txt`
