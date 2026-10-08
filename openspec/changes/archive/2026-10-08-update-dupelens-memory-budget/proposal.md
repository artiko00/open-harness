# dupelens: memoria acotada en proyectos grandes

Feature ID: **F-024** (`.agent/feature-list.json`)
Affected tools: **dupelens** (0.4.1). linelens, secretlens, testlens y scopelens no cambian; ningún
módulo de `tools/_shared/` cambia.
Risk: **medium** (cambia el conteo de hallazgos cuando hay tres o más copias de un mismo bloque, y
agrega el exit code `2` a dupelens)

## Why

Usuarios con proyectos grandes reportan que `dupelens check` consume toda la RAM. Se reprodujo el 2026-10-08:
sobre `/usr/lib/python3.14` (237 MB de `.py`) el proceso pasó de 1,5 GB a 22 GB en dos minutos, con el
swap lleno, y `systemd-oomd` mató el scope de la terminal entera (1114 procesos).

Las mediciones aisladas (cgroup con `MemoryMax=3G`) separan dos causas:

| Árbol | Fuente | RSS pico |
|---|---|---|
| stdlib de Python, numpy, matplotlib (poca duplicación) | 5-11 MB | 20-23x el fuente |
| `asyncio` copiado k = 2 / 4 / 8 / 16 veces | 1 / 2 / 4 / 8 MB | 30 / 67 / 221 / 791 MB |
| `asyncio` copiado 32 veces | 16 MB | > 3 GB (OOM) |
| `ansible_collections/fortinet/fortios` (código generado) | 20 MB | > 3 GB (OOM) |

1. **Eje cuadrático.** `detect` enumera todos los pares de cada bucket de hash y materializa un
   `Match` por ventana y por par antes de fusionar: L × k(k−1)/2 objetos para un bloque de L ventanas
   repetido k veces. Cada duplicación de k cuadruplica memoria y tiempo. Es lo que hace caer a proyectos
   de apenas 16-20 MB con código generado o copiado.
2. **Constante lineal alta.** Cada token se retiene como string dos veces (crudo y normalizado), más
   dos `Fingerprint` de 40 B y un `map[uint64][]Fingerprint`: ~20x el fuente aun sin duplicación.

Además, la herramienta no tiene ningún techo: cuando corre en un pre-commit con varios proyectos o
worktrees en paralelo, una sola instancia puede tumbar la sesión del usuario. Limitar el runtime de Go
no alcanza: con `GOMEMLIMIT=300MiB` el caso k = 16 bajó a 512 MB (el heap vivo ya era de 523 MB) y tardó
16 s en lugar de 3,7 s por thrashing del GC.

## What Changes

- **Hallazgos anclados a la primera ocurrencia.** Para cada grupo de ventanas idénticas se empareja cada
  ocurrencia solo con la primera en orden canónico (archivo, posición): k−1 pares en lugar de
  k(k−1)/2. Con tres copias idénticas A, B y C se reportan A-B y A-C; B-C queda implícito.
  **BREAKING** en el conteo de hallazgos (no en el resultado de `--fail`).
- **Fusión por diagonal** de las semillas ancladas: el número de semillas queda acotado por el número
  de ventanas del proyecto, y desaparece el eje cuadrático en memoria y CPU. Cada semilla se extiende
  hasta el clon máximo de su diagonal, para que un cambio de ancla a mitad de bloque no lo fragmente.
- **Conteo de tokens = tramo real.** 0.4.x sumaba un token por cada par de ventanas fusionado, aunque
  vinieran de alineaciones distintas: en código periódico o repetitivo declaraba bloques de "50
  tokens" que en ninguna alineación eran contiguos (medido: 31 tokens reales en `matplotlib/pyplot.py`
  declarados como 71). **BREAKING** en los conteos: esos hallazgos dejan de reportarse. En stdlib,
  numpy, matplotlib y yt_dlp el resultado de `--fail` no cambia.
- **Representación compacta** (interna, sin cambio observable): tokens internados (id + línea), forma
  normalizada por tabla en lugar de una copia, huellas (hash, posición) en un slice ordenado y
  prefiltro que descarta las ventanas únicas antes de materializarlas.
- **Presupuesto de memoria `maxMemory`**: absoluto (`"2GiB"`, `"512MiB"`) o porcentaje de la memoria
  disponible al arrancar (`"40%"`). Default `auto` = mín(1 GiB, 25 % de la memoria disponible). Se fija
  con `--max-memory`, la variable de entorno `DUPELENS_MAX_MEMORY` o la clave `maxMemory` en cualquier
  archivo de la cadena de configuración, en ese orden de precedencia.
