# ADR-018: Config unificada multi-ecosistema (pyproject.toml + composer.json)

**Fecha:** 2026-05-13
**Estado:** Aceptada
**Aplica a:** linelens, dupelens, secretlens, testlens (todos)

## Contexto

[ADR-014](adr-014-config-en-package-json.md) introdujo fallback de config desde `package.json` (key `<tool>`) en cada uno de los 4 tools, alineándose con la convención de prettier, eslintConfig, stylelint en el ecosistema Node.

Al expandir distribución a **PyPI** y **Packagist**, los ecosistemas Python y PHP tienen sus propios archivos central de proyecto:

- **Python (PEP 518)**: `pyproject.toml` con tablas `[tool.<name>]`. Convención de ruff, black, mypy, pytest.
- **PHP**: `composer.json` con `extra.<vendor>.<config>`. Convención de phpstan, php-cs-fixer.

Si los binarios Go no leen esos archivos, los usuarios de Python y PHP se ven forzados a mantener `<tool>.json` separados — fricción innecesaria que rompe la promesa de "config en el archivo central del ecosistema".

## Decisión

Cada uno de los 4 binarios soporta **5 fuentes de configuración** ordenadas por precedencia:

```
1. CLI flags (--max, --min-tokens, --no-color, …)        ← siempre ganan
2. <tool>.json en la raíz                                ← archivo dedicado
3. pyproject.toml → [tool.<name>]                        ← Python idiomático
4. package.json → "<name>": { … }                         ← Node idiomático (ya existente)
5. composer.json → "extra": { "open-harness": { "<name>": … } }  ← PHP idiomático
6. defaults compilados en el binario                     ← última instancia
```

### Comportamiento en proyectos políglotas

Si un proyecto tiene `pyproject.toml` **y** `package.json` simultáneamente (caso real: proyecto Python con frontend Node), gana el **primero por orden de la chain**. Es decir: `pyproject.toml` sobreescribe a `package.json` para los campos que define, pero `package.json` aporta los campos que `pyproject.toml` omite.

Esto NO es una fusión profunda; es un fallback secuencial: el primer archivo que tenga un valor para un campo gana. Los archivos posteriores en la chain solo se consultan para campos que los anteriores no definen.

### Semántica verificable del merge por campo

Esta es la semántica que hoy implementan `config_chain.go` (recorre la cadena `pyproject.toml → package.json → composer.json`) y `config_merge.go` (combina campo a campo) en cada tool. Se enuncia como invariantes verificables por test:

1. **Gana el primer archivo de la cadena que define cada campo.** El merge es *por campo*, no *por archivo*: si `pyproject.toml` define `maxLines` y `package.json` define `exclude`, el resultado toma `maxLines` del primero y `exclude` del segundo. "Definir" significa aportar un valor no-cero / no-vacío; un campo ausente o en su valor cero (`0`, `""`, slice vacío) se considera **no definido** y deja pasar al siguiente archivo de la cadena.

2. **Los arrays son atómicos.** Colecciones como `rules`, `exclude`, `patterns`, `allowlist` NO se concatenan ni se fusionan elemento a elemento: gana **entero** el array del primer archivo de la cadena que lo define (no-vacío). Un `package.json` con `exclude: ["vendor"]` no se suma a un `composer.json` con `exclude: ["node_modules"]`; se toma solo el primero de la cadena que traiga la lista.

3. **Al agotar la cadena, defaults compilados.** Tras recorrer los tres archivos, los campos que ninguno definió los rellena `applyConfigDefaults` con los defaults compilados en el binario. Esto garantiza que el `Config` resultante siempre esté completo, aunque no exista ningún archivo de config.

Consecuencias verificables de estas tres reglas:

- Dos archivos con campos disjuntos producen la **unión** de sus campos (regla 1).
- Dos archivos que definen el **mismo** campo escalar → gana el primero de la cadena; el segundo se ignora para ese campo (regla 1).
- Dos archivos que definen el **mismo** array → gana el array del primero entero, sin mezcla (regla 2).
- Cero archivos presentes → `Config` == defaults compilados (regla 3).

