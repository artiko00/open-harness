package langsyntax

// family agrupa los lenguajes que comparten la forma de sus declaraciones de
// import. Se resuelve por extensión: el mismo prefijo significa cosas distintas
// según el lenguaje (`use` importa en Rust y PHP, `using` abre un recurso en C#).
type family int

const (
	famJS family = iota + 1
	famPy
	famGo
	famRb
	famRust
	famJVM
	famPHP
	famC
	famCS
	famDart
	famSwift
)

// extFamily mapea las extensiones de código a su familia de sintaxis. Cubre el
// mismo universo que pathmatch.CodeExtensions(); una extensión ausente deja el
// fuente intacto.
var extFamily = map[string]family{
	".ts": famJS, ".tsx": famJS, ".js": famJS, ".jsx": famJS,
	".mjs": famJS, ".cjs": famJS,
	".py": famPy, ".go": famGo, ".rb": famRb, ".rs": famRust,
	".java": famJVM, ".kt": famJVM, ".kts": famJVM, ".scala": famJVM,
	".php": famPHP,
	".c":   famC, ".cc": famC, ".cpp": famC, ".h": famC, ".hpp": famC,
	".m": famC, ".mm": famC,
	".cs": famCS, ".dart": famDart, ".swift": famSwift,
}

// importPrefixes: la línea abre una declaración de import si empieza con alguna
// de estas palabras (con frontera de palabra).
var importPrefixes = map[family][]string{
	famJS:    {"import"},
	famPy:    {"import"},
	famGo:    {"import", "package"},
	famRb:    {"require", "require_relative", "load"},
	famRust:  {"use", "extern"},
	famJVM:   {"import", "package"},
	famPHP:   {"use", "namespace", "require", "require_once", "include", "include_once"},
	famC:     {"#include", "#import"},
	famCS:    {"using"},
	famDart:  {"import", "export", "part", "library"},
	famSwift: {"import"},
}

// importPairs: la línea abre una declaración si empieza con la palabra [0] Y
// contiene la subcadena [1]. Distingue `export … from …` de un `export` de
// definición, y `from x import y` de cualquier otro uso de `from`.
var importPairs = map[family][][2]string{
	famJS: {
		{"export", " from "},
		{"const", "require("}, {"let", "require("}, {"var", "require("},
	},
	famPy: {{"from", " import"}},
	famC:  {{"using", "namespace"}},
	famCS: {{"global", "using"}},
}

// importNegPairs: la línea NO es un import aunque empiece con la palabra [0], si
// contiene la subcadena [1]. Excluye el `using (…)` de recurso de C#/C++ y el
// `import(…)` dinámico de JS, que son sentencias ejecutables.
var importNegPairs = map[family][][2]string{
	famJS: {{"import", "("}},
	famC:  {{"using", "("}},
	famCS: {{"using", "("}},
}
