# CLAUDE.md

Contexto para Claude Code sobre `open-harness`.

El workflow del proyecto —filosofía, TDD, exit codes, quality gates, boundaries—
vive en un solo lugar y se importa acá para que no haya dos verdades:

@AGENTS.md

Lo que sigue es lo específico de trabajar este repo con Claude Code.

---

## Antes de tocar `tools/_shared/`

Los cinco binarios comparten cuatro módulos: `tomlmin` (parser TOML),
`configload` (cadena de configuración), `pathmatch` (globs) y `langsyntax`
(sintaxis por familia de lenguaje). Un cambio ahí **se publica en los cinco
tools a la vez**, así que:

- Corré los cinco suites, no solo el módulo tocado: `for m in tools/*/ tools/_shared/*/; do (cd $m && go test ./...); done`
- Un bug en un módulo compartido es un release de los cinco. Presupuestá el
  trabajo completo (bump ×5 + npm ×26 + PyPI ×6) antes de empezar.
- `go build ./...` desde la raíz **no funciona**: es un Go workspace con un
  módulo por directorio. Hay que entrar a cada uno.

## Ramas

Se trabaja directo en `develop` y de ahí a `main`. **No crear ramas feature.**
Se publica desde `main`.

## Comandos con salida filtrada

El entorno reescribe algunos comandos vía proxy y recorta la salida. Si
necesitás la salida completa de un `go test -v`, escribí el resultado a un
archivo y leelo, o usá `rtk proxy <cmd>`.

## Verificar un fix de config end-to-end

Los tests unitarios no alcanzan para la cadena de configuración: armá un repo
temporal con el manifiesto real (`pyproject.toml`, `package.json` o
`composer.json`) y corré el binario compilado contra él. Un valor que **cambie
el resultado** (por ejemplo `maxLines` bajo que produzca una violación) prueba
que la config se aplicó; un "OK" puede significar que se cayó a los defaults en
silencio. Comparar contra el binario publicado en `npm/@open_harness/<tool>-linux-x64/bin/`
da el antes/después sin recompilar nada.

## OpenSpec

Se usa el CLI `openspec` (no las skills `sdd-*`):

```bash
openspec new change <nombre>      # crea openspec/changes/<nombre>/
openspec validate <nombre> --strict
openspec archive <nombre> -y      # promueve los delta specs a openspec/specs/
```

Un change lleva `proposal.md`, `specs/<capability>/spec.md` con los deltas
(`## ADDED Requirements`, cada Requirement con sus `#### Scenario`) y `tasks.md`
con checkboxes por fase.
