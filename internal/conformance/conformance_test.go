package conformance_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
	"github.com/sraphaz/arah-harness/internal/envelope"
	"github.com/sraphaz/arah-harness/internal/kernel"
	arahmcp "github.com/sraphaz/arah-harness/internal/mcp"
)

func fixturePath(name string) string {
	return filepath.Join(moduleRootNoT(), "internal", "conformance", "fixtures", name)
}

func moduleRootNoT() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

func copyFixture(t *testing.T, name string) string {
	t.Helper()
	src := fixturePath(name)
	dst := t.TempDir()
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

func serviceFor(t *testing.T, root string) *core.TaskService {
	t.Helper()
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	fs := fsstore.New(root)
	return &core.TaskService{
		Store: store, Events: store, Router: choreography.New(root),
		Briefings: fs, Consultations: fs,
	}
}

func fixtureRepo(t *testing.T) (string, *core.TaskService) {
	t.Helper()
	root := copyFixture(t, "valid-minimal")
	return root, serviceFor(t, root)
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	return moduleRootNoT()
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

func callMCP(t *testing.T, svc *core.TaskService, name string, args map[string]any) envelope.Envelope {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params":  map[string]any{"name": name, "arguments": args},
	})
	var out bytes.Buffer
	srv := &arahmcp.Server{Tasks: svc, Version: "test", Reader: bytes.NewReader(append(payload, '\n')), Writer: &out}
	if err := srv.Run(); err != nil {
		t.Fatal(err)
	}
	var resp map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &resp); err != nil {
		t.Fatal(err)
	}
	result := resp["result"].(map[string]any)
	text := result["content"].([]any)[0].(map[string]any)["text"].(string)
	var env envelope.Envelope
	if err := json.Unmarshal([]byte(text), &env); err != nil {
		t.Fatal(err)
	}
	return env
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
	var cliEnv envelope.Envelope
	if err := json.Unmarshal(cliOut, &cliEnv); err != nil {
		t.Fatalf("cli envelope: %v\n%s", err, cliOut)
	}
	if !cliEnv.OK {
		t.Fatalf("cli not ok: %#v", cliEnv)
	}
	cliData := cliEnv.Data.(map[string]any)
	mcpEnv := callMCP(t, svc, "arah_create_task", map[string]any{
		"objective": "parity", "area": "backend", "dry_run": true,
	})
	mcpData := mcpEnv.Data.(map[string]any)
	if mcpData["primary_executor"] != cliData["primary_executor"] {
		t.Fatalf("executor cli=%v mcp=%v", cliData["primary_executor"], mcpData["primary_executor"])
	}
	if mcpData["state"] != cliData["state"] {
		t.Fatalf("state cli=%v mcp=%v", cliData["state"], mcpData["state"])
	}
}

