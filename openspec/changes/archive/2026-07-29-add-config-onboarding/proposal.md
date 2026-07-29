# Add config onboarding: --tutorial, suite init, changelogs y guía de configuración

Feature ID: **F-020** (`.agent/feature-list.json`)
Affected tools: **los 5 lenses** (linelens, dupelens, secretlens, testlens, scopelens) + **meta** (open-harness)
Risk: **low** (features aditivos: un flag nuevo, un comando en el launcher, y documentación)

## Why

open-harness se publica a npm y PyPI, pero un usuario que instala un lens hoy no tiene forma de
descubrir **qué** puede configurar sin leer el código o el README completo. Y quien adopta la suite
tiene que correr cinco `init` por separado para arrancar. Faltan tres piezas de onboarding antes de
publicar:

1. **No hay changelog.** Un usuario que salta de 0.2.x a 0.3.0 (un release con 3 cambios de
   comportamiento) no tiene un documento que le diga qué cambió por tool. `docs/UPGRADING.md` existe
   pero es una guía única; falta el `CHANGELOG.md` por paquete que npm/PyPI muestran en la página.
2. **Arrancar la suite es tedioso.** Cada tool tiene `init` (crea su `<tool>.json`), pero no hay un
   comando único que deje el proyecto configurado de una.
3. **La config es opaca en la terminal.** Para saber qué hace `minEntropy` o `maxNesting` hay que
   salir a buscar el README. Un `--tutorial` que explique cada clave donde el usuario ya está —en la
   CLI— cierra ese hueco.

## What Changes

- **`--tutorial` en los 5 tools.** `<tool> --tutorial` imprime una guía estática: cada clave de
  config con su default y un ejemplo, los flags disponibles, y (donde aplica) los cambios de
  comportamiento de 0.3.0. Salida a stdout, sin ANSI con `--no-color`, exit 0.
- **`open-harness init` (comando nuevo del meta).** Crea los cinco archivos de config en la raíz del
  proyecto de una sola corrida, delegando en el `init` de cada tool. Cada tool conserva su `init`
  individual.
- **`CHANGELOG.md` por paquete** (5 tools + meta), formato *Keep a Changelog*: `0.3.0` para los
  cuatro originales + meta (los fixes de F-018 y los 3 BREAKING), `0.1.0` para scopelens.
- **`docs/CONFIGURATION.md`**: guía central de configuración de los 5 tools (claves, defaults,
  precedencia de la cadena multi-ecosistema, ejemplos por lenguaje). Enlazada desde el README.

## Capabilities

### New Capabilities

| Capability | Cubre |
|---|---|
| `tutorial-command` | `<tool> --tutorial` imprime la guía de configuración de cada tool |
| `suite-init` | `open-harness init` crea los 5 archivos de config en la raíz |
| `release-docs` | `CHANGELOG.md` por paquete + `docs/CONFIGURATION.md` |

### Modified Capabilities

Ninguna (todo aditivo; no cambia el comportamiento de `check` ni de la carga de config existente).

## Scope

### In Scope

- El flag `--tutorial` en los 5 tools, con tests y 100% de cobertura.
- El comando `open-harness init` en el launcher del meta (npm y PyPI).
- Los 6 `CHANGELOG.md` y `docs/CONFIGURATION.md`.
- Actualizar los README para mencionar `--tutorial`, `open-harness init` y enlazar la guía de config.

### Out of Scope

- Un `--tutorial` interactivo (pregunta/respuesta): se elige la guía impresa estática por ser
  idiomática, testeable al 100% y coherente con el patrón stdlib de los tools.
- Cambiar el formato o las claves de config existentes.
- Publicar a npm/PyPI (es el paso siguiente, separado).

## Impact

| Área | Impacto | Detalle |
|---|---|---|
| `tools/<tool>/tutorial.go` (×5) | New | Contenido y render del `--tutorial`; case nuevo en `run()` |
| `tools/<tool>/main.go` (×5) | Modified | Rutea `--tutorial`/`tutorial` a la nueva función |
| `npm/@open_harness/open-harness/bin/open-harness.js` | New | Comando `open-harness init` (spawnea los 5 init) |
| `pypi/open_harness/src/open_harness/__init__.py` | Modified | Entry point `open-harness init` equivalente |
| `npm/@open_harness/open-harness/package.json` | Modified | Agrega el bin `open-harness` |
| `tools/<tool>/CHANGELOG.md`, `.../open-harness/CHANGELOG.md` | New | 6 changelogs |
| `docs/CONFIGURATION.md`, `README.md` | New/Modified | Guía de config + enlaces |
| Cobertura / dependencias | Sin cambios | stdlib only (ADR-002); 100% en los 5 tools (ADR-011) |

## Risks

| Riesgo | Prob. | Mitigación |
|---|---|---|
| El `--tutorial` se desincroniza de los defaults reales de la config | Media | Un test por tool que verifica que el tutorial nombra las claves reales del `Config` struct |
| El `open-harness init` del launcher no llega a 100% cov (JS/Python, no Go) | Baja | El requisito de 100% es de los módulos Go (ADR-011); el launcher se cubre con un smoke test |
| El bin nuevo del meta rompe el empaquetado npm existente | Baja | Se agrega como bin adicional; los 4 bins de tools no cambian |

## Rollback Plan

Cada pieza es aditiva e independiente: revertir el commit de `--tutorial`, el de `open-harness init`
o el de docs por separado con `git revert`. Ningún cambio toca el comportamiento de `check`, así que
un rollback no afecta a los usuarios que ya corren los gates.

## Success Criteria

- [ ] `<tool> --tutorial` imprime la guía y devuelve exit 0 en los 5 tools; `--no-color` sin ANSI.
- [ ] El tutorial de cada tool nombra todas las claves de config que ese tool acepta (test).
- [ ] `open-harness init` crea los 5 `<tool>.json` en la raíz en una corrida.
- [ ] Existen los 6 `CHANGELOG.md` y `docs/CONFIGURATION.md`, enlazado desde el README.
- [ ] 100% de cobertura en los 5 tools; los 5 gates pasan sobre el repo.
