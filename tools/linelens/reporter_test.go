package main

import "testing"

func TestReport_Empty(t *testing.T) {
	if count := report([]FileResult{}, true); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestReport_NoViolationsNoColor(t *testing.T) {
	results := []FileResult{
		{RelPath: "a.go", Lines: 50, MaxLines: 100},
		{RelPath: "b.go", Lines: 80, MaxLines: 100},
	}
	if count := report(results, true); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestReport_NoViolationsColor(t *testing.T) {
	results := []FileResult{
		{RelPath: "a.go", Lines: 50, MaxLines: 100},
	}
	if count := report(results, false); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestReport_ViolationsNoColor(t *testing.T) {
	results := []FileResult{
		{RelPath: "big.go", Lines: 150, MaxLines: 100},
		{RelPath: "bigger.go", Lines: 200, MaxLines: 100},
	}
	if count := report(results, true); count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestReport_ViolationsColorYellow(t *testing.T) {
	// excess=30, maxLines/2=50, 30 <= 50 → yellow
	results := []FileResult{{RelPath: "big.go", Lines: 130, MaxLines: 100}}
	if count := report(results, false); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestReport_ViolationsColorRed(t *testing.T) {
	// excess=60, maxLines/2=50, 60 > 50 → red
	results := []FileResult{{RelPath: "big.go", Lines: 160, MaxLines: 100}}
	if count := report(results, false); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestReport_LongPathNoColor(t *testing.T) {
	longPath := "src/very/deeply/nested/path/to/some/file/that/exceeds/sixty/chars.go"
	results := []FileResult{{RelPath: longPath, Lines: 150, MaxLines: 100}}
	if count := report(results, true); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestReport_LongPathColor(t *testing.T) {
	longPath := "src/very/deeply/nested/path/to/some/file/that/exceeds/sixty/chars.go"
	results := []FileResult{{RelPath: longPath, Lines: 160, MaxLines: 100}}
	if count := report(results, false); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}
