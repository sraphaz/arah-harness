// Package conformance holds harness conformance checks for arah-core (H-20).
// Fixtures are created in-process; proofs cover CLI↔MCP parity, dry-run, and error codes.
package conformance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
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

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

func buildArahCLI(t *testing.T) string {
	t.Helper()
	name := "arah"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/arah")
	cmd.Dir = moduleRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./cmd/arah: %v\n%s", err, out)
	}
	return bin
}

func TestDryRunCreateDoesNotPersist(t *testing.T) {
	_, svc := fixtureRepo(t)
	res, err := svc.Create("plan only", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Path != "dry-run" || res.Contract.State != core.StateExecuting {
		t.Fatalf("path=%s state=%s", res.Path, res.Contract.State)
	}
	if _, _, err := svc.Get(res.Contract.TaskID); err == nil {
		t.Fatal("dry-run create must not persist")
	}
}

func TestStableErrorCodes(t *testing.T) {
	_, svc := fixtureRepo(t)
	created, err := svc.Create("need evidence", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Complete(created.Contract.TaskID, nil, core.MutateOptions{})
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "EXECUTION.COMPLETION_EVIDENCE_REQUIRED" {
		t.Fatalf("got %#v", err)
	}
	_, err = svc.Block(created.Contract.TaskID, "", core.MutateOptions{})
	de, ok = err.(*core.DomainError)
	if !ok || de.Code != "EXECUTION.BLOCKING_REASON_REQUIRED" {
		t.Fatalf("got %#v", err)
	}
}

func TestCLIMCPParityOnCreateDecision(t *testing.T) {
	root, svc := fixtureRepo(t)
	bin := buildArahCLI(t)

	cliCmd := exec.Command(bin, "task", "create",
		"--objective", "parity",
		"--area", "backend",
		"--dry-run",
		"--json",
		"--target", root,
	)
	cliOut, err := cliCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cli create: %v\n%s", err, cliOut)
	}
	cliEnv := mustEnvelope(t, cliOut)
	if !cliEnv.OK {
		t.Fatalf("cli not ok: %#v", cliEnv)
	}
	cliData := mustDataMap(t, cliEnv)

	mcpEnv, isErr := callMCPTool(t, svc, "arah_create_task", map[string]any{
		"objective": "parity",
		"area":      "backend",
		"dry_run":   true,
	})
	if isErr {
		t.Fatalf("mcp error: %#v", mcpEnv)
	}
	mcpData := mustDataMap(t, mcpEnv)

	assertParityFields(t, cliData, mcpData, "primary_executor", "state", "dry_run")
	if cliData["state"] != string(core.StateExecuting) {
		t.Fatalf("expected executing plan, got %v", cliData["state"])
	}
}

func TestCLIMCPParityOnCompleteDryRun(t *testing.T) {
	root, svc := fixtureRepo(t)
	bin := buildArahCLI(t)
	created, err := svc.Create("complete parity", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Contract.TaskID
	ev := "path/file.go updated"

	cliCmd := exec.Command(bin, "task", "complete",
		"--task-id", id,
		"--evidence", ev,
		"--dry-run",
		"--json",
		"--target", root,
	)
	cliOut, err := cliCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cli complete: %v\n%s", err, cliOut)
	}
	cliEnv := mustEnvelope(t, cliOut)
	if !cliEnv.OK {
		t.Fatalf("cli not ok: %#v", cliEnv)
	}
	cliData := mustDataMap(t, cliEnv)

	mcpEnv, isErr := callMCPTool(t, svc, "arah_complete_task", map[string]any{
		"task_id":  id,
		"evidence": []any{ev},
		"dry_run":  true,
	})
	if isErr {
		t.Fatalf("mcp error: %#v", mcpEnv)
	}
	mcpData := mustDataMap(t, mcpEnv)

	assertParityFields(t, cliData, mcpData, "state", "dry_run", "idempotent")
	if cliData["state"] != string(core.StateDone) {
		t.Fatalf("expected done plan, got %v", cliData["state"])
	}
	if cliData["dry_run"] != true || mcpData["dry_run"] != true {
		t.Fatalf("dry_run cli=%v mcp=%v", cliData["dry_run"], mcpData["dry_run"])
	}
	got, _, err := svc.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != core.StateExecuting {
		t.Fatalf("persisted state mutated: %s", got.State)
	}
}

func TestCLIMCPParityOnBlockDryRun(t *testing.T) {
	root, svc := fixtureRepo(t)
	bin := buildArahCLI(t)
	created, err := svc.Create("block parity", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Contract.TaskID
	reason := "missing credential X"

	cliCmd := exec.Command(bin, "task", "block",
		"--task-id", id,
		"--reason", reason,
		"--dry-run",
		"--json",
		"--target", root,
	)
	cliOut, err := cliCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cli block: %v\n%s", err, cliOut)
	}
	cliEnv := mustEnvelope(t, cliOut)
	if !cliEnv.OK {
		t.Fatalf("cli not ok: %#v", cliEnv)
	}
	cliData := mustDataMap(t, cliEnv)

	mcpEnv, isErr := callMCPTool(t, svc, "arah_block_task", map[string]any{
		"task_id": id,
		"reason":  reason,
		"dry_run": true,
	})
	if isErr {
		t.Fatalf("mcp error: %#v", mcpEnv)
	}
	mcpData := mustDataMap(t, mcpEnv)

	assertParityFields(t, cliData, mcpData, "state", "dry_run", "idempotent")
	if cliData["state"] != string(core.StateBlocked) {
		t.Fatalf("expected blocked plan, got %v", cliData["state"])
	}
	got, _, err := svc.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != core.StateExecuting {
		t.Fatalf("persisted state mutated: %s", got.State)
	}
}

func TestMCPStableErrorCodes(t *testing.T) {
	_, svc := fixtureRepo(t)
	created, err := svc.Create("need evidence", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Contract.TaskID

	env, isErr := callMCPTool(t, svc, "arah_complete_task", map[string]any{
		"task_id":  id,
		"evidence": []any{},
	})
	if !isErr {
		t.Fatal("expected MCP isError for empty evidence")
	}
	if env.OK || env.Code != "EXECUTION.COMPLETION_EVIDENCE_REQUIRED" {
		t.Fatalf("complete code=%s ok=%v", env.Code, env.OK)
	}

	env, isErr = callMCPTool(t, svc, "arah_block_task", map[string]any{
		"task_id": id,
		"reason":  "",
	})
	if !isErr {
		t.Fatal("expected MCP isError for empty reason")
	}
	if env.OK || env.Code != "EXECUTION.BLOCKING_REASON_REQUIRED" {
		t.Fatalf("block code=%s ok=%v", env.Code, env.OK)
	}
}

func TestCompleteDryRunLeavesTaskExecuting(t *testing.T) {
	_, svc := fixtureRepo(t)
	created, err := svc.Create("stay executing", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	planned, err := svc.Complete(created.Contract.TaskID, []string{"file.go updated"}, core.MutateOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if planned.Contract.State != core.StateDone || !strings.HasPrefix(planned.Path, "dry-run") {
		t.Fatalf("planned=%s path=%s", planned.Contract.State, planned.Path)
	}
	if planned.Diff == "" || !strings.Contains(planned.Diff, "+ state: done") {
		t.Fatalf("expected complete diff, got %q", planned.Diff)
	}
	got, _, err := svc.Get(created.Contract.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != core.StateExecuting {
		t.Fatalf("persisted state mutated: %s", got.State)
	}
}

func callMCPTool(t *testing.T, svc *core.TaskService, name string, args map[string]any) (envelope.Envelope, bool) {
	t.Helper()
	params := map[string]any{"name": name, "arguments": args}
	rawParams, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	in := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":%s}`+"\n", rawParams)
	var out bytes.Buffer
	srv := &arahmcp.Server{Tasks: svc, Version: "test", Reader: strings.NewReader(in), Writer: &out}
	if err := srv.Run(); err != nil {
		t.Fatal(err)
	}
	var resp map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &resp); err != nil {
		t.Fatalf("rpc decode: %v\n%s", err, out.String())
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result: %#v", resp)
	}
	isErr, _ := result["isError"].(bool)
	text := result["content"].([]any)[0].(map[string]any)["text"].(string)
	var env envelope.Envelope
	if err := json.Unmarshal([]byte(text), &env); err != nil {
		t.Fatalf("envelope: %v\n%s", err, text)
	}
	return env, isErr
}

func mustEnvelope(t *testing.T, raw []byte) envelope.Envelope {
	t.Helper()
	var env envelope.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("envelope: %v\n%s", err, raw)
	}
	return env
}

func mustDataMap(t *testing.T, env envelope.Envelope) map[string]any {
	t.Helper()
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type %T", env.Data)
	}
	return data
}

func assertParityFields(t *testing.T, cli, mcp map[string]any, fields ...string) {
	t.Helper()
	for _, f := range fields {
		cv, cok := cli[f]
		mv, mok := mcp[f]
		if !cok || !mok {
			t.Fatalf("%s missing: cli_ok=%v mcp_ok=%v", f, cok, mok)
		}
		if reflect.TypeOf(cv) != reflect.TypeOf(mv) {
			t.Fatalf("%s type cli=%T mcp=%T values cli=%v mcp=%v", f, cv, mv, cv, mv)
		}
		if !reflect.DeepEqual(cv, mv) {
			t.Fatalf("%s cli=%v mcp=%v", f, cv, mv)
		}
	}
}
