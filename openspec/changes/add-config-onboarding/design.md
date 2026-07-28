## Context

Feature de onboarding, aditivo y de bajo riesgo: un flag nuevo en los 5 tools, un comando en el
launcher del meta, y documentación. No toca `check` ni la carga de config. Restricciones vigentes:
stdlib only (ADR-002), 100% de statements en los módulos Go (ADR-011), TDD (ADR-013), archivos bajo
100 líneas (ADR-005), patrón `run([]string) int`.

## Decisiones

### D1. `--tutorial` como case en `run()`, contenido en `tutorial.go`

Cada tool agrega `case "--tutorial", "tutorial":` en el `switch` de `run()`, que llama a una función
en un archivo nuevo `tutorial.go`. El contenido es texto estático (un template por tool) que se
imprime a stdout. Respeta `--no-color` reutilizando el patrón de color existente del tool.

Alternativa descartada: un `--tutorial` interactivo. Rompe la testeabilidad a 100% (stdin), agrega
estado, y no encaja con el patrón stdlib. La guía impresa es idiomática (`git help`, `go help`).

### D2. El tutorial se testea contra las claves reales del `Config`

Para que la guía no se desincronice de la config, cada tool tiene un test que verifica que el texto
del tutorial menciona cada clave JSON del struct `Config` de ese tool. Si alguien agrega una clave a
la config y no al tutorial, el test falla. Es la misma disciplina que evitó la deriva de docs en F-018.

### D3. `open-harness init` delega, no reimplementa

El comando del meta NO reimplementa los defaults de cada config (sería duplicar y desincronizar).
Spawnea el `init` de cada tool (los binarios que el meta ya empaqueta como optionalDependencies).
Reporta qué creó y qué ya existía; no sobrescribe. Vive en el launcher: `bin/open-harness.js` (npm) y
el entry point de `pypi/open_harness` (Python). Como es código de distribución (no un módulo Go), no
aplica el 100% de ADR-011; se cubre con un smoke test que corre `open-harness init` en un temp dir.

### D4. Changelogs derivados de lo ya documentado

Los `CHANGELOG.md` no inventan historia: derivan de `docs/UPGRADING.md`, los mensajes de commit y los
ADRs. Formato *Keep a Changelog*. La `docs/CONFIGURATION.md` consolida lo que hoy está disperso entre
los README de cada tool y los ADR-003/008/014/018.

## Architecture Decision Impact

- No crea ni modifica ADRs. Reutiliza el patrón de CLI y config existente.
- **Extiende ADR-009**: el `--tutorial` es parte de la superficie que el propio repo ejercita.
- **No modifica** ADR-002 (stdlib) ni ADR-011 (los tools siguen al 100%).

## Testing strategy

TDD estricto. Por tool: test de que `--tutorial` imprime y sale 0; test de que menciona cada clave de
`Config` (D2); test de `--no-color` sin ANSI. El launcher del meta: smoke test que crea los 5 archivos
en un temp dir y verifica que no sobrescribe. Los changelogs y la guía de config se validan con un
test/gate liviano (existencia + que mencionan las claves y versiones correctas).
