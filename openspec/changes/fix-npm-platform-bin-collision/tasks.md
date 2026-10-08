# Tasks — fix-npm-platform-bin-collision

`design.md` se omite: el cambio quita un campo de manifiestos y la decisión está explicada en
`proposal.md` (ninguna de las condiciones del artefacto aplica).

## 1. Reproducción (red)

- [x] 1.1 Experimento A/B con tarballs locales del wrapper de dupelens y sus 4 plataformas
  (`optionalDependencies` por `file:`). Verificación: con `bin` en las plataformas,
  `node_modules/.bin/` queda vacío y `npx --no-install dupelens` falla; sin `bin`, imprime la versión.
- [x] 1.2 Mismo experimento con el meta sobre la variante sin `bin`. Verificación: `npx dupelens version`
  funciona instalando solo el meta.
- [x] 1.3 (red) Gate en `scripts/check-versions.sh`: falla si un paquete de plataforma declara `bin`.
  Verificación: con los manifiestos actuales el script sale con 1 y nombra los 20 paquetes.

## 2. Corrección (green)

- [x] 2.1 Quitar `bin` de los 20 `npm/@open_harness/<tool>-<plataforma>/package.json`. Verificación:
  `check-versions.sh` en verde.

## 3. Release

- [ ] 3.1 Bump: linelens/secretlens/testlens 0.3.3 → 0.3.4, dupelens 0.5.0 → 0.5.1, scopelens
  0.2.1 → 0.2.2, meta 0.3.6 → 0.3.7 (npm y PyPI), y CHANGELOGs. Verificación: `check-versions.sh` en
  verde.
- [ ] 3.2 Build: `build-npm-all.sh`, `build-npm.sh scopelens` y `build-pypi.sh` de los cinco tools y el
  meta. Verificación: cada binario linux-x64 imprime su versión nueva y `twine check` pasa.
- [ ] 3.3 Tests de los cinco tools y los cuatro `_shared`, y los cinco gates sobre el repo.
  Verificación: todo en verde.
- [ ] 3.4 Publicar npm (20 plataformas, 5 wrappers, meta) y PyPI (5 tools, `open-harness-suite`); tags.
  Verificación: `npm publish`/`twine upload` sin errores.
- [ ] 3.5 Verificación desde el registry. Verificación: en proyectos limpios, `npm install` + `npx <tool>
  version` para los cinco wrappers y para el meta; `pip install open-harness-suite==0.3.7`.

## 4. Cierre

- [ ] 4.1 Marcar F-025 en `.agent/feature-list.json`, actualizar `.agent/claude-progress.txt` y archivar
  el change.
