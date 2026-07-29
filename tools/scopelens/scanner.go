package main

import (
	"context"
	"sort"
	"strings"
)

// scanResult reúne lo medido: rama, base efectiva, merge-base abreviado y el
// conjunto de archivos tocados (ordenado, sin duplicados).
type scanResult struct {
	Branch    string
	Base      string
	MergeBase string
	Files     []string
	Churn     map[string]lineStat
}

// measure calcula la unión del diff acumulado (merge-base...HEAD) con el índice.
// Todo error operativo se propaga para mapearse a exit 2; nunca degrada a 0.
func measure(run gitRunner, flagBase, cfgBase string, stagedOnly bool) (scanResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()

	var res scanResult
	if err := ensureMeasurable(ctx, run); err != nil {
		return res, err
	}
	res.Branch = branchName(ctx, run)

	set := map[string]struct{}{}
	churn := map[string]lineStat{}
	staged, err := diffFiles(ctx, run, "--cached")
	if err != nil {
		return res, err
	}
	addAll(set, staged)
	if err := accumulateChurn(ctx, run, "--cached", churn); err != nil {
		return res, err
	}

	if !stagedOnly && headExists(ctx, run) {
		if err := accumulateBranch(ctx, run, flagBase, cfgBase, &res, set, churn); err != nil {
			return res, err
		}
	}
	res.Files = sortedKeys(set)
	res.Churn = churn
	return res, nil
}

// accumulateBranch resuelve la base, calcula el merge-base y suma su diff.
func accumulateBranch(ctx context.Context, run gitRunner, flagBase, cfgBase string, res *scanResult, set map[string]struct{}, churn map[string]lineStat) error {
	base, err := resolveBase(ctx, run, flagBase, cfgBase)
	if err != nil {
		return err
	}
	res.Base = base
	mb, err := run(ctx, "merge-base", base, "HEAD")
	if err != nil {
		return operationalErr(err)
	}
	mergeBase := strings.TrimSpace(string(mb))
	short, err := run(ctx, "rev-parse", "--short", mergeBase)
	if err != nil {
		return operationalErr(err)
	}
	res.MergeBase = strings.TrimSpace(string(short))
	committed, err := diffFiles(ctx, run, mergeBase+"...HEAD")
	if err != nil {
		return err
	}
	addAll(set, committed)
	return accumulateChurn(ctx, run, mergeBase+"...HEAD", churn)
}

func diffFiles(ctx context.Context, run gitRunner, spec string) ([]string, error) {
	out, err := run(ctx, "diff", "--name-only", "--diff-filter=ACMRD", "-M", spec)
	if err != nil {
		return nil, operationalErr(err)
	}
	return parseNameOnly(out), nil
}

func headExists(ctx context.Context, run gitRunner) bool {
	_, err := run(ctx, "rev-parse", "--verify", "--quiet", "HEAD")
	return err == nil
}

func addAll(set map[string]struct{}, paths []string) {
	for _, p := range paths {
		set[p] = struct{}{}
	}
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