- **Exceso del presupuesto: exit `2`** con un mensaje en stderr que nombra el presupuesto aplicado, su
  origen y cómo subirlo. **BREAKING**: dupelens incorpora el exit code `2` ("no se pudo medir") que ya
  usa scopelens.

## Capabilities

### New Capabilities

_Ninguna._

### Modified Capabilities

- `duplicate-detection`: el techo de memoria pasa de "proporcional al fuente" a acotado por un
  presupuesto configurable y sin eje cuadrático; los hallazgos se anclan a la primera ocurrencia; el
  conteo de tokens refleja el tramo duplicado real.
- `cli-contract`: el contrato de exit codes incorpora el `2` de dupelens cuando se excede el
  presupuesto de memoria.

## Scope

### In Scope
- Semillas ancladas, fusión por diagonal y representación compacta en `tools/dupelens/`.
- `--max-memory`, `DUPELENS_MAX_MEMORY` y `maxMemory`, con `auto` como default.
- Detección de memoria disponible: `/proc/meminfo` y cgroup (v2 y v1) en Linux, `hw.memsize` en macOS,
  `GlobalMemoryStatusEx` en Windows; todo con stdlib (ADR-002).
- `debug.SetMemoryLimit` y puntos de control con `runtime/metrics`; exit `2` al exceder.
- ADR nuevo; `init`, `--tutorial`, README, CHANGELOG, `docs/CONFIGURATION.md` y `docs/UPGRADING.md`.
- Bump dupelens 0.4.1 → 0.5.0.

### Out of Scope
- Coordinación entre procesos (lock compartido para serializar instancias en paralelo): el porcentaje
  sobre la memoria *disponible* y el tope absoluto acotan el caso sin estado compartido.
- Partición por hash en varias pasadas o derrame a disco para escanear por encima del presupuesto:
  con la representación compacta el costo baja a un múltiplo chico del fuente, y el exceso se
  informa con exit `2` en lugar de degradar.
- Winnowing u otro muestreo de huellas: pierde la garantía de detección completa.
- Reportar grupos de clones (k ocurrencias en un solo hallazgo): cambia el schema del reporte.
- Cambios en los otros cuatro tools o en `tools/_shared/`.

## Impact

| Área | Impacto | Detalle |
|---|---|---|
| `tools/dupelens/duplicates.go`, `match_util.go`, `merge.go` | Modified | semillas ancladas y fusión por diagonal |
| `tools/dupelens/tokenizer.go`, `normalize.go`, `fingerprint.go`, `collect.go`, `scanner.go` | Modified | tokens internados, tabla de normalización, huellas compactas y prefiltro |
| `tools/dupelens/membudget*.go` | New | resolución del presupuesto, memoria disponible por SO y puntos de control |
| `tools/dupelens/config.go`, `check_cmd.go` | Modified | clave `maxMemory`, flag `--max-memory`, env var y exit `2` |
| `tools/dupelens/init_cmd.go`, `tutorial.go`, `help.go` | Modified | documentar la clave y el flag |
| `openspec/specs/cli-contract` | Modified | exit `2` en dupelens |
| `docs/adr-024-*.md` | New | presupuesto de memoria y anclaje de hallazgos |
| Dependencias / coverage | Sin cambios | stdlib (ADR-002); 100 % statement coverage (ADR-011) |
| Distribución | dupelens solamente | npm (5 paquetes) + PyPI (wheels de dupelens) + meta |

## Rollback Plan

Revertir el commit y republicar 0.4.1 como versión recomendada. Para quien ya actualizó y necesita
escanear algo mayor que el default sin degradar: `--max-memory 100%` (o la clave equivalente) retira
el tope absoluto. El conteo anclado no tiene escape por config.

## Success Criteria

- [ ] `asyncio` copiado 32 veces y `fortinet/fortios` terminan por debajo de 1 GiB de RSS.
- [ ] Sobre código con poca duplicación, el RSS pico baja a ≤ 3x el tamaño del fuente.
- [ ] Con tres copias idénticas se reportan dos hallazgos anclados a la primera copia.
- [ ] Un escaneo que excede el presupuesto termina con exit `2` y un mensaje con el presupuesto y su origen.
- [ ] `--max-memory`, `DUPELENS_MAX_MEMORY` y `maxMemory` respetan la precedencia flag > env > config > auto.
- [ ] Dentro de un cgroup con `memory.max`, `auto` respeta el límite del cgroup.
- [ ] 100 % statement coverage; dupelens pasa su propio gate sobre el repo.
