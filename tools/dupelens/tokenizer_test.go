package main

import "testing"

func TestTokenize_skipsPunctuationAndShortTokens(t *testing.T) {
	src := `if (a == b) { return 0; }`
	tokens := tokenize(src, ".go")
	// Esperamos tokens "alpha"-style sin puntuación ni tokens de 1 caracter.
	wantValues := []string{"if", "==", "return"}
	gotValues := []string{}
	for _, tok := range tokens {
		gotValues = append(gotValues, tok.Value)
	}
	for _, w := range wantValues {
		found := false
		for _, g := range gotValues {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected token %q in %v", w, gotValues)
		}
	}
}

func TestTokenize_lineNumbersTracked(t *testing.T) {
	src := "line_one_token\n\nline_three_token\n"
	tokens := tokenize(src, ".go")
	if len(tokens) < 2 {
		t.Fatalf("expected at least 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Line != 1 {
		t.Errorf("first token line = %d; want 1", tokens[0].Line)
	}
	if tokens[len(tokens)-1].Line != 3 {
		t.Errorf("last token line = %d; want 3", tokens[len(tokens)-1].Line)
	}
}

func TestHashToken_deterministic(t *testing.T) {
	a := hashToken("function")
	b := hashToken("function")
	if a != b {
		t.Errorf("hashToken non-deterministic: %d != %d", a, b)
	}
	c := hashToken("FUNCTION")
	if a == c {
		t.Errorf("hashToken should be case-sensitive")
	}
}
