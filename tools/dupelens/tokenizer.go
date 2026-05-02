package main

import (
	"strings"
	"unicode"
)

// Token representa una unidad léxica extraída del código fuente.
type Token struct {
	Value string
	Line  int
}

// tokenize parte el contenido en tokens simples.
// Estrategia: separar por espacios y puntuación, ignorar tokens triviales.
// Es language-agnostic — suficiente para detectar duplicados estructurales
// sin parsear gramática específica de cada lenguaje.
//
// Pre-procesa con stripCommentsAndStrings para descartar contenido literal
// y comentarios, que generan ruido (false positives) en duplicate detection.
func tokenize(src string) []Token {
	src = stripCommentsAndStrings(src)
	var tokens []Token
	line := 1
	var current strings.Builder

	flush := func() {
		t := strings.TrimSpace(current.String())
		if t != "" && !isTrivialToken(t) {
			tokens = append(tokens, Token{Value: t, Line: line})
		}
		current.Reset()
	}

	for _, r := range src {
		if r == '\n' {
			flush()
			line++
			continue
		}
		if unicode.IsSpace(r) || isPunctuation(r) {
			flush()
			continue
		}
		current.WriteRune(r)
	}
	flush()
	return tokens
}

// isPunctuation identifica caracteres que separan tokens pero no son
// significativos para detección de duplicados (paréntesis, comas, etc.).
func isPunctuation(r rune) bool {
	switch r {
	case '(', ')', '{', '}', '[', ']', ',', ';', ':', '.':
		return true
	}
	return false
}

// isTrivialToken descarta tokens muy cortos o ruidosos.
// Tokens de 1 caracter rara vez aportan a duplicación significativa.
func isTrivialToken(t string) bool {
	if len(t) <= 1 {
		return true
	}
	return false
}
