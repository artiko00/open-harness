package main

// keywords es la unión mínima de palabras reservadas de los lenguajes que
// dupelens analiza. Al normalizar identificadores para detectar clones Type-2,
// las keywords se preservan (son la columna vertebral estructural del código);
// solo los nombres definidos por el usuario colapsan a "ID". Una unión simple
// es suficiente y estable para esta heurística cross-language.
var keywords = map[string]bool{
	"func": true, "return": true, "if": true, "else": true, "for": true,
	"while": true, "var": true, "let": true, "const": true, "class": true,
	"def": true, "import": true, "from": true, "package": true, "type": true,
	"struct": true, "interface": true, "fn": true, "pub": true, "impl": true,
	"match": true, "switch": true, "case": true, "break": true, "continue": true,
	"new": true, "this": true, "self": true, "public": true, "private": true,
	"static": true, "void": true, "true": true, "false": true, "null": true,
	"nil": true, "none": true, "and": true, "or": true, "not": true,
	"in": true, "is": true, "do": true, "end": true, "then": true,
	"elif": true, "try": true, "catch": true, "except": true, "finally": true,
	"throw": true, "raise": true, "with": true, "as": true, "async": true,
	"await": true, "yield": true, "defer": true, "go": true, "chan": true,
	"map": true, "range": true, "select": true, "enum": true, "trait": true,
}

// isKeyword indica si el token es una palabra reservada conocida.
func isKeyword(v string) bool {
	return keywords[v]
}
