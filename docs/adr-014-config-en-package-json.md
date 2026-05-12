# ADR-014: Soporte de configuración desde `package.json`

**Fecha:** 2026-05-12
**Estado:** Aceptada
**Aplica a:** linelens, dupelens, secretlens, testlens

## Contexto

Cada herramienta del monorepo lee su configuración desde un archivo dedicado en la raíz del proyecto consumidor (`linelens.json`, `dupelens.json`, `secretlens.json`, `testlens.json`). Para proyectos Node/TypeScript que ya consumen los tools vía npm, mantener un archivo JSON extra por linter añade ruido en la raíz del repo.

El ecosistema npm tiene una convención bien establecida: las herramientas permiten configuración inline en `package.json` bajo una key dedicada (`prettier`, `eslintConfig`, `stylelint`, `husky`, etc.). Adoptarla simplifica la integración y reduce la cantidad de archivos top-level.

## Decisión

Cada herramienta soporta una **fuente secundaria** de configuración: una key dedicada en `package.json` con el mismo nombre del binario (`linelens`, `dupelens`, `secretlens`, `testlens`).

### Precedencia

1. **CLI flags** (`--max`, `--min-tokens`, etc.) — siempre ganan.
2. **`--config <path>` explícito o archivo dedicado** (`linelens.json`, etc.) — gana sobre `package.json`.
3. **`package.json` con la key correspondiente** — fallback.
4. **Defaults compilados en el binario** — última instancia.

### Ámbito de búsqueda

Solo se busca `package.json` en el **mismo directorio** donde estaría el archivo dedicado (en la práctica: el cwd del comando, o `filepath.Dir(--config)`). **No** hay búsqueda hacia arriba en el árbol de directorios.

### Forma de la key

Key directa por tool, no anidada bajo un namespace:

```json
{
  "name": "my-project",
  "linelens": { "default": { "maxLines": 100 } },
  "dupelens": { "default": { "minTokens": 50 } }
}
```

Justificación: es el patrón estándar (prettier, eslintConfig, stylelint). Más fácil de descubrir y de buscar en docs externas. El "costo" de 4 keys top-level es marginal frente a los 4 archivos JSON que reemplazan.

## Consecuencias

**Positivas:**
- Proyectos Node/TypeScript pueden eliminar los archivos JSON dedicados y centralizar en `package.json`.
- Más alineado con la convención de la comunidad npm.
- Cero deps externas (solo `encoding/json` de stdlib, mismo costo que el parser actual).

**Negativas:**
- Doble fuente de verdad: si un proyecto define ambos, la regla de precedencia debe quedar muy clara (se documenta en el README de cada tool).
- Aumenta levemente la complejidad de `loadConfig` (+1 función auxiliar, +1 archivo `config_pkg.go` para respetar el límite de 100 líneas).

**Neutras:**
- Búsqueda hacia arriba en el árbol queda fuera del scope. Si en el futuro aparece un caso real de monorepo donde se justifique, se evaluará en un ADR aparte.

## Implementación

Cada tool implementa el fallback en un archivo `config_pkg.go` separado del `config.go` original. La función `loadConfigFromPackageJSON(dir string) (Config, bool, error)`:

1. Lee `<dir>/package.json`.
2. Parsea como `map[string]json.RawMessage`.
3. Extrae la sección bajo la key del tool.
4. Si existe → unmarshalea a `Config`, aplica defaults, retorna `(cfg, true, nil)`.
5. Si no existe → retorna `(defaults, false, nil)`.

`loadConfig(path)` invoca este helper solo cuando el archivo dedicado **no existe**. Si el archivo dedicado está malformado, el error se propaga sin caer al fallback.

Tests cubren los 4 escenarios principales por tool (archivo dedicado presente, ausente, `package.json` con/sin key, key malformada) más casos de error de I/O, manteniendo 100% statement coverage ([ADR-011](adr-011-cobertura-100-como-estandar.md)).