`<tool>.json` dedicado y los CLI flags se resuelven **fuera** de esta cadena de fallback (niveles 1 y 2 de la precedencia de arriba, con mayor prioridad); la semántica descrita aquí gobierna exclusivamente los tres archivos idiomáticos de ecosistema y el relleno final con defaults.

### Por qué un parser TOML mínimo en Go puro

Go stdlib no incluye `encoding/toml`. Tres opciones:

| Opción | Trade-off |
|---|---|
| Dep externa `github.com/BurntSushi/toml` | Rompe ADR-002 (cero deps). Requiere ADR de excepción. |
| Convertir TOML → JSON con tool externo en build time | No sirve para users finales que escriben TOML a mano. |
| **Parser TOML mínimo (subset) en Go puro** | Preserva ADR-002. ~150 líneas de código bien delimitado. |

Elegimos la tercera. El parser solo necesita reconocer el subset que aparece típicamente en `[tool.<name>]` de un `pyproject.toml`:

- Tablas top-level con dot notation: `[tool.linelens]`, `[tool.linelens.default]`.
- Arrays of tables: `[[tool.linelens.rules]]`.
- Asignaciones simples: `maxLines = 100`, `pattern = "**/*_test.go"`.
- Inline tables: `default = { maxLines = 100 }`.
- Arrays of strings: `exclude = ["node_modules", "vendor"]`.
- Arrays of inline tables: `rules = [{ pattern = "…", maxLines = 300 }]`.
- Comentarios `#` (línea entera o trailing).

NO soporta (cae en `error` claro al encontrarlos): datetimes, hex/oct/bin int literals, multi-line strings/arrays, dotted keys dentro de una tabla.

El parser vive en `tools/_shared/tomlmin/` con su propio `go.mod`. Cada tool lo importa via Go workspace `use`.

## API del parser

```go
package tomlmin

// ExtractAsJSON parses the TOML document and returns the JSON-encoded
// bytes of the table at sectionPath (dotted, e.g. "tool.linelens").
// Returns (nil, false, nil) if the section is absent; (nil, false, err)
// if the document contains syntax outside the supported subset.
func ExtractAsJSON(toml []byte, sectionPath string) ([]byte, bool, error)
```

Cada tool delega así:

```go
data, found, err := tomlmin.ExtractAsJSON(readFile(pyproject), "tool.linelens")
if err != nil  { return defaults, err }
if !found      { /* fall through to next source in chain */ }
var cfg Config
json.Unmarshal(data, &cfg)  // reusa el struct que ya tenemos
```

El `Config` struct **no cambia** en ningún tool. Eso evita drift entre fuentes y mantiene la semántica idéntica para los 4 archivos posibles.

## Bump semántico

`0.2.0` para los 4 tools. Es feature nueva sin breaking change: los proyectos existentes que solo usan `<tool>.json` o `package.json` siguen funcionando idénticamente.

## Consecuencias

**Positivas:**
- Python users pueden centralizar config en `pyproject.toml` siguiendo PEP 518 + convención de ruff/black.
- PHP users pueden centralizar config en `composer.json` siguiendo convención de phpstan.
- Un solo formato canónico de struct por tool — sin drift entre formatos.
- Cero deps externas, ADR-002 intacto.

**Negativas:**
- ~150 LOC nuevos por el parser TOML (más tests). Mitigado por estar aislado en `tools/_shared/tomlmin/` con 100% coverage propio.
- 2 fuentes nuevas por tool en `loadConfig` — la lógica de chain crece. Mitigado por extraer cada source a su propio `config_<source>.go` (≤ 50 LOC cada uno).

**Neutras:**
- La chain de 6 niveles podría confundir si alguien tiene los 4 archivos simultáneamente. Documentado explícitamente en cada README de paquete con tabla de precedencia.
