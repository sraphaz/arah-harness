package core

import "testing"

func TestTransitionRejectsRerouteAfterExecuting(t *testing.T) {
	c := &Contract{
		TaskID:          "task-test",
		State:           StateExecuting,
		PrimaryExecutor: "backend",
		IntentType:      IntentExecution,
	}
	err := c.Transition(StateRouted, "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	de, ok := err.(*DomainError)
	if !ok || de.Code != "EXECUTION.REROUTE_AFTER_EXECUTING_FORBIDDEN" {
		t.Fatalf("got %#v", err)
	}
}

func TestCompleteRequiresEvidence(t *testing.T) {
	c := &Contract{
		TaskID:          "task-test",
		State:           StateExecuting,
		PrimaryExecutor: "backend",
		IntentType:      IntentExecution,
	}
	err := c.Complete(nil)
	if err == nil {
		t.Fatal("expected evidence error")
	}
	de := err.(*DomainError)
	if de.Code != "EXECUTION.COMPLETION_EVIDENCE_REQUIRED" {
		t.Fatalf("code=%s", de.Code)
	}
}

func TestCompleteHappyPath(t *testing.T) {
	c := &Contract{
		TaskID:          "task-test",
		State:           StateExecuting,
		PrimaryExecutor: "backend",
		IntentType:      IntentExecution,
	}
	if err := c.Complete([]string{"internal/core/domain.go updated; go test ./... passed"}); err != nil {
		t.Fatal(err)
	}
	if c.State != StateDone {
		t.Fatalf("state=%s", c.State)
	}
	if len(c.Execution.CompletionEvidence) == 0 {
		t.Fatal("evidence missing")
	}
}

func TestBlockRequiresReason(t *testing.T) {
	c := &Contract{TaskID: "t", State: StateExecuting, PrimaryExecutor: "backend"}
	err := c.Block("  ")
	if err == nil {
		t.Fatal("expected reason")
	}
}

func TestNewContractRejectsEmptyObjective(t *testing.T) {
	_, err := NewContract("", "backend", WorkStandard, IntentExecution, ResolvedRouting{PrimaryExecutor: "backend"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHasConcreteEvidence(t *testing.T) {
	c := &Contract{IntentType: IntentExecution}
	if c.HasConcreteEvidence([]string{"análise apenas"}) {
		t.Fatal("analysis-only should fail")
	}
	if !c.HasConcreteEvidence([]string{"docs/adr/002.md updated"}) {
		t.Fatal("path should pass")
	}
}
