package main

import "testing"

func TestRuleForFile(t *testing.T) {
	rules := []Rule{
		{Pattern: "**/*.spec.*", MaxLines: 300},
		{Pattern: "**/migrations/**", Skip: true},
	}

	r := ruleForFile("src/auth.spec.ts", rules)
	if r == nil || r.MaxLines != 300 {
		t.Errorf("expected spec rule, got %v", r)
	}

	r = ruleForFile("db/migrations/001.sql", rules)
	if r == nil || !r.Skip {
		t.Errorf("expected skip rule, got %v", r)
	}

	r = ruleForFile("src/auth.ts", rules)
	if r != nil {
		t.Errorf("expected nil rule, got %v", r)
	}
}
