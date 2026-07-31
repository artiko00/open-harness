package main

import "testing"

func tokenValuesExt(src, ext string) []string {
	tokens := tokenize(src, ext, false)
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, t.Value)
	}
	return out
}

func tokenValues(src string) []string {
	return tokenValuesExt(src, ".go")
}

func contains(slice []string, v string) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func TestTokenize_stripsDoubleQuoteStringContents(t *testing.T) {
	src := `prefix "secret_inside_string" suffix`
	got := tokenValues(src)
	if contains(got, "secret_inside_string") {
		t.Errorf("contenido de string entre comillas dobles leaked: %v", got)
	}
	if !contains(got, "prefix") || !contains(got, "suffix") {
		t.Errorf("expected prefix/suffix preserved, got %v", got)
	}
}

func TestTokenize_stripsSingleQuoteStringContents(t *testing.T) {
	src := `before 'inner_text_a' after`
	got := tokenValues(src)
	if contains(got, "inner_text_a") {
		t.Errorf("contenido entre comillas simples leaked: %v", got)
	}
}

func TestTokenize_stripsBacktickStringContents(t *testing.T) {
	src := "raw `template_inner` end"
	got := tokenValues(src)
	if contains(got, "template_inner") {
		t.Errorf("contenido entre backticks leaked: %v", got)
	}
	if !contains(got, "raw") || !contains(got, "end") {
		t.Errorf("expected raw/end preserved, got %v", got)
	}
}

func TestTokenize_handlesEscapedQuotesInString(t *testing.T) {
	src := `before "with \"quoted\" middle" leakedKeyword after`
	got := tokenValues(src)
	if contains(got, "with") || contains(got, "quoted") || contains(got, "middle") {
		t.Errorf("escaped quotes mal manejados: %v", got)
	}
	if !contains(got, "leakedKeyword") || !contains(got, "after") {
		t.Errorf("tokens fuera de string deben preservarse: %v", got)
	}
}

func TestTokenize_stripsSingleLineCommentSlashSlash(t *testing.T) {
	src := "before // ignoredComment garbage\nafter"
	got := tokenValues(src)
	if contains(got, "ignoredComment") || contains(got, "garbage") {
		t.Errorf("contenido de // leaked: %v", got)
	}
	if !contains(got, "before") || !contains(got, "after") {
		t.Errorf("contexto pre/post comentario debe preservarse: %v", got)
	}
}

func TestTokenize_stripsHashCommentForPython(t *testing.T) {
	src := "code # python_comment_here\ndef foo"
	got := tokenValuesExt(src, ".py")
	if contains(got, "python_comment_here") {
		t.Errorf("contenido de # leaked: %v", got)
	}
	if !contains(got, "code") || !contains(got, "def") || !contains(got, "foo") {
		t.Errorf("contexto debe preservarse: %v", got)
	}
}

func TestTokenize_hashNotCommentForGo(t *testing.T) {
	// En Go '#' no inicia comentario: el token posterior no se descarta.
	src := "value #keepToken"
	got := tokenValuesExt(src, ".go")
	if !contains(got, "#keepToken") {
		t.Errorf("'#' en .go no debe iniciar comentario: %v", got)
	}
}

func TestTokenize_stripsMultiLineBlockComment(t *testing.T) {
	src := "head /* this is\n  multiline_inside\n  more_garbage */ tail"
	got := tokenValues(src)
	if contains(got, "multiline_inside") || contains(got, "more_garbage") {
		t.Errorf("contenido de /* */ leaked: %v", got)
	}
	if !contains(got, "head") || !contains(got, "tail") {
		t.Errorf("contexto pre/post bloque debe preservarse: %v", got)
	}
}

func TestTokenize_blockCommentPreservesLineNumbers(t *testing.T) {
	src := "head /* a\nb\nc */ after"
	tokens := tokenize(src, ".go", false)
	var headLine, afterLine int
	for _, tok := range tokens {
		if tok.Value == "head" {
			headLine = tok.Line
		}
		if tok.Value == "after" {
			afterLine = tok.Line
		}
	}
	if headLine != 1 {
		t.Errorf("head line = %d; want 1", headLine)
	}
	if afterLine != 3 {
		t.Errorf("after line = %d; want 3 (tras 2 saltos en bloque)", afterLine)
	}
}

func TestTokenize_emptyFileProducesNoTokens(t *testing.T) {
	if got := tokenize("", ".go", false); len(got) != 0 {
		t.Errorf("empty input must produce 0 tokens, got %d", len(got))
	}
}

func TestTokenize_onlyCommentsFile_producesNoTokens(t *testing.T) {
	src := "// solo comentarios\n# otro\n/* y bloque\n a la vez */"
	if got := tokenize(src, ".py", false); len(got) != 0 {
		t.Errorf("file con solo comments no debe producir tokens, got %d: %v", len(got), got)
	}
}

func TestTokenize_singleLineFileStillWorks(t *testing.T) {
	src := "alpha beta gamma"
	got := tokenValues(src)
	for _, want := range []string{"alpha", "beta", "gamma"} {
		if !contains(got, want) {
			t.Errorf("expected %q in tokens, got %v", want, got)
		}
	}
}

func TestTokenize_stringWithCommentSyntaxInsideNotRemoved_butStringContentStripped(t *testing.T) {
	src := `outer "// not a comment, just text" tail`
	got := tokenValues(src)
	if contains(got, "not") || contains(got, "comment") {
		t.Errorf("contenido de string que parece comentario leaked: %v", got)
	}
	if !contains(got, "outer") || !contains(got, "tail") {
		t.Errorf("outer/tail debe preservarse: %v", got)
	}
}
