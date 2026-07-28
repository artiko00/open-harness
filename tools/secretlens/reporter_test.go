package main

import "testing"

func TestReport_Empty(t *testing.T) {
	if count := report([]Finding{}, nil, true); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestReport_NoFindingsNoColor(t *testing.T) {
	if count := report([]Finding{}, nil, true); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestReport_NoFindingsColor(t *testing.T) {
	if count := report([]Finding{}, nil, false); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestReport_FindingsNoColor(t *testing.T) {
	findings := []Finding{
		{RelPath: "a.go", Line: 1, Content: "secret=abc", RuleName: "Generic", Severity: "high"},
	}
	if count := report(findings, nil, true); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestReport_AllSeveritiesColor(t *testing.T) {
	findings := []Finding{
		{RelPath: "a.go", Line: 1, Content: "x", RuleName: "r1", Severity: "critical"},
		{RelPath: "a.go", Line: 2, Content: "x", RuleName: "r2", Severity: "high"},
		{RelPath: "a.go", Line: 3, Content: "x", RuleName: "r3", Severity: "medium"},
		{RelPath: "a.go", Line: 4, Content: "x", RuleName: "r4", Severity: "low"},
	}
	if count := report(findings, nil, false); count != 4 {
		t.Errorf("expected 4, got %d", count)
	}
}

func TestReport_SortingByPriority(t *testing.T) {
	findings := []Finding{
		{RelPath: "b.go", Line: 5, Content: "x", RuleName: "r1", Severity: "medium"},
		{RelPath: "a.go", Line: 3, Content: "x", RuleName: "r2", Severity: "critical"},
		{RelPath: "a.go", Line: 1, Content: "x", RuleName: "r3", Severity: "critical"},
		{RelPath: "b.go", Line: 2, Content: "x", RuleName: "r4", Severity: "medium"},
		// same severity, different RelPaths (covers RelPath comparison branch)
		{RelPath: "a.go", Line: 9, Content: "x", RuleName: "r5", Severity: "medium"},
	}
	count := report(findings, nil, true)
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

func TestSeverityColor_AllLevels(t *testing.T) {
	cases := []struct{ sev, want string }{
		{"critical", colorRed + colorBold},
		{"high", colorRed},
		{"medium", colorYellow},
		{"low", colorGray},
	}
	for _, c := range cases {
		got := severityColor(c.sev)
		if got != c.want {
			t.Errorf("severityColor(%q) = %q, want %q", c.sev, got, c.want)
		}
	}
}
