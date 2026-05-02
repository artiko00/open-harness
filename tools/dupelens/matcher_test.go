package main

import "testing"

func TestMatchGlob_basicPatterns(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"*.go", "main.go", true},
		{"*.go", "tools/dupelens/main.go", true},
		{"**/*_test.go", "a/b/c_test.go", true},
		{"node_modules", "node_modules", true},
		{"node_modules", "src/node_modules", true},
		{"**/migrations/**", "db/migrations/001.sql", true},
		{"*.go", "main.ts", false},
	}
	for _, c := range cases {
		got := matchGlob(c.pattern, c.path)
		if got != c.want {
			t.Errorf("matchGlob(%q, %q) = %v; want %v", c.pattern, c.path, got, c.want)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	excludes := []string{"node_modules", "vendor", ".git"}
	if !isExcluded("node_modules/foo/bar.js", excludes) {
		t.Errorf("expected node_modules path to be excluded")
	}
	if !isExcluded("src/vendor/lib.go", excludes) {
		t.Errorf("expected vendor segment to be excluded")
	}
	if isExcluded("src/main.go", excludes) {
		t.Errorf("expected normal path NOT to be excluded")
	}
}

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
