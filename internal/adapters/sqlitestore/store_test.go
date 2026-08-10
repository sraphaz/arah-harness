package sqlitestore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
)

func TestSQLiteLifecycleAndMigration(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644)

	// Seed a filesystem-only contract to migrate.
	fs := fsstore.New(root)
	seed := &core.Contract{
		Version: "1.0", TaskID: "task-seed-fs", Objective: "seed",
		WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		State: core.StateExecuting, PrimaryExecutor: "backend",
		Execution: core.Execution{}, Result: core.Result{},
	}
	if _, err := fs.Save(seed); err != nil {
		t.Fatal(err)
	}

	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	got, ref, err := store.Get("task-seed-fs")
	if err != nil {
		t.Fatal(err)
	}
	if got.Objective != "seed" {
		t.Fatalf("objective=%s", got.Objective)
	}
	if ref == "" || ref[:7] != "sqlite:" {
		t.Fatalf("ref=%s", ref)
	}

	svc := &core.TaskService{
		Store:  store,
		Events: store,
		Router: choreography.New(root),
	}
	c, _, err := svc.Create("sqlite path", "backend", core.WorkStandard, core.IntentExecution)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".arah", "local", "runtime.db")); err != nil {
		t.Fatal("runtime.db missing")
	}
	c2, _, err := svc.Complete(c.TaskID, []string{"internal/adapters/sqlitestore/store.go updated"})
	if err != nil {
		t.Fatal(err)
	}
	if c2.State != core.StateDone {
		t.Fatalf("state=%s", c2.State)
	}
	evs, err := store.ListByTask(c.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) < 2 {
		t.Fatalf("expected timeline events, got %d", len(evs))
	}
}
