# ADR-024: dupelens — presupuesto de memoria y exit 2 al excederlo

**Estado:** Aceptado
**Fecha:** 2026-10-08
**Aplica a:** dupelens
**Extiende:** ADR-012 (Rabin-Karp sobre AST), ADR-023 (exit 2 = "no pude medir"), ADR-011 (cobertura)

## Contexto

Usuarios con repositorios grandes reportaron que `dupelens check` consumía toda la RAM. Se reprodujo el
2026-10-08: sobre un `site-packages` de 226 MB de `.py`, 0.4.1 pasó de 1,5 GB a 22 GB en dos minutos y
`systemd-oomd` mató el scope de la terminal entera. Las mediciones aisladas separaron dos causas: un eje
cuadrático en el número de copias de un bloque y una constante lineal de ~20x el fuente (ver la
extensión de ADR-012 para el algoritmo que las elimina).

Aun con memoria lineal, un repositorio suficientemente grande —o varios proyectos corriendo sus hooks en
paralelo— puede llevar la máquina al límite. La herramienta no tenía ningún techo, y el límite del
runtime de Go (`GOMEMLIMIT`) es blando: con 300 MiB, un caso cuyo heap vivo era de 523 MB terminó en
512 MB de RSS y tardó 16 s en lugar de 3,7 s por thrashing del GC.

## Decisión

### Un solo valor, absoluto o porcentaje

`maxMemory` (clave), `--max-memory` (flag) y `DUPELENS_MAX_MEMORY` (entorno) aceptan `<n>MiB`, `<n>GiB`
o `<n>%`, con precedencia flag > entorno > config > `auto`. Un valor explícito se aplica tal cual.

`auto` = **el menor entre 1 GiB y el 25 % de la memoria disponible**. El porcentaje es sobre la memoria
*disponible*, no la total: con varias instancias en paralelo, cada instancia nueva ve lo que las
anteriores ya ocupan. El tope de 1 GiB acota las máquinas grandes.

Se descartaron dos claves combinadas con mínimo (`maxMemory` + `maxMemoryPercent`): quien sube el tope
absoluto podría recibir menos por el porcentaje sin verlo en su config.

### Memoria disponible por plataforma, solo stdlib

| Plataforma | Fuente |
|---|---|
| Linux | `MemAvailable` de `/proc/meminfo` y el cgroup propio (v2 `memory.max − memory.current`, v1 `limit_in_bytes − usage_in_bytes`), recorriendo los ancestros: el menor |
| macOS | `hw.memsize` vía `syscall.Sysctl` (la RAM total: la disponible requiere cgo) |
| Windows | `GlobalMemoryStatusEx` vía `syscall.NewLazyDLL` |
| Otras | indeterminable → 1 GiB |

### Cómo se hace cumplir

1. `debug.SetMemoryLimit` al 90 % del presupuesto, para que el GC recolecte antes de llegar al tope.
2. Puntos de control después de cada archivo, antes de cada reserva grande y al crecer los slices de
   candidatos, semillas y hallazgos. Cada uno lee `runtime/metrics` (memoria tomada del SO menos la
   devuelta, la misma magnitud que acota `SetMemoryLimit`). Ante un exceso aparente se fuerza
   `debug.FreeOSMemory` y se vuelve a medir: solo el exceso que sobrevive cuenta.

Se descartaron `setrlimit(RLIMIT_AS)` (el runtime de Go reserva espacio virtual de más y no existe en
Windows) y un goroutine vigilante (no determinista en tests; los puntos de control ya acotan el exceso).

### Exceder el presupuesto es exit 2

Igual que en scopelens (ADR-023), `2` significa "no se pudo medir": dupelens corta sin reporte parcial,
con o sin `--fail`, y explica el presupuesto, su origen y cómo subirlo. Un pre-commit lo trata como
falla: un escaneo inconcluso nunca cuenta como aprobado. Los errores de uso y de config siguen en `1`.

## Consecuencias

- **Positivo:** ninguna ejecución de dupelens puede tumbar la sesión del usuario; el exceso termina en
  un error explicado en lugar de un OOM-kill. Con el algoritmo nuevo, 226 MB de fuente entran en 388 MB.
- **Negativo:** un repositorio que no entra en el presupuesto bloquea el pre-commit hasta que se suba
  el valor o se excluyan rutas. El mensaje nombra las tres formas de hacerlo.
- **Negativo:** el porcentaje se mide una vez al arrancar; instancias lanzadas en el mismo instante ven
  la misma cifra. La coordinación entre procesos queda fuera de alcance; el tope de 1 GiB acota el caso.
- **ADR-011 (cobertura 100 %):** `memavail_darwin.go` y `memavail_windows.go` no se compilan en Linux,
  donde se mide la cobertura. Se aceptan sin cubrir porque se limitan a una llamada al SO; toda la
  lógica (parseo de `/proc/meminfo` y del cgroup, combinación, resolución del presupuesto) vive en
  funciones portables con 100 % de cobertura, y los cuatro binarios se compilan en cada release.