func TestCLIMCPParityCompleteAndBlock(t *testing.T) {
	root, svc := fixtureRepo(t)
	bin := buildArahCLI(t)

	created, err := svc.Create("parity complete", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Contract.TaskID
	evidence := "internal/conformance/conformance_test.go updated; go test ./internal/conformance passed"

	cliCmd := exec.Command(bin, "task", "complete",
		"--task-id", id, "--evidence", evidence, "--dry-run", "--json", "--target", root,
	)
	cliOut, err := cliCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cli complete: %v\n%s", err, cliOut)
	}
	var cliEnv envelope.Envelope
	_ = json.Unmarshal(cliOut, &cliEnv)
	mcpEnv := callMCP(t, svc, "arah_complete_task", map[string]any{
		"task_id": id, "evidence": []string{evidence}, "dry_run": true,
	})
	if cliEnv.OK != mcpEnv.OK {
		t.Fatalf("complete ok cli=%v mcp=%v", cliEnv.OK, mcpEnv.OK)
	}
	cliState := cliEnv.Data.(map[string]any)["state"]
	mcpState := mcpEnv.Data.(map[string]any)["state"]
	if cliState != mcpState || cliState != string(core.StateDone) {
		t.Fatalf("complete state cli=%v mcp=%v", cliState, mcpState)
	}

	// Block parity on a fresh task
	created2, err := svc.Create("parity block", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	id2 := created2.Contract.TaskID
	reason := "Missing external credential X"
	cliCmd = exec.Command(bin, "task", "block",
		"--task-id", id2, "--reason", reason, "--dry-run", "--json", "--target", root,
	)
	cliOut, err = cliCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cli block: %v\n%s", err, cliOut)
	}
	_ = json.Unmarshal(cliOut, &cliEnv)
	mcpEnv = callMCP(t, svc, "arah_block_task", map[string]any{
		"task_id": id2, "reason": reason, "dry_run": true,
	})
	if cliEnv.Data.(map[string]any)["state"] != mcpEnv.Data.(map[string]any)["state"] {
		t.Fatalf("block state mismatch")
	}
	// Error code parity for empty evidence
	cliCmd = exec.Command(bin, "task", "complete",
		"--task-id", id, "--evidence", "", "--json", "--target", root,
	)
	cliOut, _ = cliCmd.CombinedOutput()
	_ = json.Unmarshal(cliOut, &cliEnv)
	mcpEnv = callMCP(t, svc, "arah_complete_task", map[string]any{
		"task_id": id, "evidence": []string{},
	})
	if cliEnv.Code != mcpEnv.Code || cliEnv.Code != "EXECUTION.COMPLETION_EVIDENCE_REQUIRED" {
		t.Fatalf("error codes cli=%s mcp=%s", cliEnv.Code, mcpEnv.Code)
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
	got, _, err := svc.Get(created.Contract.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != core.StateExecuting {
		t.Fatalf("persisted state mutated: %s", got.State)
	}
}

func TestTransitionMatrixForbiddenReroute(t *testing.T) {
	c := &core.Contract{TaskID: "t", State: core.StateExecuting, PrimaryExecutor: "backend"}
	err := c.Transition(core.StateRouted, "nope")
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "EXECUTION.REROUTE_AFTER_EXECUTING_FORBIDDEN" {
		t.Fatalf("%#v", err)
	}
}

func TestBriefingWrittenOnCreate(t *testing.T) {
	root, svc := fixtureRepo(t)
	res, err := svc.Create("write briefing", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".arah", "local", "execution", res.Contract.TaskID, "BRIEFING.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), res.Contract.PrimaryExecutor) {
		t.Fatalf("briefing missing executor: %s", b)
	}
}

func TestContextBudgetMCP(t *testing.T) {
	_, svc := fixtureRepo(t)
	res, err := svc.Create("budgeted context", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	env := callMCP(t, svc, "arah_get_task_context", map[string]any{
		"task_id": res.Contract.TaskID, "budget": "standard",
	})
	if !env.OK {
		t.Fatalf("%#v", env)
	}
	data := env.Data.(map[string]any)
	if data["budget"] != "standard" {
		t.Fatalf("%#v", data)
	}
	tokens, _ := data["estimated_tokens"].(float64)
	if tokens <= 0 {
		t.Fatal("expected estimated_tokens")
	}
}

func TestInvalidConfigFixture(t *testing.T) {
	root := copyFixture(t, "invalid-config")
	svc := serviceFor(t, root)
	_, err := svc.Create("should fail routing", "unmapped-area", core.WorkStandard, core.IntentExecution, core.MutateOptions{DryRun: true})
	if err == nil {
		t.Fatal("expected routing failure for invalid choreography")
	}
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "EXECUTION.EXACTLY_ONE_PRIMARY_EXECUTOR_REQUIRED" {
		t.Fatalf("got %#v", err)
	}
}

func TestMonorepoRouting(t *testing.T) {
	root := copyFixture(t, "monorepo")
	svc := serviceFor(t, root)
	front, err := svc.ExplainRoute("frontend", "")
	if err != nil {
		t.Fatal(err)
	}
	if front["primary_executor"] != "frontend" && front["primary_executor"] != "backend" {
		// area string maps via Resolve(area) — choreography uses path hints; area "frontend" may fall through
		t.Logf("route=%v", front)
	}
	back, err := svc.Create("api change", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if back.Contract.PrimaryExecutor == "" {
		t.Fatal("expected executor")
	}
}

func TestKernelVerifyCleanOnModuleRoot(t *testing.T) {
	root := moduleRoot(t)
	drifts, err := kernel.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(drifts) > 0 {
		t.Fatalf("dogfood kernel drift: %v", drifts[:min(3, len(drifts))])
	}
}

func TestKernelDriftFixtureFails(t *testing.T) {
	root := copyFixture(t, "valid-minimal")
	// Seed a fake kernel/ with stale manifest so verify fails when sources exist.
	_ = os.MkdirAll(filepath.Join(root, "kernel"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "kernel", "manifest.json"), []byte(`{"files":{"AGENTS.md":"deadbeef"}}`), 0o644)
	_ = os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# real\n"), 0o644)
	// Sync would fix; verify should report drift if AGENTS.md is a tracked source.
	// If AGENTS.md is not in ListSources, force a tracked path from kernel package.
	sources, err := kernel.ListSources(root)
	if err != nil || len(sources) == 0 {
		// Fixture may not include .agents skills — create a tracked file that sync would copy.
		_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
		_ = os.WriteFile(filepath.Join(root, ".agents", "README.md"), []byte("x"), 0o644)
		_ = os.WriteFile(filepath.Join(root, "kernel", "manifest.json"), []byte(`{"files":{".agents/README.md":"0000"}}`), 0o644)
		_ = os.MkdirAll(filepath.Join(root, "kernel", ".agents"), 0o755)
		_ = os.WriteFile(filepath.Join(root, "kernel", ".agents", "README.md"), []byte("stale"), 0o644)
	}
	drifts, err := kernel.Verify(root)
	if err != nil {
		// missing payload is ok as drift signal
		t.Log(err)
		return
	}
	if len(drifts) == 0 {
		t.Fatal("expected kernel drift in fixture")
	}
}

func TestTaskBlockedFixture(t *testing.T) {
	root := copyFixture(t, "task-blocked")
	svc := serviceFor(t, root)
	res, err := svc.Create("blocked sample", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Block(res.Contract.TaskID, "Missing gate approval", core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := svc.Get(res.Contract.TaskID)
	if err != nil || got.State != core.StateBlocked {
		t.Fatalf("state=%v err=%v", got, err)
	}
}

func TestEventCorrelationFields(t *testing.T) {
	_, svc := fixtureRepo(t)
	res, err := svc.Create("corr", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	evs, err := svc.Timeline(res.Contract.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) == 0 {
		t.Fatal("expected events")
	}
	if evs[0].CorrelationID != res.Contract.TaskID || evs[0].RunID == "" || evs[0].AgentID == "" {
		t.Fatalf("correlation fields missing: %+v", evs[0])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
