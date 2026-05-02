# ADR-004: Detección de archivos binarios por extensión y bytes nulos

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

Al escanear un proyecto real aparecen archivos que no son código fuente: imágenes, fuentes, PDFs, binarios compilados. Contar las "líneas" de un JPG o de un ejecutable no tiene sentido y genera falsos positivos ruidosos.

Se necesita una estrategia para identificar y omitir estos archivos.

## Decisión

Doble verificación en `binary.go`:

1. **Por extensión** (`isBinaryPath`): lista de extensiones conocidas (`.jpg`, `.png`, `.exe`, `.wasm`, etc.). Es O(1) y no requiere abrir el archivo.
2. **Por contenido** (`isBinaryContent`): lee los primeros 512 bytes y detecta bytes nulos (`\x00`). Cubre binarios con extensiones no listadas (ej: binario compilado sin extensión).

El orden importa: la verificación por extensión va primero porque es más barata. La verificación por contenido solo corre si la extensión no es reconocida.

## Consecuencias

- **Positivo:** cero falsos positivos por archivos binarios en proyectos reales.
- **Positivo:** cubre el caso del binario compilado `linelens` (sin extensión), que se detecta por bytes nulos.
- **Negativo:** la lista de extensiones requiere mantenimiento manual si surgen nuevos formatos.
- **Negativo:** leer 512 bytes de cada archivo no-binario-por-extensión agrega un syscall extra. Impacto negligible en proyectos típicos (<10,000 archivos).
- **Alternativa descartada:** excluir archivos sin extensión — demasiado agresivo, omitía `Makefile`, `Dockerfile`, etc.
