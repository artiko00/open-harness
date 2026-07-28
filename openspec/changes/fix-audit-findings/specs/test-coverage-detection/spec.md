# test-coverage-detection

Detección de archivos fuente sin test de testlens 0.2.5. Hoy el modo por defecto produce el 100% de
falsos positivos en Go, TypeScript, Python, Ruby, Rust y Dart; la detección de lenguaje itera un
`map` de Go y devuelve resultados distintos entre corridas idénticas; y un archivo de test vacío
marca cubierto un paquete entero.

## ADDED Requirements

### Requirement: Detección de lenguaje determinista

La inferencia de lenguaje SHALL recorrer los lenguajes soportados en un orden fijo y SHALL elegir el
lenguaje con mayor cantidad de archivos, no el primero que supere un umbral. Dos ejecuciones con el
mismo árbol y la misma configuración SHALL producir el mismo resultado.

#### Scenario: Resultado estable entre corridas

- **GIVEN** el repositorio open-harness en su raíz
- **WHEN** se ejecuta `testlens check --no-color` veinte veces consecutivas
- **THEN** las veinte salidas son idénticas

#### Scenario: Gana el lenguaje mayoritario

- **GIVEN** un monorepo con 8 archivos `.go` y 3 archivos `.ts`
- **WHEN** se ejecuta `testlens check --dir .`
- **THEN** el lenguaje detectado es Go

#### Scenario: Proyecto pequeño se detecta igual

- **GIVEN** un proyecto Go con solo 3 archivos `.go`
- **WHEN** se ejecuta `testlens check --dir .`
- **THEN** el lenguaje detectado es Go
- **AND** no se emite el aviso de que no se pudo detectar el lenguaje

### Requirement: El modo automático usa el mapeo real del lenguaje

Cuando el lenguaje se infiere, testlens SHALL usar el `languageMapping` completo de ese lenguaje,
incluyendo `testSuffixes`, `testDirs`, `mirrors` y `packageBased`. NO MUST construirse un mapeo
genérico sintético.

#### Scenario: Proyecto Go sin flags

- **GIVEN** un proyecto Go con `pkg/a.go`, `pkg/b.go` y `pkg/a_test.go`
- **WHEN** se ejecuta `testlens check --dir . --no-color`
- **THEN** el resultado es idéntico al de `testlens check --dir . --lang go --no-color`

#### Scenario: Paridad entre modo automático y explícito

- **GIVEN** fixtures de Go, TypeScript, Python, Ruby, Rust y Dart, todos con sus tests presentes
- **WHEN** se ejecuta `testlens check --dir .` sin `--lang` sobre cada fixture
- **THEN** ninguno reporta archivos sin test

### Requirement: Un archivo de test cuenta solo si contiene tests

Un archivo candidato SHALL considerarse cobertura únicamente si contiene al menos una función de
test según la convención del lenguaje (`func Test`, `it(`, `test(`, `def test_`, `#[test]`).
Un archivo de test vacío NO MUST marcar como cubierto ningún archivo fuente ni ningún paquete.

#### Scenario: Archivo de test vacío no cubre el paquete

- **GIVEN** un paquete Go con seis archivos fuente y un `zzz_test.go` cuyo contenido es solo `package svc`
- **WHEN** se ejecuta `testlens check --dir . --lang go --fail`
- **THEN** los seis archivos se reportan como sin test
- **AND** el exit code es 1

#### Scenario: Archivo de test con contenido sí cubre

- **GIVEN** el mismo paquete con un `svc_test.go` que declara `func TestF1(t *testing.T)`
- **WHEN** se ejecuta `testlens check --dir . --lang go`
- **THEN** el paquete se considera cubierto

### Requirement: Los archivos de test no se reportan como fuente

Un archivo de test NO MUST aparecer en el reporte como archivo fuente sin test. La identificación
SHALL contemplar tanto sufijos como prefijos de test, incluido el prefijo `test_` de pytest.

#### Scenario: pytest no se exige tests a sí mismo

- **GIVEN** un proyecto Python con `app/calc.py` y `tests/test_calc.py`
- **WHEN** se ejecuta `testlens check --dir . --lang python`
- **THEN** `tests/test_calc.py` no aparece en el reporte

#### Scenario: El layout app/ + tests/ se resuelve

- **GIVEN** el mismo proyecto Python
- **WHEN** se ejecuta `testlens check --dir . --lang python`
- **THEN** `app/calc.py` se considera cubierto por `tests/test_calc.py`

### Requirement: Exclusiones por defecto de archivos que no llevan test

testlens SHALL excluir por defecto los archivos que por convención no llevan test: `__init__.py`,
`conftest.py`, `*_pb2.py`, `*.pb.go`, `*_gen.go`, `*.g.dart`, migraciones y archivos de settings.
La lista SHALL ser configurable.

#### Scenario: Generados y config no se reportan

- **GIVEN** un proyecto Python con `app/__init__.py`, `app/settings.py`, `app/models_pb2.py` y `app/migrations/0001_initial.py`
- **WHEN** se ejecuta `testlens check --dir . --lang python`
- **THEN** ninguno de los cuatro se reporta

#### Scenario: Un archivo de lógica sí se exige

- **GIVEN** el mismo proyecto con `app/calc.py` sin test
- **WHEN** se ejecuta `testlens check --dir . --lang python`
- **THEN** `app/calc.py` se reporta como sin test

### Requirement: Código de producción sin usar

testlens NO MUST mantener rutas de reporte y escaneo que el flujo real no ejecuta. Las funciones
alcanzables únicamente desde los tests SHALL eliminarse o incorporarse al flujo real.

#### Scenario: Sin implementaciones paralelas de reporte

- **WHEN** se inspecciona el paquete testlens
- **THEN** existe una única implementación del reporte
- **AND** la cobertura de statements se mantiene en 100% sin código inalcanzable desde `check`
