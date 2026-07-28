package main

import "github.com/artiko00/open-harness/tools/_shared/pathmatch"

// ruleForFile retorna la regla aplicable al path (primera coincidencia).
func ruleForFile(relPath string, rules []Rule) *Rule {
	for i := range rules {
		if pathmatch.MatchGlob(rules[i].Pattern, relPath) {
			return &rules[i]
		}
	}
	return nil
}
