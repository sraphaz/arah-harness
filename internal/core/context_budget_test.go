package core

import (
	"strings"
	"testing"
)

func TestBuildTaskContextBudgets(t *testing.T) {
	c := &Contract{
		TaskID:          "task-ctx",
		Objective:       strings.Repeat("x", 400),
		State:           StateExecuting,
		PrimaryExecutor: "backend",
		WorkClass:       WorkStandard,
		Scope:           Scope{Area: "backend", AllowedPaths: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m"}},
		Participants:    Participants{Consultants: []string{"solutions-architect"}},
		Limits:          Limits{1, 1, 1},
	}
	events := []Event{{Kind: "task.created"}, {Kind: "task.started"}}
	min := BuildTaskContext(c, events, BudgetMinimal, "")
	if min.EstimatedTokens <= 0 || len(min.Objective) >= 400 {
		t.Fatalf("minimal should truncate: %#v", min)
	}
	std := BuildTaskContext(c, events, BudgetStandard, "")
	if len(std.AllowedPaths) != maxPathsStandard {
		t.Fatalf("standard paths=%d", len(std.AllowedPaths))
	}
	full := BuildTaskContext(c, events, BudgetFull, RenderBriefing(c))
	if full.Contract == nil || full.BriefingMarkdown == "" {
		t.Fatal("full missing contract/briefing")
	}
	if !(min.EstimatedTokens < std.EstimatedTokens && std.EstimatedTokens < full.EstimatedTokens) {
		t.Fatalf("token order min=%d std=%d full=%d", min.EstimatedTokens, std.EstimatedTokens, full.EstimatedTokens)
	}
}

func TestRenderBriefingContainsExecutor(t *testing.T) {
	c := &Contract{
		TaskID: "task-1", Objective: "ship", PrimaryExecutor: "spec-steward",
		Scope:  Scope{Area: "harness", AllowedPaths: []string{"internal/**"}},
		Limits: Limits{2, 2, 1},
	}
	b := RenderBriefing(c)
	if !strings.Contains(b, "spec-steward") || !strings.Contains(b, "task-1") {
		t.Fatalf("briefing incomplete: %s", b)
	}
}
