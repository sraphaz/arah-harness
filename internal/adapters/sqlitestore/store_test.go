package sqlitestore_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

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
	created, err := svc.Create("sqlite path", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	c := created.Contract
	if _, err := os.Stat(filepath.Join(root, ".arah", "local", "runtime.db")); err != nil {
		t.Fatal("runtime.db missing")
	}
	done, err := svc.Complete(c.TaskID, []string{"internal/adapters/sqlitestore/store.go updated"}, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	c2 := done.Contract
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

func sqliteTaskState(t *testing.T, dbPath, taskID string) string {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var state string
	if err := db.QueryRow(`SELECT state FROM tasks WHERE task_id = ?`, taskID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	return state
}

func TestPeekAndDryRunDoNotReconcileSQLite(t *testing.T) {
	root := t.TempDir()
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	executing := &core.Contract{
		Version: "1.0", TaskID: "task-dry-run-peek", Objective: "no reconcile on dry-run",
		WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		State: core.StateExecuting, PrimaryExecutor: "backend",
		History: []core.HistoryEntry{{At: "2026-08-10T01:00:00Z", From: "routed", To: "executing"}},
	}
	if _, err := store.Save(executing); err != nil {
		t.Fatal(err)
	}

	// Filesystem is ahead (PS complete); SQLite row stays executing until Get reconciles.
	done := *executing
	done.State = core.StateDone
	done.Execution.CompletionEvidence = []string{"scripts/agents/task-control.ps1 completed"}
	done.History = append(append([]core.HistoryEntry{}, executing.History...), core.HistoryEntry{
		At: "2026-08-10T02:00:00Z", From: "executing", To: "done", Note: "ps complete",
	})
	if _, err := fsstore.New(root).Save(&done); err != nil {
		t.Fatal(err)
	}

	peeked, _, err := store.Peek("task-dry-run-peek")
	if err != nil {
		t.Fatal(err)
	}
	if peeked.State != core.StateDone {
		t.Fatalf("peek should see fresher FS state, got %s", peeked.State)
	}
	if got := sqliteTaskState(t, store.DBPath, "task-dry-run-peek"); got != string(core.StateExecuting) {
		t.Fatalf("peek must not write reconcile; sqlite state=%s", got)
	}

	svc := &core.TaskService{Store: store}
	_, err = svc.Complete("task-dry-run-peek", []string{"would complete"}, core.MutateOptions{DryRun: true})
	if err == nil {
		t.Fatal("dry-run complete on terminal FS state should fail without mutating sqlite")
	}
	if got := sqliteTaskState(t, store.DBPath, "task-dry-run-peek"); got != string(core.StateExecuting) {
		t.Fatalf("dry-run complete must leave sqlite untouched; state=%s", got)
	}
}

func TestIdempotentCompleteReconcilesMissingEvent(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644)
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	svc := &core.TaskService{Store: store, Events: store, Router: choreography.New(root)}
	created, err := svc.Create("reconcile event", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	ev := []string{"internal/adapters/sqlitestore/store.go updated"}
	// Simulate Save-without-event: mark done in store only.
	done := *created.Contract
	if err := done.Complete(ev); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(&done); err != nil {
		t.Fatal(err)
	}
	before, err := store.ListByTask(created.Contract.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	completedBefore := 0
	for _, e := range before {
		if e.Kind == "task.completed" {
			completedBefore++
		}
	}
	if completedBefore != 0 {
		t.Fatalf("setup expected no completed event, got %d", completedBefore)
	}

	res, err := svc.Complete(created.Contract.TaskID, ev, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Idempotent {
		t.Fatal("expected idempotent reconcile path")
	}
	after, err := store.ListByTask(created.Contract.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	completed := 0
	for _, e := range after {
		if e.Kind == "task.completed" {
			completed++
		}
	}
	if completed != 1 {
		t.Fatalf("expected exactly one task.completed after reconcile, got %d", completed)
	}
	// Second retry must not duplicate the event.
	res2, err := svc.Complete(created.Contract.TaskID, ev, core.MutateOptions{})
	if err != nil || !res2.Idempotent {
		t.Fatalf("second retry: err=%v idempotent=%v", err, res2 != nil && res2.Idempotent)
	}
	after2, err := store.ListByTask(created.Contract.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	completed = 0
	for _, e := range after2 {
		if e.Kind == "task.completed" {
			completed++
		}
	}
	if completed != 1 {
		t.Fatalf("duplicate completed events: %d", completed)
	}
}

func TestReserveAndReleaseConsultationSlot(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644)
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	svc := &core.TaskService{Store: store, Events: store, Router: choreography.New(root)}
	created, err := svc.Create("slot", "backend", core.WorkArchitectural, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Contract.TaskID
	max := created.Contract.Limits.MaxConsultations

	c1, err := store.ReserveConsultationSlot(id, max)
	if err != nil {
		t.Fatal(err)
	}
	if c1.Counters.Consultations != 1 {
		t.Fatalf("after reserve: %d", c1.Counters.Consultations)
	}

	// Peer process bumps counter + mutates objective while first holds a stale snapshot.
	peer, _, err := store.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	peer.Counters.Consultations = 2
	peer.Objective = "peer-updated"
	if _, err := store.Save(peer); err != nil {
		t.Fatal(err)
	}

	released, err := store.ReleaseConsultationSlot(id)
	if err != nil {
		t.Fatal(err)
	}
	if released.Counters.Consultations != 1 {
		t.Fatalf("release must decrement fresh counter 2→1, got %d", released.Counters.Consultations)
	}
	if released.Objective != "peer-updated" {
		t.Fatalf("release must not clobber peer fields; objective=%s", released.Objective)
	}
}

