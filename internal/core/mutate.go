package core

import (
	"fmt"
	"strings"
)

// MutationResult is the plan→validate→apply outcome for task mutations (H-14).
type MutationResult struct {
	Contract   *Contract
	Path       string
	Diff       string // textual before→after summary (empty when idempotent no-op)
	Idempotent bool
	DryRun     bool
}

type mutateSnapshot struct {
	State            State
	Evidence         []string
	BlockingReason   string
	PrimaryExecutor  string
	Objective        string
	ChoreographyRule string
}

func snapContract(c *Contract) mutateSnapshot {
	if c == nil {
		return mutateSnapshot{}
	}
	br := ""
	if c.Result.BlockingReason != nil {
		br = *c.Result.BlockingReason
	}
	return mutateSnapshot{
		State:            c.State,
		Evidence:         append([]string{}, c.Execution.CompletionEvidence...),
		BlockingReason:   br,
		PrimaryExecutor:  c.PrimaryExecutor,
		Objective:        c.Objective,
		ChoreographyRule: c.ChoreographyRule,
	}
}

func formatMutationDiff(before, after mutateSnapshot) string {
	var b strings.Builder
	line := func(prefix, text string) {
		b.WriteString(prefix)
		b.WriteString(text)
		b.WriteByte('\n')
	}
	if before.State == "" && after.State != "" {
		line("+ ", fmt.Sprintf("state: %s", after.State))
	} else if before.State != after.State {
		line("- ", fmt.Sprintf("state: %s", before.State))
		line("+ ", fmt.Sprintf("state: %s", after.State))
	}
	if before.PrimaryExecutor == "" && after.PrimaryExecutor != "" {
		line("+ ", fmt.Sprintf("primary_executor: %s", after.PrimaryExecutor))
	} else if before.PrimaryExecutor != after.PrimaryExecutor {
		line("- ", fmt.Sprintf("primary_executor: %s", before.PrimaryExecutor))
		line("+ ", fmt.Sprintf("primary_executor: %s", after.PrimaryExecutor))
	}
	if before.Objective == "" && after.Objective != "" {
		line("+ ", fmt.Sprintf("objective: %s", after.Objective))
	} else if before.Objective != after.Objective {
		line("- ", fmt.Sprintf("objective: %s", before.Objective))
		line("+ ", fmt.Sprintf("objective: %s", after.Objective))
	}
	if before.ChoreographyRule == "" && after.ChoreographyRule != "" {
		line("+ ", fmt.Sprintf("choreography_rule: %s", after.ChoreographyRule))
	}
	addedEv := diffStrings(before.Evidence, after.Evidence)
	for _, e := range addedEv {
		line("+ ", "evidence: "+e)
	}
	if before.BlockingReason != after.BlockingReason {
		if before.BlockingReason != "" {
			line("- ", "blocking_reason: "+before.BlockingReason)
		}
		if after.BlockingReason != "" {
			line("+ ", "blocking_reason: "+after.BlockingReason)
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func diffStrings(before, after []string) []string {
	seen := map[string]bool{}
	for _, s := range before {
		seen[s] = true
	}
	var out []string
	for _, s := range after {
		if s == "" || seen[s] {
			continue
		}
		out = append(out, s)
	}
	return out
}

func evidenceSameSet(c *Contract, evidence []string) bool {
	if c == nil {
		return false
	}
	req := normalizedEvidenceSet(evidence)
	if len(req) == 0 {
		return false
	}
	have := normalizedEvidenceSet(c.Execution.CompletionEvidence)
	if len(have) != len(req) {
		return false
	}
	for e := range req {
		if !have[e] {
			return false
		}
	}
	return true
}

func normalizedEvidenceSet(items []string) map[string]bool {
	out := map[string]bool{}
	for _, e := range items {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		out[e] = true
	}
	return out
}
