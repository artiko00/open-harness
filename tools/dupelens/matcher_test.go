package main

import "testing"

func TestRuleForFile_firstMatchWins(t *testing.T) {
	rules := []Rule{
		{Pattern: "**/*_test.go", Skip: true},
		{Pattern: "**/*.go", MinTokens: 30},
	}
	r := ruleForFile("foo/bar_test.go", rules)
	if r == nil || !r.Skip {
		t.Errorf("expected first rule (skip) to match")
	}
	r = ruleForFile("foo/bar.go", rules)
	if r == nil || r.MinTokens != 30 {
		t.Errorf("expected second rule (minTokens=30) to match")
	}
}
