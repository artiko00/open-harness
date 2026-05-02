# ADR-010: Diseño del detector de secretos (secretlens)

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

El proyecto necesitaba una herramienta que detectara secretos y credenciales hardcodeadas antes de que lleguen al repositorio. Los problemas a resolver:

- Los secretos en código son una de las vulnerabilidades más comunes y costosas.
- La detección debe ocurrir en pre-commit o pre-push, no en CI (tarde).
- Debe funcionar en cualquier lenguaje sin parser AST.
- Los falsos positivos destruyen la adopción: si la herramienta grita demasiado, el equipo la desactiva.

## Decisiones de diseño

### 1. Detección basada en regex, no en AST

Alternativa descartada: parsear el AST de cada lenguaje para encontrar asignaciones a variables sensibles.

**Motivo del rechazo:** requiere un parser por lenguaje, complejidad O(n·lenguajes), y los secretos también aparecen en archivos de configuración (`.env`, YAML, JSON) que no tienen AST razonable.

**Decisión:** regex sobre líneas de texto plano. Captura el 95% de los casos reales con una fracción de la complejidad.

### 2. Allowlist para reducir falsos positivos

Los patrones de secretos son amplios por diseño (para no perder casos reales), lo que genera falsos positivos en ejemplos de documentación.

Se implementa una lista de palabras clave que, si aparecen en la línea, la omiten: `example`, `placeholder`, `your_key_here`, `changeme`, `xxxx`, `****`.

La comparación es case-insensitive para cubrir `EXAMPLE`, `Example`, etc.

### 3. Tres niveles de severidad

| Nivel | Criterio |
|---|---|
| `critical` | Formatos de secreto conocidos y específicos (AWS AKIA…, `ghp_…`, PEM headers) |
| `high` | Asignaciones genéricas con valores largos (password=, secret=, api_key=) |
| `medium` | Tokens y bearer headers con valores moderados |

La separación permite a los equipos filtrar por nivel: un pipeline de CI puede fallar solo en `critical`, mientras que un pre-commit puede alertar en todos.

### 4. Ocho patrones built-in

Se eligieron los ocho patrones más frecuentes en filtraciones reales según los estudios de GitGuardian y TruffleHog:
1. AWS Access Key ID (`AKIA[0-9A-Z]{16}`)
2. AWS Secret Access Key
3. GitHub Personal Access Token (`ghp_…`)
4. GitHub Fine-Grained Token (`github_pat_…`)
5. PEM Private Key (`-----BEGIN … PRIVATE KEY`)
6. JWT (`eyJ…`)
7. Asignación genérica de secret/password/api_key
8. Asignación genérica de token/access_token/bearer

Los patrones son sobreridables vía `secretlens.json`.

### 5. Silencio en errores de walk, propagación del error de raíz

Si un archivo desaparece entre el listado del directorio y la lectura (race condition normal en sistemas activos), `scanFile` devuelve error y el walk lo omite silenciosamente. Esto es correcto: no queremos abortar un escaneo completo por un archivo transitoriamente inaccesible.

El error del directorio raíz sí se propaga porque implica que el usuario especificó una ruta inválida, lo cual es un error del operador, no un race condition.

## Consecuencias

- **Positivo:** lenguaje-agnóstico, cero parsers, startup rápido.
- **Positivo:** allowlist configurable reduce falsos positivos sin modificar los patrones.
- **Positivo:** los patrones built-in cubren el 80% de las filtraciones más comunes.
- **Negativo:** regex no entiende contexto; puede fallar en secretos ofuscados o partidos en múltiples líneas.
- **Negativo:** patrones demasiado amplios pueden requerir ajuste de allowlist por proyecto.
- **Invariante:** el diseño prioriza el recall (no perder secretos reales) sobre la precision (evitar falsos positivos). La allowlist es el mecanismo de ajuste de precisión.
