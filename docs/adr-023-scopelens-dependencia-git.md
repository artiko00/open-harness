# ADR-023: scopelens — git como dependencia de binario externo

**Estado:** Aceptado
**Fecha:** 2026-07-28
**Aplica a:** scopelens
**Extiende:** ADR-018 (config multi-ecosistema), ADR-009 (proyecto protegido por su propia herramienta)

## Contexto

scopelens mide un "presupuesto de archivos por PR": cuenta los archivos que cambian entre la rama actual y su base (`origin/HEAD`, `main` o `master`) y falla si el diff supera `maxFiles`. Para conocer ese diff necesita, inevitablemente, información que solo vive en el historial de git: qué archivos difieren entre dos refs.

Esto choca de frente con **ADR-002 (cero dependencias externas)**, que prohíbe cualquier módulo de Go fuera de la stdlib y de `tools/_shared/*`. La pregunta es: ¿scopelens viola ADR-002 al depender de git?

Hay dos formas de obtener el diff:

1. **Parsear el directorio `.git/` a mano** (leer refs, packfiles, árboles) desde Go puro con stdlib.
2. **Invocar el binario `git`** vía `os/exec` y leer su salida.

## Decisión

**scopelens invoca el binario `git` externo mediante `os/exec`. git es una dependencia de _binario de sistema_, no una dependencia de _runtime de Go_.**

ADR-002 restringe el grafo de módulos Go (`go.mod`): sigue siendo stdlib + `tools/_shared/*`, sin una sola línea en `require`. La presencia de git en el `PATH` es del mismo orden que la presencia de `sh`, `go` o `lefthook`: una herramienta del entorno de desarrollo, no un paquete importado. **ADR-002 queda intacto.**

### Por qué se descarta parsear `.git/` a mano

- Reimplementar la lectura de refs empaquetadas, packfiles con delta-encoding y árboles es cientos de líneas de código de altísimo riesgo, contra ADR-005 (100 líneas por archivo) y ADR-013 (TDD con 100% de cobertura sobre cada rama de un parser binario).
- Reproduciría lógica que git ya resuelve de forma canónica; cualquier divergencia con `git diff` sería un bug silencioso en la medición.
- El objetivo de scopelens es *medir*, no *reimplementar git*. La fuente de verdad del diff debe ser git mismo.

### Contrato de la invocación

- Se usa `git --no-optional-locks ...` para no tomar locks del working tree: scopelens **solo lee**, nunca debe interferir con un `git` que el desarrollador esté corriendo en paralelo (alineado con ADR-009: la herramienta que protege el repo no puede degradarlo).
- Toda invocación corre bajo un **`context.Context` con timeout**: un git colgado (red, filesystem lento, repo corrupto) no puede colgar el hook. Vencido el timeout, el proceso se cancela.
- La resolución de la base es por precedencia `--base > config > origin/HEAD > main > master`; si ninguna ref existe, se falla ruidoso pidiendo `--base` explícito.

### exit 2 = "no pude medir"

scopelens distingue tres resultados:

- **exit 0** — medido, dentro del presupuesto.
- **exit 1** — medido, excede el presupuesto (solo con `--fail`).
- **exit 2** — **no pude medir**: git ausente en el `PATH`, timeout, ref base irresoluble, config inválida o error de uso.

exit 2 nunca se confunde con "presupuesto excedido". Un entorno sin git no reporta un falso 0 ni un falso 1: reporta honestamente que la medición no fue posible, y el hook decide (por defecto, un exit 2 en pre-commit frena el commit con un mensaje accionable en vez de pasar en verde).

## Consecuencias

**Positivo:**

- ADR-002 se respeta al pie de la letra: `go.mod` sin dependencias externas.
- La medición es exactamente `git diff`; cero divergencia con la fuente de verdad.
- `--no-optional-locks` + timeout hacen la herramienta segura para correr dentro de un hook, sin carreras con el git del usuario.
- exit 2 como estado explícito evita el antipatrón de "fallar en verde" (ADR-009).

**Negativo:**

- scopelens requiere `git` en el `PATH`. En un entorno sin git (algunos CI minimalistas), scopelens sale con exit 2. Es aceptable: sin git no hay diff que medir.

**Neutral:**

- La superficie de `os/exec` se concentra en un único runner (`gitcmd.go`), testeable inyectando un `gitRunner` fake; el 100% de cobertura (ADR-011/013) se mantiene sin invocar git real en los tests unitarios.
