# Release process

Cómo cortar un release de open-harness a npm y PyPI. El objetivo es que todo el
release sea **verificable con un solo comando** (`scripts/check-versions.sh`)
antes de publicar nada.

## Modelo de versiones

La **fuente de verdad** de la versión de un tool es la constante
`const version = "X.Y.Z"` de `tools/<tool>/main.go`.

- Los cuatro núcleos (`linelens`, `dupelens`, `secretlens`, `testlens`) comparten
  una **versión unificada**. El meta-paquete `open-harness` la comparte también.
- `scopelens` lleva su **propio ciclo de versiones** (empezó en `0.1.0`), porque
  se incorporó después.

Todo lo demás se **deriva** de ahí y `check-versions.sh` lo verifica:
`open-harness.json`, `README.md`, `AGENTS.md`, los `package.json` de npm (tool +
4 plataformas + pins), y los `pyproject.toml` de PyPI (tool + meta + pins) más el
`__version__` del meta.

## Pasos

1. **Bump.** Editar `const version` en el/los `tools/<tool>/main.go` que cambian.
   Los cuatro núcleos se mueven juntos; `scopelens` por separado.

2. **Sincronizar los manifiestos derivados** a la nueva versión:
   - `open-harness.json` (bloque de cada tool + `version` raíz del meta).
   - `README.md` y `AGENTS.md` (los tags `vX.Y.Z` de la tabla y del árbol).
   - npm: `bash scripts/build-npm-all.sh` sincroniza los `package.json` de los
     cuatro núcleos desde `main.go` y recompila los binarios. `bash scripts/build-npm.sh scopelens`
     para scopelens. El **meta** `npm/@open_harness/open-harness/package.json` se
     edita a mano: su `version` y sus `dependencies` a cada tool.
   - PyPI: los `pyproject.toml` de `pypi/open_harness_<tool>/` y de
     `pypi/open_harness/` (meta) se editan a mano, más el `__version__` de
     `pypi/open_harness/src/open_harness/__init__.py`. Los build scripts NO
     derivan estas versiones — por eso el gate del paso 3.

3. **Verificar** (el gate único):
   ```bash
   bash scripts/check-versions.sh
   ```
   Sale `0` si `main.go`, docs, npm y PyPI coinciden; `1` nombrando cada fuente
   divergente. **No publicar si esto no pasa.**

4. **Construir los artefactos** (si no se hizo en el paso 2):
   - npm: `bash scripts/build-npm-all.sh` (+ `build-npm.sh scopelens`).
   - PyPI: `bash scripts/build-pypi.sh <tool>` por cada tool + `open-harness` (meta).
     Genera los wheels por plataforma en `pypi/<pkg>/dist/`.

5. **Publicar** (desde `main`, ver flujo de ramas abajo):
   - npm: publicar **primero las 4 plataformas** de cada tool, después el
     `package.json` del tool, y por último el meta. `npm publish --access public`
     en cada dir. (Publicar el wrapper antes que sus plataformas deja a los
     usuarios con "platform not supported".)
   - PyPI: `twine check pypi/*/dist/*` y luego `twine upload` de cada
     `pypi/open_harness_<tool>/dist/*` y del meta.

6. **Tag**: `git tag -a vX.Y.Z -m "..."` y `git push origin vX.Y.Z`.

## Flujo de ramas

`main` es la rama de release; se publica **desde `main`**. El trabajo va en
`develop` → PR a `main` → merge → publicar. Ver [ADR-009](adr-009-proyecto-protegido-por-su-propia-herramienta.md).

## Peculiaridades conocidas

- **Meta PyPI bloqueado.** El paquete `open-harness` en PyPI da
  `400 "name too similar to an existing project"` por los `open-harness-<tool>`.
  No existe en PyPI y se acepta así: los usuarios de pip instalan los tools por
  separado. En npm el meta sí existe.
- **Propagación de npm.** Tras `npm publish`, `registry.npmjs.org` puede tardar
  unos minutos en reflejar la versión (la API muestra `versions: []` por caché de
  CDN). Un `403 "cannot publish over previously published versions"` al reintentar
  confirma que sí se publicó.
- **Secret scanning en fixtures.** GitHub bloquea el push si detecta credenciales
  con formato real en el historial. Los fixtures de secretlens se **ensamblan en
  runtime** (partes concatenadas) para no dejar literales completos; los ejemplos
  en docs usan valores partidos o excluidos (`docs/`, `**/CHANGELOG.md` salen del
  scan de secretlens).
- **Presupuesto de scopelens.** Agregar un tool nuevo o un bump que toca >
  `maxFiles` archivos supera el propio gate de `scopelens`; esos commits de
  release usan `git commit --no-verify` tras verificar los otros gates a mano.

## Un solo comando de verificación

`scripts/check-versions.sh` es el chequeo que cierra la deriva de versiones entre
las ~10 fuentes por tool (main.go, manifiesto, docs, npm ×6, PyPI ×2). Correrlo
antes de publicar evita releases con manifiestos desincronizados.
