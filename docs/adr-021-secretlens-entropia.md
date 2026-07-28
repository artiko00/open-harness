# ADR-021: Entropía de Shannon como filtro en secretlens

**Estado:** Aceptado
**Fecha:** 2026-07-27
**Aplica a:** secretlens
**Extiende:** [ADR-010](adr-010-secretlens-diseno-detector.md)

## Contexto

[ADR-010](adr-010-secretlens-diseno-detector.md) fijó el diseño de secretlens: detección por regex, priorizando recall (no perder secretos reales) sobre precision. Ese sesgo es correcto para reglas de **prefijo fuerte** (AWS `AKIA…`, `ghp_…`, PEM headers): el propio prefijo ya identifica el formato con certeza casi total, así que capturar es sinónimo de acertar.

El problema aparece en las **reglas genéricas** `KEY=VALUE` (`password=…`, `api_key=…`, `token=…`). Su regex debe ser amplio por diseño, y eso genera falsos positivos sobre valores que sintácticamente parecen una asignación de secreto pero no lo son:

- Placeholders y ejemplos que la allowlist no alcanza (`api_key=changeme123`, `password=aaaaaaaa`).
- Valores estructurados de baja aleatoriedad: rutas, nombres de variable, flags booleanos, enums (`token=enabled`).
- Repeticiones y secuencias (`secret=00000000`, `key=abcabcabc`).

Un secreto real, en cambio, tiende a ser una cadena de **alta aleatoriedad** por construcción. Esa diferencia es medible: la entropía de Shannon del valor capturado.

## Decisión

Se añade la **entropía de Shannon como FILTRO sobre el valor capturado**, no como detector.

Distinción clave:

- **No** es un detector: secretlens nunca reporta una línea solo por tener alta entropía. La detección la siguen haciendo los patrones regex de [ADR-010](adr-010-secretlens-diseno-detector.md). La entropía únicamente puede **descartar** un match que un patrón ya produjo.
- El filtro se calcula sobre el **valor capturado** (el grupo `VALUE` de la asignación), no sobre la línea entera. Medir la línea completa contaminaría la señal con el nombre de la clave y la sintaxis circundante.

### Umbral

`defaultMinEntropy = 3.0` bits por carácter. Un valor uniforme (`"00000000"`) da 0; un secreto aleatorio se acerca a `log2(alfabeto)` (≈ 4.3 para hex, ≈ 6 para base64). El umbral 3.0 deja pasar secretos reales de longitud razonable y corta la mayoría de placeholders/valores estructurados de baja aleatoriedad.

Es configurable vía `secretlens.json` (`minEntropy`); `0` en config se rellena con el default compilado.

### Aplicación selectiva: `entropyGate`

El filtro se aplica **solo a las reglas genéricas**, gobernado por un campo por patrón:

```go
EntropyGate bool `json:"entropyGate"`
```

- Reglas genéricas `KEY=VALUE`: `entropyGate = true` → el valor capturado debe superar `minEntropy` para reportarse.
- Reglas de prefijo fuerte (AWS, GitHub, PEM, JWT): `entropyGate = false` → se reportan siempre, sin gate. El prefijo ya es prueba suficiente; someterlas a un umbral de entropía solo introduciría el riesgo de perder un secreto real (viola el invariante de recall de [ADR-010](adr-010-secretlens-diseno-detector.md)).

La condición efectiva en el motor de match es:

```go
if c.rule.EntropyGate && shannonEntropy(value) < cfg.MinEntropy {
    // descartar: parece asignación de secreto pero el valor es de baja entropía
}
```

## Consecuencias

**Positivas:**
- Menos falsos positivos en las reglas genéricas sin tocar sus regex ni la allowlist.
- El recall de las reglas de prefijo fuerte queda intacto: nunca pasan por el gate.
- Umbral configurable por proyecto: un repo con secretos de test cortos puede bajarlo, uno estricto subirlo.

**Negativas:**
- Un secreto genérico real de muy baja entropía (p. ej. una contraseña débil `password=aaaaaaaa`) puede quedar por debajo del umbral y no reportarse. Es un trade-off consciente: esos valores también son indistinguibles de un placeholder, y la señal correcta ahí es una política de contraseñas, no un detector de secretos.
- El umbral 3.0 es una heurística global; proyectos con convenciones inusuales pueden necesitar ajustarlo.

**Neutras:**
- El cálculo es O(n) sobre el valor capturado (típicamente decenas de caracteres), despreciable frente al escaneo.
- Extiende [ADR-010](adr-010-secretlens-diseno-detector.md) sin revertirlo: el sesgo hacia recall sigue vigente donde importa (prefijos fuertes); el gate solo refina la precision donde el patrón es inherentemente ambiguo.
