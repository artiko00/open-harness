# npm: el comando de cada tool queda enlazado tras `npm install`

Feature ID: **F-025** (`.agent/feature-list.json`)
Affected tools: **linelens, dupelens, secretlens, testlens, scopelens** y el meta `@open_harness/open-harness`
(solo la distribución por npm; el código Go y PyPI no cambian de comportamiento).
Risk: **low** (se quita un campo de 20 manifiestos; el wrapper no lo usa)

## Why

Tras `npm install --save-dev @open_harness/dupelens`, `npx dupelens check` falla con
`404 Not Found - GET https://registry.npmjs.org/dupelens`: el comando no queda en `node_modules/.bin/` y
npx lo busca en el registry. Se detectó el 2026-10-08 verificando dupelens 0.5.0 y pasa igual con
0.4.1, con los cinco tools y con el meta (solo queda enlazado `open-harness`).

Causa: los 20 paquetes de plataforma (`@open_harness/<tool>-<os>-<cpu>`) declaran
`"bin": { "<tool>": ... }`, el mismo nombre que el bin del wrapper. Con npm 10, las dependencias
opcionales de las otras plataformas fallan por `os`/`cpu` y, al revertirlas, npm borra
`node_modules/.bin/<tool>`: el link que acababa de crear para el wrapper. Reproducido con tarballs
locales: con `bin` en las plataformas el link desaparece; sin él, `npx dupelens` funciona, tanto con el
wrapper como con el meta.

## What Changes

- Los paquetes de plataforma dejan de declarar `bin`. El wrapper ya resuelve el binario con
  `require.resolve('<plataforma>/package.json')` y nunca usó ese campo.
- `scripts/check-versions.sh` falla si un paquete de plataforma vuelve a declarar `bin`, para que la
  regresión no llegue a un publish.
- Release de los cinco tools para republicar los paquetes corregidos: linelens, secretlens y testlens
  0.3.4; dupelens 0.5.1; scopelens 0.2.2; meta 0.3.7. PyPI se republica con las mismas versiones para
  mantener la versión única por tool que verifica `check-versions.sh`.

## Capabilities

### New Capabilities

_Ninguna._

### Modified Capabilities

- `release-metadata`: los paquetes npm de plataforma no exponen comandos; el gate de versiones lo
  verifica.

## Scope

### In Scope
- Quitar `bin` de `npm/@open_harness/<tool>-<plataforma>/package.json` (20 archivos).
- Chequeo en `scripts/check-versions.sh`.
- Bump, CHANGELOGs y publicación en npm (26 paquetes) y PyPI (5 tools + `open-harness-suite`).
- Verificación desde el registry: `npm install` + `npx <tool> version` para los cinco tools y el meta.

### Out of Scope
- Cambios en el código Go de los tools.
- El shim del meta (`bin/<tool>.js`): se conserva para gestores que no elevan los bins transitivos
  (pnpm, yarn Berry).
- Composer/Packagist.

## Impact

| Área | Impacto | Detalle |
|---|---|---|
| `npm/@open_harness/*-*/package.json` | Modified | sin `bin` (20 archivos) |
| `scripts/check-versions.sh` | Modified | falla si una plataforma declara `bin` |
| `tools/*/main.go`, manifiestos npm/PyPI, `open-harness.json`, README, AGENTS | Modified | bump de versión |
| CHANGELOGs | Modified | entrada de la versión |

## Rollback Plan

Republicar la versión anterior de cada paquete como `latest` con `npm dist-tag`. El cambio no toca los
binarios ni la config.

## Success Criteria

- [ ] En un proyecto limpio, `npm install --save-dev @open_harness/<tool>` deja `node_modules/.bin/<tool>`
      y `npx <tool> version` imprime la versión nueva, para los cinco tools.
- [ ] `npm install @open_harness/open-harness` deja disponibles los cinco comandos.
- [ ] `bash scripts/check-versions.sh` falla si se agrega `bin` a un paquete de plataforma.
