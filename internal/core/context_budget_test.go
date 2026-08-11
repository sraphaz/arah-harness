package core

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestBuildTaskContextNormalizesBudget(t *testing.T) {
	c := &Contract{TaskID: "t", Objective: "o", State: StateExecuting, PrimaryExecutor: "backend"}
	tc := BuildTaskContext(c, nil, ContextBudget(" FULL "), "")
	if tc.Budget != BudgetFull || tc.Contract == nil {
		t.Fatalf("expected full disclosure for %q, got budget=%s contract=%v", " FULL ", tc.Budget, tc.Contract != nil)
	}
	tc = BuildTaskContext(c, nil, ContextBudget("MINIMAL"), "")
	if tc.Budget != BudgetMinimal {
		t.Fatalf("budget=%s", tc.Budget)
	}
}

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

func TestRecentEventKindsUsesTail(t *testing.T) {
	events := make([]Event, 0, 12)
	for i := 0; i < 12; i++ {
		events = append(events, Event{Kind: fmt.Sprintf("k%d", i)})
	}
	got := recentEventKinds(events, 8)
	if len(got) != 8 || got[0] != "k4" || got[7] != "k11" {
		t.Fatalf("got %v", got)
	}
	c := &Contract{TaskID: "t", State: StateExecuting, PrimaryExecutor: "backend", Scope: Scope{Area: "backend"}}
	std := BuildTaskContext(c, events, BudgetStandard, "")
	if std.RecentEventKinds[0] != "k4" || std.RecentEventKinds[len(std.RecentEventKinds)-1] != "k11" {
		t.Fatalf("standard used head instead of tail: %v", std.RecentEventKinds)
	}
}

func TestTruncateUTF8Safe(t *testing.T) {
	s := strings.Repeat("ação", 50) // multibyte runes
	got := truncate(s, 10)
	if !utf8.ValidString(got) {
		t.Fatalf("invalid utf8: %q", got)
	}
	runes := []rune(got)
	if runes[len(runes)-1] != '…' {
		t.Fatalf("expected ellipsis, got %q", got)
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
