package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func compilePatterns(rules []PatternRule) ([]compiledRule, error) {
	compiled := make([]compiledRule, 0, len(rules))
	for _, r := range rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", r.Pattern, err)
		}
		compiled = append(compiled, compiledRule{r, re})
	}
	return compiled, nil
}

func scanFile(path, relPath string, compiled []compiledRule, allowlist []string) ([]Finding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var findings []Finding
	lineNum := 0
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 1024*1024), 1024*1024)

	for s.Scan() {
		lineNum++
		line := s.Text()

		if isAllowed(line, allowlist) {
			continue
		}

		for _, c := range compiled {
			if c.re.MatchString(line) {
				findings = append(findings, Finding{
					RelPath:  relPath,
					Line:     lineNum,
					Content:  truncate(strings.TrimSpace(line), 80),
					RuleName: c.rule.Name,
					Severity: c.rule.Severity,
				})
			}
		}
	}

	return findings, s.Err()
}

func isAllowed(line string, allowlist []string) bool {
	lower := strings.ToLower(line)
	for _, a := range allowlist {
		if strings.Contains(lower, strings.ToLower(a)) {
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
