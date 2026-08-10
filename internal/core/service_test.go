package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/core"
)

func testSvc(t *testing.T) *core.TaskService {
	t.Helper()
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte(`
version: 2
rules:
  - id: craft-backend
    paths: ["backend/**"]
    execution:
      primary_executor: backend
    agents:
      - id: backend
        type: operational
        role: executor
`), 0o644)
	return &core.TaskService{Store: fsstore.New(root), Router: choreography.New(root)}
}

func TestCreateIncludesDiff(t *testing.T) {
	svc := testSvc(t)
	res, err := svc.Create("ship diff", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Idempotent || res.Diff == "" {
		t.Fatalf("expected non-idempotent diff, got %#v", res)
	}
	if !strings.Contains(res.Diff, "+ state: executing") {
		t.Fatalf("diff missing state: %q", res.Diff)
	}
	if !strings.Contains(res.Diff, "+ primary_executor: backend") {
		t.Fatalf("diff missing executor: %q", res.Diff)
	}
}

func TestCompleteIdempotent(t *testing.T) {
	svc := testSvc(t)
	created, err := svc.Create("done twice", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	ev := []string{"internal/core/mutate.go updated", "go test ./internal/core passed"}
	first, err := svc.Complete(created.Contract.TaskID, ev, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if first.Idempotent || !strings.Contains(first.Diff, "+ state: done") {
		t.Fatalf("first complete: %#v diff=%q", first.Idempotent, first.Diff)
	}
	second, err := svc.Complete(created.Contract.TaskID, ev, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Idempotent {
		t.Fatal("expected idempotent re-complete")
	}
	if second.Diff != "" {
		t.Fatalf("idempotent diff should be empty, got %q", second.Diff)
	}
	if second.Contract.State != core.StateDone {
		t.Fatalf("state=%s", second.Contract.State)
	}
	// Partial evidence must not be treated as an idempotent replay.
	_, err = svc.Complete(created.Contract.TaskID, []string{ev[0]}, core.MutateOptions{})
	if err == nil {
		t.Fatal("expected terminal error for partial evidence replay")
	}
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "EXECUTION.TERMINAL_STATE_IMMUTABLE" {
		t.Fatalf("got %#v", err)
	}
}

func TestBlockIdempotent(t *testing.T) {
	svc := testSvc(t)
	created, err := svc.Create("block twice", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	reason := "missing credential X"
	first, err := svc.Block(created.Contract.TaskID, reason, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if first.Idempotent || !strings.Contains(first.Diff, "+ blocking_reason: "+reason) {
		t.Fatalf("first block: idempotent=%v diff=%q", first.Idempotent, first.Diff)
	}
	second, err := svc.Block(created.Contract.TaskID, reason, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Idempotent || second.Diff != "" {
		t.Fatalf("expected idempotent block, got %#v", second)
	}
}
