package core

import (
	"fmt"
	"strings"
)

// ContextBudget controls progressive disclosure of task context to agents (NOOA-inspired).
type ContextBudget string

const (
	BudgetMinimal  ContextBudget = "minimal"
	BudgetStandard ContextBudget = "standard"
	BudgetFull     ContextBudget = "full"
)

// ParseContextBudget normalizes a budget string; empty defaults to standard.
func ParseContextBudget(s string) ContextBudget {
	switch ContextBudget(strings.ToLower(strings.TrimSpace(s))) {
	case BudgetMinimal:
		return BudgetMinimal
	case BudgetFull:
		return BudgetFull
	default:
		return BudgetStandard
	}
}

// TaskContext is a budgeted, model-facing view of a task (not the full dump).
type TaskContext struct {
	Budget           ContextBudget `json:"budget"`
	TaskID           string        `json:"task_id"`
	State            State         `json:"state"`
	PrimaryExecutor  string        `json:"primary_executor"`
	Objective        string        `json:"objective"`
	ChoreographyRule string        `json:"choreography_rule,omitempty"`
	WorkClass        WorkClass     `json:"work_class,omitempty"`
	Area             string        `json:"area,omitempty"`
	AllowedPaths     []string      `json:"allowed_paths,omitempty"`
	Consultants      []string      `json:"consultants,omitempty"`
	Limits           *Limits       `json:"limits,omitempty"`
	RecentEventKinds []string      `json:"recent_event_kinds,omitempty"`
	Evidence         []string      `json:"evidence,omitempty"`
	BlockingReason   *string       `json:"blocking_reason,omitempty"`
	BriefingMarkdown string        `json:"briefing_markdown,omitempty"`
	Contract         *Contract     `json:"contract,omitempty"`
	Events           []Event       `json:"events,omitempty"`
	EstimatedTokens  int           `json:"estimated_tokens"`
	DisclosureNotes  []string      `json:"disclosure_notes"`
}

// EstimateTokens approximates LLM tokens as chars/4 (deterministic proxy).
func EstimateTokens(s string) int {
	if s == "" {
		return 0
	}
	n := (len(s) + 3) / 4
	if n < 1 {
		return 1
	}
	return n
}

// EstimateTokensAny marshals a rough size from stringified content.
func EstimateTokensAny(parts ...string) int {
	return EstimateTokens(strings.Join(parts, "\n"))
}

const maxObjectiveMinimal = 160
const maxPathsStandard = 12
const maxEventsStandard = 8

// BuildTaskContext renders a progressive-disclosure view for the given budget.
func BuildTaskContext(c *Contract, events []Event, budget ContextBudget, briefing string) *TaskContext {
	if budget == "" {
		budget = BudgetStandard
	}
	tc := &TaskContext{
		Budget:           budget,
		TaskID:           c.TaskID,
		State:            c.State,
		PrimaryExecutor:  c.PrimaryExecutor,
		Objective:        c.Objective,
		ChoreographyRule: c.ChoreographyRule,
		DisclosureNotes:  []string{},
	}

	switch budget {
	case BudgetMinimal:
		tc.Objective = truncate(c.Objective, maxObjectiveMinimal)
		tc.DisclosureNotes = append(tc.DisclosureNotes,
			"minimal: objective truncated; use budget=standard|full or arah_get_timeline for more")
	case BudgetStandard:
		tc.WorkClass = c.WorkClass
		tc.Area = c.Scope.Area
		tc.AllowedPaths = capStrings(c.Scope.AllowedPaths, maxPathsStandard)
		tc.Consultants = append([]string{}, c.Participants.Consultants...)
		lim := c.Limits
		tc.Limits = &lim
		tc.Evidence = append([]string{}, c.Execution.CompletionEvidence...)
		tc.BlockingReason = c.Result.BlockingReason
		start := 0
		if len(events) > maxEventsStandard {
			start = len(events) - maxEventsStandard
		}
		for _, ev := range events[start:] {
			tc.RecentEventKinds = append(tc.RecentEventKinds, ev.Kind)
		}
		if len(c.Scope.AllowedPaths) > maxPathsStandard {
			tc.DisclosureNotes = append(tc.DisclosureNotes,
				fmt.Sprintf("standard: allowed_paths capped at %d; use budget=full for all", maxPathsStandard))
		}
		if len(events) > maxEventsStandard {
			tc.DisclosureNotes = append(tc.DisclosureNotes,
				fmt.Sprintf("standard: only last %d event kinds; use arah_get_timeline for full history", maxEventsStandard))
		}
	case BudgetFull:
		tc.WorkClass = c.WorkClass
		tc.Area = c.Scope.Area
		tc.AllowedPaths = append([]string{}, c.Scope.AllowedPaths...)
		tc.Consultants = append([]string{}, c.Participants.Consultants...)
		lim := c.Limits
		tc.Limits = &lim
		tc.Evidence = append([]string{}, c.Execution.CompletionEvidence...)
		tc.BlockingReason = c.Result.BlockingReason
		tc.BriefingMarkdown = briefing
		cp := *c
		tc.Contract = &cp
		tc.Events = append([]Event{}, events...)
		tc.DisclosureNotes = append(tc.DisclosureNotes, "full: includes contract + timeline + briefing")
	}

	tc.EstimatedTokens = EstimateTokens(stringifyTaskContext(tc))
	return tc
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func capStrings(in []string, n int) []string {
	if len(in) <= n {
		return append([]string{}, in...)
	}
	return append([]string{}, in[:n]...)
}

func stringifyTaskContext(tc *TaskContext) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s %s %s %s ", tc.Budget, tc.TaskID, tc.State, tc.PrimaryExecutor, tc.Objective)
	b.WriteString(strings.Join(tc.AllowedPaths, " "))
	b.WriteString(strings.Join(tc.Consultants, " "))
	b.WriteString(strings.Join(tc.RecentEventKinds, " "))
	b.WriteString(strings.Join(tc.Evidence, " "))
	b.WriteString(tc.BriefingMarkdown)
	if tc.Contract != nil {
		fmt.Fprintf(&b, " contract:%s", tc.Contract.TaskID)
		b.WriteString(strings.Join(tc.Contract.Scope.AllowedPaths, " "))
		b.WriteString(tc.Contract.Objective)
	}
	for _, ev := range tc.Events {
		fmt.Fprintf(&b, " %s %s", ev.Kind, ev.At)
	}
	return b.String()
}

// BaselinePromptTokens estimates the legacy "dump everything" agent context size.
func BaselinePromptTokens(agentsMD, execControlMD, contractYAML, briefing string) int {
	return EstimateTokensAny(agentsMD, execControlMD, contractYAML, briefing)
}
