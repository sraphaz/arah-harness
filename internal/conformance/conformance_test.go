// Package conformance holds harness conformance checks for arah-core (H-20).
// Fixtures are created in-process; proofs cover CLI↔MCP parity, dry-run, and error codes.
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
	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
	"github.com/sraphaz/arah-harness/internal/envelope"
	"github.com/sraphaz/arah-harness/internal/evidence"
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
	var cliEnv envelope.Envelope
	if err := json.Unmarshal(cliOut, &cliEnv); err != nil {
		t.Fatalf("cli envelope: %v\n%s", err, cliOut)
	}
	if !cliEnv.OK {
		t.Fatalf("cli not ok: %#v", cliEnv)
	}
	cliData, ok := cliEnv.Data.(map[string]any)
	if !ok {
		t.Fatalf("cli data type %T", cliEnv.Data)
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
	var mcpEnv envelope.Envelope
	if err := json.Unmarshal([]byte(text), &mcpEnv); err != nil {
		t.Fatal(err)
	}
	mcpData := mcpEnv.Data.(map[string]any)

	if mcpData["primary_executor"] != cliData["primary_executor"] {
		t.Fatalf("executor cli=%v mcp=%v", cliData["primary_executor"], mcpData["primary_executor"])
	}
	if mcpData["state"] != cliData["state"] {
		t.Fatalf("state cli=%v mcp=%v", cliData["state"], mcpData["state"])
	}
	if cliData["dry_run"] != true || mcpData["dry_run"] != true {
		t.Fatalf("dry_run cli=%v mcp=%v", cliData["dry_run"], mcpData["dry_run"])
	}
	if cliData["state"] != string(core.StateExecuting) {
		t.Fatalf("expected executing plan, got %v", cliData["state"])
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

func TestCLIMCPParityOnEvidenceGraph(t *testing.T) {
	root, svc := fixtureRepo(t)
	_ = os.MkdirAll(filepath.Join(root, "docs", "specs"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "docs", "specs", "demo.spec.yaml"), []byte(
		"id: demo-spec\ntitle: Demo\ncovers:\n  - cmd/\ndepends_on:\n  - other-spec\n"), 0o644)
	created, err := svc.Create("parity graph", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Complete(created.Contract.TaskID, []string{"cmd/arah/main.go updated"}, core.MutateOptions{}); err != nil {
		t.Fatal(err)
	}

	// Release the fixture DB before the CLI subprocess opens it.
	store, ok := svc.Store.(*sqlitestore.Store)
	if !ok {
		t.Fatal("expected sqlitestore.Store")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	bin := buildArahCLI(t)
	cliOut, err := exec.Command(bin, "evidence", "graph", "--json", "--target", root).CombinedOutput()
	if err != nil {
		t.Fatalf("cli evidence graph: %v\n%s", err, cliOut)
	}
	var cliEnv envelope.Envelope
	if err := json.Unmarshal(cliOut, &cliEnv); err != nil {
		t.Fatalf("cli envelope: %v\n%s", err, cliOut)
	}
	if !cliEnv.OK {
		t.Fatalf("cli not ok: %#v", cliEnv)
	}
	assertEvidenceGraphPayload(t, cliEnv.Data)
	cliRaw, err := json.Marshal(cliEnv.Data)
	if err != nil {
		t.Fatal(err)
	}

	store2, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store2.Close() })
	svc2 := &core.TaskService{Store: store2, Events: store2, Router: svc.Router}
	eb := &evidence.Builder{RepoRoot: root, Store: store2, Events: store2}
	in := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"arah_get_evidence_graph","arguments":{}}}` + "\n"
	var out bytes.Buffer
	srv := &arahmcp.Server{Tasks: svc2, Evidence: eb, Version: "test", Reader: strings.NewReader(in), Writer: &out}
	if err := srv.Run(); err != nil {
		t.Fatal(err)
	}
	text := mcpResultText(t, out.Bytes())
	var mcpEnv envelope.Envelope
	if err := json.Unmarshal([]byte(text), &mcpEnv); err != nil {
		t.Fatal(err)
	}
	assertEvidenceGraphPayload(t, mcpEnv.Data)
	mcpRaw, err := json.Marshal(mcpEnv.Data)
	if err != nil {
		t.Fatal(err)
	}
	if string(cliRaw) != string(mcpRaw) {
		t.Fatalf("evidence graph CLI≠MCP\ncli=%s\nmcp=%s", cliRaw, mcpRaw)
	}
}

func mcpResultText(t *testing.T, raw []byte) string {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(raw))), &resp); err != nil {
		t.Fatalf("rpc decode: %v\n%s", err, raw)
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result: %#v", resp)
	}
	if result["isError"] == true {
		t.Fatalf("mcp error: %#v", result)
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("missing content: %#v", result)
	}
	item, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("bad content item: %#v", content[0])
	}
	text, ok := item["text"].(string)
	if !ok || text == "" {
		t.Fatalf("missing text: %#v", item)
	}
	return text
}

func assertEvidenceGraphPayload(t *testing.T, data any) {
	t.Helper()
	m, ok := data.(map[string]any)
	if !ok || m == nil {
		t.Fatalf("expected graph object, got %T %#v", data, data)
	}
	nodes, _ := m["nodes"].([]any)
	edges, _ := m["edges"].([]any)
	if len(nodes) == 0 || len(edges) == 0 {
		t.Fatalf("expected non-empty evidence graph, nodes=%d edges=%d", len(nodes), len(edges))
	}
}
