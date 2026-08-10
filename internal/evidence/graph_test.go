package evidence_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
	"github.com/sraphaz/arah-harness/internal/evidence"
)

func writeDemoSpec(t *testing.T, root string) {
	t.Helper()
	_ = os.MkdirAll(filepath.Join(root, "docs", "specs"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "docs", "specs", "demo.spec.yaml"), []byte(
		"id: demo-spec\ntitle: Demo\ncovers:\n  - internal/core/\ndepends_on:\n  - base-spec\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "docs", "specs", "base.spec.yaml"), []byte(
		"id: base-spec\ntitle: Base\ncovers:\n  - docs/\n"), 0o644)
}

func TestEvidenceGraphFromTask(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644)
	writeDemoSpec(t, root)

	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	svc := &core.TaskService{Store: store, Events: store, Router: choreography.New(root)}
	created, err := svc.Create("ship core change", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Complete(created.Contract.TaskID, []string{"internal/core/domain.go updated"}, core.MutateOptions{}); err != nil {
		t.Fatal(err)
	}

	g, err := (&evidence.Builder{RepoRoot: root, Store: store, Events: store}).Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) < 2 || len(g.Edges) < 1 {
		t.Fatalf("nodes=%d edges=%d", len(g.Nodes), len(g.Edges))
	}
	hasSpec := false
	hasAssigned := false
	hasImplements := false
	hasProduced := false
	hasDepends := false
	hasRun := false
	for _, n := range g.Nodes {
		if n.Type == "spec" {
			hasSpec = true
		}
		if n.Type == "run" {
			hasRun = true
		}
	}
	for _, e := range g.Edges {
		switch e.Rel {
		case "assigned_to":
			hasAssigned = true
		case "implements":
			hasImplements = true
		case "produced":
			hasProduced = true
		case "depends_on":
			hasDepends = true
		}
	}
	if !hasSpec || !hasAssigned || !hasImplements || !hasProduced || !hasDepends || !hasRun {
		t.Fatalf("expected spec+assigned+implements+produced+depends_on+run, graph=%+v", g)
	}
}

func TestEvidenceGraphDeterministicExport(t *testing.T) {
	root := t.TempDir()
	writeDemoSpec(t, root)
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	svc := &core.TaskService{Store: store, Events: store, Router: choreography.New(root)}
	created, err := svc.Create("stable", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Complete(created.Contract.TaskID, []string{"docs/ROADMAP.md touched"}, core.MutateOptions{}); err != nil {
		t.Fatal(err)
	}
	b := &evidence.Builder{RepoRoot: root, Store: store, Events: store}
	a, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	c, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	ra, _ := json.Marshal(a)
	rc, _ := json.Marshal(c)
	if string(ra) != string(rc) {
		t.Fatalf("evidence graph export not stable\nA=%s\nB=%s", ra, rc)
	}
}

func TestEvidenceGraphRejectsFreeTextImplements(t *testing.T) {
	root := t.TempDir()
	writeDemoSpec(t, root)
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	c := &core.Contract{
		Version: "1.0", TaskID: "task-x", Objective: "please implement demo-spec runtime-cohesion",
		State: core.StateDone, PrimaryExecutor: "backend",
		WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		Execution: core.Execution{CompletionEvidence: []string{"ran tests only"}},
	}
	if _, err := store.Save(c); err != nil {
		t.Fatal(err)
	}
	g, err := (&evidence.Builder{RepoRoot: root, Store: store}).Build()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range g.Edges {
		if e.Rel == "implements" {
			t.Fatalf("free-text objective must not create implements: %+v", e)
		}
	}
}

func TestEvidenceNodeIDsDoNotCollideOnPrefix(t *testing.T) {
	root := t.TempDir()
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	prefix := strings.Repeat("a", 64)
	a := &core.Contract{
		Version: "1.0", TaskID: "task-a", Objective: "a", State: core.StateDone,
		PrimaryExecutor: "backend", WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		Execution: core.Execution{CompletionEvidence: []string{prefix + "-one"}},
	}
	b := &core.Contract{
		Version: "1.0", TaskID: "task-b", Objective: "b", State: core.StateDone,
		PrimaryExecutor: "backend", WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		Execution: core.Execution{CompletionEvidence: []string{prefix + "-two"}},
	}
	if _, err := store.Save(a); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(b); err != nil {
		t.Fatal(err)
	}
	g, err := (&evidence.Builder{RepoRoot: root, Store: store}).Build()
	if err != nil {
		t.Fatal(err)
	}
	evidenceNodes := 0
	ids := map[string]string{}
	for _, n := range g.Nodes {
		if n.Type != "evidence" {
			continue
		}
		evidenceNodes++
		if prev, ok := ids[n.ID]; ok && prev != n.Label {
			t.Fatalf("colliding evidence id %s for %q and %q", n.ID, prev, n.Label)
		}
		ids[n.ID] = n.Label
	}
	if evidenceNodes < 2 {
		t.Fatalf("expected distinct evidence nodes, got %d", evidenceNodes)
	}
}
