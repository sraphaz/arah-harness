package evidence_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
	"github.com/sraphaz/arah-harness/internal/evidence"
)

func TestEvidenceGraphFromTask(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "docs", "specs"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "docs", "specs", "demo.spec.yaml"), []byte(
		"id: demo-spec\ntitle: Demo\ncovers:\n  - internal/core/\n"), 0o644)

	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	svc := &core.TaskService{Store: store, Events: store, Router: choreography.New(root)}
	c, _, err := svc.Create("implement demo-spec", "backend", core.WorkStandard, core.IntentExecution)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.Complete(c.TaskID, []string{"internal/core/domain.go updated"})
	if err != nil {
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
	for _, n := range g.Nodes {
		if n.Type == "spec" {
			hasSpec = true
		}
	}
	for _, e := range g.Edges {
		if e.Rel == "assigned_to" {
			hasAssigned = true
		}
	}
	if !hasSpec || !hasAssigned {
		t.Fatalf("expected spec+assigned_to, graph=%+v", g)
	}
}
