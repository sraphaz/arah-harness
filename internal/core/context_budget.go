package core

import (
	"encoding/json"
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
	Budget            ContextBudget `json:"budget"`
	TaskID            string        `json:"task_id"`
	State             State         `json:"state"`
	PrimaryExecutor   string        `json:"primary_executor"`
	Objective         string        `json:"objective"`
	ChoreographyRule  string        `json:"choreography_rule,omitempty"`
	WorkClass         WorkClass     `json:"work_class,omitempty"`
	Area              string        `json:"area,omitempty"`
	AllowedPaths      []string      `json:"allowed_paths,omitempty"`
	Consultants       []string      `json:"consultants,omitempty"`
	Limits            *Limits       `json:"limits,omitempty"`
	RecentEventKinds  []string      `json:"recent_event_kinds,omitempty"`
	Evidence          []string      `json:"evidence,omitempty"`
	BlockingReason    *string       `json:"blocking_reason,omitempty"`
	BriefingMarkdown  string        `json:"briefing_markdown,omitempty"`
	Contract          *Contract     `json:"contract,omitempty"`
	Events            []Event       `json:"events,omitempty"`
	EstimatedTokens   int           `json:"estimated_tokens"`
	DisclosureNotes   []string      `json:"disclosure_notes"`
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
		tc.RecentEventKinds = recentEventKinds(events, maxEventsStandard)
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

// recentEventKinds returns up to n kinds from the tail of an ascending event list.
func recentEventKinds(events []Event, n int) []string {
	if n <= 0 || len(events) == 0 {
		return nil
	}
	start := 0
	if len(events) > n {
		start = len(events) - n
	}
	out := make([]string, 0, len(events)-start)
	for _, ev := range events[start:] {
		out = append(out, ev.Kind)
	}
	return out
}

func stringifyTaskContext(tc *TaskContext) string {
	// Projection of every model-facing field except estimated_tokens (avoid recursion).
	type eventProj struct {
		ID            string         `json:"id"`
		TaskID        string         `json:"task_id,omitempty"`
		Kind          string         `json:"kind"`
		At            string         `json:"at"`
		Payload       map[string]any `json:"payload,omitempty"`
		TraceID       string         `json:"trace_id,omitempty"`
		RunID         string         `json:"run_id,omitempty"`
		CorrelationID string         `json:"correlation_id,omitempty"`
		AgentID       string         `json:"agent_id,omitempty"`
		SessionID     string         `json:"session_id,omitempty"`
	}
	proj := struct {
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
		Events           []eventProj   `json:"events,omitempty"`
		DisclosureNotes  []string      `json:"disclosure_notes"`
	}{
		Budget:           tc.Budget,
		TaskID:           tc.TaskID,
		State:            tc.State,
		PrimaryExecutor:  tc.PrimaryExecutor,
		Objective:        tc.Objective,
		ChoreographyRule: tc.ChoreographyRule,
		WorkClass:        tc.WorkClass,
		Area:             tc.Area,
		AllowedPaths:     tc.AllowedPaths,
		Consultants:      tc.Consultants,
		Limits:           tc.Limits,
		RecentEventKinds: tc.RecentEventKinds,
		Evidence:         tc.Evidence,
		BlockingReason:   tc.BlockingReason,
		BriefingMarkdown: tc.BriefingMarkdown,
		Contract:         tc.Contract,
		DisclosureNotes:  tc.DisclosureNotes,
	}
	for _, ev := range tc.Events {
		proj.Events = append(proj.Events, eventProj{
			ID: ev.ID, TaskID: ev.TaskID, Kind: ev.Kind, At: ev.At, Payload: ev.Payload,
			TraceID: ev.TraceID, RunID: ev.RunID, CorrelationID: ev.CorrelationID,
			AgentID: ev.AgentID, SessionID: ev.SessionID,
		})
	}
	b, err := json.Marshal(proj)
	if err != nil {
		return fmt.Sprintf("%s %s %s", tc.Budget, tc.TaskID, tc.Objective)
	}
	return string(b)
}

// BaselinePromptTokens estimates the legacy "dump everything" agent context size.
func BaselinePromptTokens(agentsMD, execControlMD, contractYAML, briefing string) int {
	return EstimateTokensAny(agentsMD, execControlMD, contractYAML, briefing)
}
