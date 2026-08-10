// Package conformance holds harness conformance checks for arah-core (H-20).
// Fixtures are created in-process; proofs cover CLI≡MCP parity, dry-run, and error codes.
package conformance_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
	"github.com/sraphaz/arah-harness/internal/envelope"
	arahmcp "github.com/sraphaz/arah-harness/internal/mcp"
)

func fixtureRepo(t *testing.T) (string, *core.TaskService) {
	t.Helper()
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte(`
version: 2
rules:
  - id: craft-backend
    paths: ["backend/**", "cmd/**", "internal/**"]
    execution:
      primary_executor: backend
    agents:
      - id: backend
        type: operational
        role: executor
      - id: solutions-architect
        type: operational
        role: consultant
`), 0o644)
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	svc := &core.TaskService{Store: store, Events: store, Router: choreography.New(root)}
	return root, svc
}

func TestDryRunCreateDoesNotPersist(t *testing.T) {
	_, svc := fixtureRepo(t)
	c, path, err := svc.Create("plan only", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if path != "dry-run" || c.State != core.StateExecuting {
		t.Fatalf("path=%s state=%s", path, c.State)
	}
	if _, _, err := svc.Get(c.TaskID); err == nil {
		t.Fatal("dry-run create must not persist")
	}
}

func TestStableErrorCodes(t *testing.T) {
	_, svc := fixtureRepo(t)
	c, _, err := svc.Create("need evidence", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.Complete(c.TaskID, nil, core.MutateOptions{})
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "EXECUTION.COMPLETION_EVIDENCE_REQUIRED" {
		t.Fatalf("got %#v", err)
	}
	_, _, err = svc.Block(c.TaskID, "", core.MutateOptions{})
	de, ok = err.(*core.DomainError)
	if !ok || de.Code != "EXECUTION.BLOCKING_REASON_REQUIRED" {
		t.Fatalf("got %#v", err)
	}
}

func TestCLIMCPParityOnCreateDecision(t *testing.T) {
	_, svc := fixtureRepo(t)
	cli, _, err := svc.Create("parity", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}

	in := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"arah_create_task","arguments":{"objective":"parity","area":"backend","dry_run":true}}}` + "\n"
	var out bytes.Buffer
	srv := &arahmcp.Server{Tasks: svc, Version: "test", Reader: strings.NewReader(in), Writer: &out}
	if err := srv.Run(); err != nil {
		t.Fatal(err)
	}
	var resp map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &resp); err != nil {
		t.Fatal(err)
	}
	result := resp["result"].(map[string]any)
	if result["isError"] == true {
		t.Fatalf("mcp error: %#v", result)
	}
	text := result["content"].([]any)[0].(map[string]any)["text"].(string)
	var env envelope.Envelope
	if err := json.Unmarshal([]byte(text), &env); err != nil {
		t.Fatal(err)
	}
	data := env.Data.(map[string]any)
	if data["primary_executor"] != cli.PrimaryExecutor {
		t.Fatalf("executor cli=%s mcp=%v", cli.PrimaryExecutor, data["primary_executor"])
	}
	if data["state"] != string(cli.State) {
		t.Fatalf("state cli=%s mcp=%v", cli.State, data["state"])
	}
	if data["dry_run"] != true {
		t.Fatalf("dry_run=%v", data["dry_run"])
	}
}

func TestCompleteDryRunLeavesTaskExecuting(t *testing.T) {
	_, svc := fixtureRepo(t)
	c, _, err := svc.Create("stay executing", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	planned, path, err := svc.Complete(c.TaskID, []string{"file.go updated"}, core.MutateOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if planned.State != core.StateDone || !strings.HasPrefix(path, "dry-run") {
		t.Fatalf("planned=%s path=%s", planned.State, path)
	}
	got, _, err := svc.Get(c.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != core.StateExecuting {
		t.Fatalf("persisted state mutated: %s", got.State)
	}
}
