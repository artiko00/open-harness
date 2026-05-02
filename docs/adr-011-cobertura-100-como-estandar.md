# ADR-011: Cobertura de tests 100% como estándar del proyecto

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

El proyecto alcanzó cobertura de ~40% en su primera implementación funcional. Antes de considerar las herramientas "terminadas" se estableció el objetivo de 100% de cobertura de statements en todos los tools del monorepo.

La pregunta era: ¿cómo hacer testeable código que llama `os.Exit`, usa `flag.ExitOnError`, y tiene ramas de error prácticamente inalcanzables?

## Decisiones de diseño

### 1. Inyección de `osExit` para testear `main()`

El patrón estándar para hacer testeable una función que llama `os.Exit`:

```go
var osExit = os.Exit

func main() {
    osExit(run(os.Args[1:]))
}
```

En tests, se sobreescribe `osExit` con un closure que captura el código sin salir del proceso. Permite verificar que `main()` llama a `run()` y propaga el código correcto.

### 2. `run()` separada del punto de entrada

`main()` delega inmediatamente a `run(args []string) int`. Esto permite invocar toda la lógica de despacho desde tests sin manipular `os.Args` ni mockear la señal de salida.

La función `main()` queda de 2 líneas; toda la lógica comprobable vive en `run()` y las funciones que llama.

### 3. `flag.ContinueOnError` en lugar de `flag.ExitOnError`

Los subcomandos (`runCheck`, `runInit`) usan `flag.NewFlagSet("...", flag.ContinueOnError)` y comprueban el error devuelto por `fs.Parse()`. Esto permite testear el comportamiento ante flags inválidos sin que el proceso del test muera.

### 4. Propagación del error de directorio raíz

`scan()` en ambas herramientas modificó su callback de `filepath.WalkDir` para retornar el error cuando `path == root`. Sin esto, escanear un directorio inexistente devuelve resultados vacíos sin error, haciendo la rama de error en `runCheck` inalcanzable.

### 5. Eliminación de la rama `io.EOF` en `isBinaryContent`

La función leía 512 bytes y comprobaba `if err != nil && err != io.EOF`. En la práctica, `f.Read` en un archivo regular nunca devuelve un error distinto de EOF, haciendo esa rama inalcanzable. Se simplificó a `n, _ := f.Read(buf)`, eliminando el dead code y la rama sin test.

### 6. Cobertura de ramas OS mediante permisos de archivo

Para cubrir ramas de error que requieren fallo de I/O:
- `chmod 0000` en un subdirectorio activa el callback de WalkDir con error en `path != root`.
- `chmod 0000` en un archivo regular activa el error de `countLines` / `scanFile`.
- Pasar una ruta de directorio como config file activa el error de `os.ReadFile` que no es `IsNotExist`.

Estos tests hacen `defer os.Chmod(path, 0755/0644)` para restaurar permisos incluso si fallan.

## Consecuencias

- **Positivo:** cualquier regresión en las rutas de error se detecta inmediatamente.
- **Positivo:** la arquitectura resultante (`run()` separada, `osExit` inyectable) es más limpia que la original con `os.Exit` embebido.
- **Positivo:** el patrón es replicable: cada nuevo tool del monorepo tiene un template claro.
- **Negativo:** los tests que usan `chmod 0000` no funcionan si el proceso corre como root (CI en algunos entornos). Se acepta porque los tests no fallan en ese caso, simplemente no ejercen esa rama.
- **Invariante:** la cobertura se mide con `go test -coverprofile` + `go tool cover -func`. El 100% es de statements, no de branches (Go no reporta branch coverage nativo). Algunas condiciones booleanas compuestas pueden quedar parcialmente cubiertas sin que la métrica lo refleje.
