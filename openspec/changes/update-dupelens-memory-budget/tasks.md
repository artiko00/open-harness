# Tasks — update-dupelens-memory-budget

## 0. Spike de validación

- [x] 0.1 Script de medición fuera del repo: corre un binario de dupelens en
  `systemd-run --user --scope -p MemoryMax=3G -p MemorySwapMax=0` y reporta pico de RSS y tiempo.
  Verificación: reproduce los valores del proposal con el binario 0.4.1 (k = 2/4/8/16 copias de
  `asyncio`, stdlib, `fortinet/fortios`).
- [x] 0.2 Prototipo descartable de D1 + D2 sobre una copia del paquete. Verificación: 32 copias de
  `asyncio` y `fortios` quedan bajo 1 GiB y la stdlib queda en ≤ 3x el fuente. Si no se cumple,
  revisar `design.md` antes de seguir. [skip] D1 y D2 comparten el modelo de datos nuevo: se validó
  directamente la implementación real en 6.1 en lugar de un prototipo aparte.

## 1. Semillas ancladas y fusión por diagonal (D1)

- [x] 1.1 (red) Test: tres archivos con el mismo bloque byte-idéntico producen exactamente los pares
  `a`-`b` y `a`-`c` (`exactCount` 2). Verificación: `go test -run Anchored` falla.
- [x] 1.2 (red) Test: un bloque presente solo en `b` y `c` se reporta como `b`-`c`; con tres copias
  `exact`, `--fail` termina con exit 1. Verificación: falla en rojo.
- [x] 1.3 (red) Test de memoria sensible a `-short`: `TotalAlloc` de 32 copias < 2,5 × `TotalAlloc` de 16
  copias. Verificación: falla con el algoritmo actual.
- [x] 1.4 (green) Semillas ancladas por clase verificada y orden por (par, diagonal, idx) con corridas
  contiguas, alimentando `mergeOnePair`. Verificación: 1.1-1.3 en verde y la suite completa de dupelens
  en verde.
- [x] 1.6 (red → green) Test: con `a` que comparte solo la mitad del bloque de `m` y `n`, el par `m`-`n`
  se reporta completo. Surgió en 6.1 (`encodings/cp850.py` ↔ `cp852.py` se fragmentaba). Se resolvió
  extendiendo cada semilla sobre su diagonal (`extend.go`). Verificación: el test pasa y stdlib baja de
  22 a 1 `exact` sin explicar.
- [x] 1.7 (red → green) Test: en código periódico ningún hallazgo declara más tokens que su tramo; el
  conteo pasa a ser el tramo en A. Verificación: el test pasa.
- [x] 1.5 (refactor) Archivos tocados ≤ 100 líneas de código y 100 % de cobertura. Verificación:
  `linelens check --fail` y `go tool cover -func`.

## 2. Representación compacta (D2)

- [x] 2.1 (red) Tests del diccionario de tokens: mismo string → mismo id; la tabla de normalización
  produce `ID`, `NUM` y keywords igual que `normalizeValue`. Verificación: falla en rojo.
- [x] 2.2 (red) Test del prefiltro: una ventana única no se materializa como huella; una repetida
  sí; un falso positivo del bitset no produce un hallazgo. Verificación: falla en rojo.
- [x] 2.3 (red) Test del hash de 61 bits: `mulMod` contra `math/big`, reducción en el borde del módulo,
  y la verificación literal ejercitada con una colisión inyectada (hashes de token igualados). No
  había un fixture de colisión `1e9+7` reutilizable. Verificación: falla en rojo.
- [x] 2.4 (green) Tokens internados (id + línea), `normOf[id]`, huellas (hash, posición) ordenadas,
  bitsets de prefiltro y hash módulo 2^61−1. Verificación: 2.1-2.3 y la suite completa en verde,
  incluido `TestMemory_scanStaysProportional`.
- [x] 2.5 (refactor) Eliminar `Fingerprint`, `groupByHash` y la copia `norm` si quedaron sin uso;
  archivos ≤ 100 líneas y 100 % de cobertura. Verificación: `go vet`, `linelens` y cobertura.

## 3. Resolución del presupuesto (D3)

- [x] 3.1 (red) Tests de parseo: `512MiB`, `2GiB`, `40%` válidos; `0MiB`, `0%`, `101%`, `mucho` y `2GB`
  inválidos con mensaje de formato. Verificación: falla en rojo.
- [x] 3.2 (red) Tests de precedencia con `t.Setenv` y configs en `t.TempDir()`: flag > env > clave
  `maxMemory` (en `dupelens.json`, `pyproject.toml`, `package.json` y `composer.json`) > `auto`.
  Verificación: falla en rojo.
- [x] 3.3 (red) Tests de `auto` con memoria disponible inyectada: 28 GiB → 1 GiB; 2 GiB → 512 MiB;
  indeterminable → 1 GiB; `40%` indeterminable → 1 GiB con aviso en stderr. Verificación: falla en rojo.
- [x] 3.4 (green) `membudget.go`: parseo, resolución con origen (flag/env/config/auto) y clave
  `maxMemory` en `Config`; flag `--max-memory` en `check_cmd.go`; valor inválido → exit 1.
  Verificación: 3.1-3.3 en verde y los tests de config de los cuatro formatos en verde.

