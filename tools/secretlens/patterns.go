package main

// defaultPatterns returns the built-in secret detection rules.
// Patterns are intentionally broad enough to catch real secrets but
// specific enough to avoid flooding with false positives.
func defaultPatterns() []PatternRule {
	return []PatternRule{
		{
			Name:     "AWS Access Key ID",
			Pattern:  `AKIA[0-9A-Z]{16}`,
			Severity: "critical",
		},
		{
			Name:     "AWS Secret Access Key",
			Pattern:  `(?i)(aws_secret_access_key|aws_secret)\s*[:=]\s*["']?[A-Za-z0-9/+]{40}["']?`,
			Severity: "critical",
		},
		{
			Name:     "GitHub Personal Access Token",
			Pattern:  `ghp_[A-Za-z0-9]{36}`,
			Severity: "critical",
		},
		{
			Name:     "GitHub Fine-Grained Token",
			Pattern:  `github_pat_[A-Za-z0-9_]{82}`,
			Severity: "critical",
		},
		{
			Name:     "PEM Private Key",
			Pattern:  `-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY`,
			Severity: "critical",
		},
		{
			Name:     "JWT Token",
			Pattern:  `eyJ[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+`,
			Severity: "high",
		},
		{
			Name:     "Generic Secret Assignment",
			Pattern:  `(?i)(secret|password|passwd|api_key|apikey|auth_token|private_key)\s*[:=]\s*["'][A-Za-z0-9+/\-_!@#$%^&*]{12,}["']`,
			Severity: "high",
		},
		{
			Name:     "Generic Token Assignment",
			Pattern:  `(?i)(token|access_token|bearer)\s*[:=]\s*["'][A-Za-z0-9+/\-_]{20,}["']`,
			Severity: "medium",
		},
	}
}
