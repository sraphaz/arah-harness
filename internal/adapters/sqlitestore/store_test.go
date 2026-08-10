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

func TestListMergesFilesystemOnlyTasks(t *testing.T) {
	root := t.TempDir()
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	fsOnly := &core.Contract{
		Version: "1.0", TaskID: "task-fs-only", Objective: "ps writer",
		WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		State: core.StateExecuting, PrimaryExecutor: "backend",
	}
	if _, err := fsstore.New(root).Save(fsOnly); err != nil {
		t.Fatal(err)
	}
	list, err := store.List("active")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range list {
		if c.TaskID == "task-fs-only" {
			found = true
		}
	}
	if !found {
		t.Fatal("filesystem-only task missing from List")
	}
}

func TestTimelineRequiresExistingTask(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644)
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	svc := &core.TaskService{Store: store, Events: store, Router: choreography.New(root)}
	_, err = svc.Timeline("task-does-not-exist")
	if err == nil {
		t.Fatal("expected TASK_NOT_FOUND")
	}
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "EXECUTION.TASK_NOT_FOUND" {
		t.Fatalf("got %#v", err)
	}
}

func TestImportPropagatesListErrors(t *testing.T) {
	root := t.TempDir()
	fs := fsstore.New(root)
	seed := &core.Contract{
		Version: "1.0", TaskID: "task-seed", Objective: "seed",
		WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		State: core.StateExecuting, PrimaryExecutor: "backend",
	}
	if _, err := fs.Save(seed); err != nil {
		t.Fatal(err)
	}
	active := filepath.Join(root, ".arah", "local", "execution", "active")
	if err := os.Chmod(active, 0o000); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(active, 0o755) }()

	_, err := sqlitestore.New(root)
	if err == nil {
		t.Fatal("expected migration to fail when filesystem List fails")
	}
}

func TestGetReconcilesFresherFilesystemState(t *testing.T) {
	root := t.TempDir()
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	executing := &core.Contract{
		Version: "1.0", TaskID: "task-ps-complete", Objective: "dual write",
		WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		State: core.StateExecuting, PrimaryExecutor: "backend",
		History: []core.HistoryEntry{{At: "2026-08-10T01:00:00Z", From: "routed", To: "executing"}},
	}
	if _, err := store.Save(executing); err != nil {
		t.Fatal(err)
	}

	// PowerShell-style update: only YAML mirror moves to done with evidence.
	done := *executing
	done.State = core.StateDone
	done.Execution.CompletionEvidence = []string{"scripts/agents/task-control.ps1 completed"}
	done.History = append(append([]core.HistoryEntry{}, executing.History...), core.HistoryEntry{
		At: "2026-08-10T02:00:00Z", From: "executing", To: "done", Note: "ps complete",
	})
	if _, err := fsstore.New(root).Save(&done); err != nil {
		t.Fatal(err)
	}

	got, _, err := store.Get("task-ps-complete")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != core.StateDone {
		t.Fatalf("expected reconciled done state, got %s", got.State)
	}
	if len(got.Execution.CompletionEvidence) == 0 {
		t.Fatal("expected completion evidence from YAML mirror")
	}

	// SQLite should have been updated so a subsequent Get stays fresh.
	got2, _, err := store.Get("task-ps-complete")
	if err != nil {
		t.Fatal(err)
	}
	if got2.State != core.StateDone {
		t.Fatalf("sqlite still stale: %s", got2.State)
	}
}

