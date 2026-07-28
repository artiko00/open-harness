package main

// providerPatterns cubre tokens con prefijo de proveedor y URIs con
// credenciales embebidas. Son señales fuertes: no se filtran por entropía.
func providerPatterns() []PatternRule {
	return []PatternRule{
		{
			Name:     "Stripe Live Secret Key",
			Pattern:  `sk_live_[0-9A-Za-z]{16,}`,
			Severity: "critical",
		},
		{
			Name:     "Slack Token",
			Pattern:  `xox[baprs]-[0-9A-Za-z-]{10,}`,
			Severity: "critical",
		},
		{
			Name:     "Slack Webhook URL",
			Pattern:  `hooks\.slack\.com/services/[A-Za-z0-9/]{10,}`,
			Severity: "high",
		},
		{
			Name:     "Google API Key",
			Pattern:  `AIza[0-9A-Za-z\-_]{35}`,
			Severity: "high",
		},
		{
			Name:     "OpenAI API Key",
			Pattern:  `sk-proj-[0-9A-Za-z\-_]{20,}`,
			Severity: "critical",
		},
		{
			Name:     "GitLab Personal Access Token",
			Pattern:  `glpat-[0-9A-Za-z\-_]{20}`,
			Severity: "critical",
		},
		{
			Name:     "npm Access Token",
			Pattern:  `npm_[0-9A-Za-z]{36}`,
			Severity: "critical",
		},
		{
			Name:     "SendGrid API Key",
			Pattern:  `SG\.[0-9A-Za-z_\-]{16,32}\.[0-9A-Za-z_\-]{16,64}`,
			Severity: "critical",
		},
		{
			Name:     "Credentials in Connection URI",
			Pattern:  `(?i)(?:postgres|postgresql|mysql|mongodb\+srv|mongodb|redis|amqp)://[^\s:@/]+:([^\s@/]+)@`,
			Severity: "high",
		},
	}
}