## 4. Memoria disponible por plataforma (D4)

- [x] 4.1 (red) Tests con fixtures de texto: `MemAvailable` de `/proc/meminfo`; cgroup v2
  (`memory.max` numérico y `max`, `memory.current`) y v1; recorrido de ancestros tomando el menor
  margen; ruta del cgroup desde `/proc/self/cgroup`. Verificación: falla en rojo.
- [x] 4.2 (green) Funciones puras de parseo + `memavail_linux.go`. Verificación: 4.1 en verde y, en
  un `systemd-run --scope -p MemoryMax=1G`, el presupuesto `auto` informado con `--verbose` refleja
  el margen del cgroup.
- [x] 4.3 (green) `memavail_darwin.go` (`hw.memsize`), `memavail_windows.go` (`GlobalMemoryStatusEx`)
  y `memavail_other.go` (indeterminable). Verificación:
  `GOOS=darwin GOARCH=arm64 go build`, `GOOS=darwin GOARCH=amd64 go build` y
  `GOOS=windows GOARCH=amd64 go build` compilan; `bash scripts/build-npm.sh dupelens` genera los
  cuatro binarios.

## 5. Hacer cumplir el presupuesto (D5, D6)

- [x] 5.1 (red) Test con el medidor inyectado por encima del presupuesto: exit 2 con y sin `--fail`,
  stdout vacío y stderr con el presupuesto, su origen y las opciones `--max-memory`,
  `DUPELENS_MAX_MEMORY`, `maxMemory` y `exclude`. Verificación: falla en rojo.
- [x] 5.2 (red) Test: dentro del presupuesto, el reporte y el exit code son idénticos a los de una
  ejecución sin puntos de control. Verificación: falla en rojo si el guard altera la salida.
- [x] 5.3 (green) `memguard.go`: `debug.SetMemoryLimit` al 90 %, lectura de `runtime/metrics` y
  puntos de control por archivo, cada 64 Ki semillas y antes del reporte; error tipado que
  `runCheck` traduce a exit 2. Verificación: 5.1-5.2 en verde.
- [x] 5.4 (refactor) Archivos ≤ 100 líneas, 100 % de cobertura en `tools/dupelens`. Verificación:
  `go test ./... -coverprofile` y `linelens check --fail`.

## 6. Verificación end-to-end

- [x] 6.1 Con el script de 0.1, comparar el binario nuevo contra
  `npm/@open_harness/dupelens-linux-x64/bin/`: 32 copias de `asyncio` (69 MB) y `fortios` (45 MB) bajo
  1 GiB; stdlib en 25 MB (2,2x el fuente); `site-packages` completo (226 MB) en 388 MB, donde 0.4.1 pasó
  de 22 GB. Los hallazgos de stdlib, numpy, matplotlib y yt_dlp se clasificaron contra 0.4.1 (se
  mantienen / implícitos por anclaje / sin explicar); los "sin explicar" inspeccionados son conteos
  inflados de 0.4.1 (ver design D1) y `--fail` coincide en los cuatro árboles.
- [x] 6.2 Exceso real: `--max-memory 20MiB` sobre `fortios` termina con exit 2 y pico de 18 MB, y dentro
  de un cgroup de 40 MB el default `auto` también corta con exit 2; el cgroup nunca mata el proceso.
  (La stdlib no sirve para esta prueba: con 25 MB de pico entra holgada en 64 MiB.)

## 7. Documentación

- [x] 7.1 `docs/adr-024-dupelens-presupuesto-de-memoria.md` (D1, D3, D5, D6 y la justificación de
  cobertura de los adaptadores por SO) y sección nueva en ADR-012. Verificación: enlaces válidos
  desde AGENTS.md.
- [x] 7.2 `tutorial.go` y `help.go` documentan `maxMemory` y `--max-memory`; `init` no escribe la clave
  (ausente = `auto`, ver design D3). Verificación: tests de `--tutorial` actualizados en verde.
- [x] 7.3 README de dupelens, `docs/CONFIGURATION.md`, `docs/UPGRADING.md` (anclaje y exit 2),
  CHANGELOG y la tabla de exit codes de AGENTS.md (sección 4.1). Verificación: revisión del diff.
  Incluye los READMEs de npm, PyPI y Composer de dupelens (el de Composer tenía el texto de linelens)
  y la línea de dupelens en los READMEs del meta paquete.
- [x] 7.4 Bump dupelens 0.4.1 → 0.5.0 en `main.go`, `open-harness.json`, README y AGENTS.md, y meta
  0.3.5 → 0.3.6 (npm y PyPI, con el pin de dupelens). Verificación: `bash scripts/check-versions.sh` en verde.

## 8. Quality gates y cierre

- [x] 8.1 Tests de los cinco tools y los cuatro `_shared`:
  `for m in tools/*/ tools/_shared/*/; do (cd $m && go test ./...); done` en verde.
- [x] 8.2 `linelens`, `dupelens`, `secretlens`, `testlens` y `scopelens` `check --fail` sobre el repo
  en verde.
- [ ] 8.3 Marcar los steps de F-024 en `.agent/feature-list.json` y actualizar
  `.agent/claude-progress.txt` con lo hecho y los próximos pasos (release).
