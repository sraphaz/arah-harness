package sqlitestore_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
)

func TestSchemaUpgradeAndRollback(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, ".arah", "local", "runtime.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatal(err)
	}

	// Seed a v1 database (baseline only — no kind index).
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
CREATE TABLE schema_meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE tasks (
  task_id TEXT PRIMARY KEY,
  state TEXT NOT NULL,
  bucket TEXT NOT NULL,
  primary_executor TEXT,
  objective TEXT,
  work_class TEXT,
  intent_type TEXT,
  choreography_rule TEXT,
  contract_yaml TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX idx_tasks_bucket ON tasks(bucket);
CREATE TABLE task_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_id TEXT NOT NULL UNIQUE,
  task_id TEXT,
  kind TEXT NOT NULL,
  at TEXT NOT NULL,
  trace_id TEXT,
  payload_json TEXT NOT NULL
);
CREATE INDEX idx_events_task ON task_events(task_id);
INSERT INTO schema_meta(key, value) VALUES('version', '1');
`)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	v, err := store.SchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 {
		t.Fatalf("expected upgrade to v2, got %d", v)
	}
	if !indexExists(t, store.DBPath, "idx_events_kind") {
		t.Fatal("expected idx_events_kind after upgrade")
	}

	// Round-trip a terminal write on the upgraded schema.
	c := &core.Contract{
		Version: "1.0", TaskID: "task-mig", Objective: "schema",
		WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		State: core.StateExecuting, PrimaryExecutor: "backend",
	}
	if _, err := store.Save(c); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(core.Event{ID: "ev-1", TaskID: c.TaskID, Kind: "task.created", Payload: map[string]any{}}); err != nil {
		t.Fatal(err)
	}

	if err := store.RollbackTo(1); err != nil {
		t.Fatal(err)
	}
	v, err = store.SchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 {
		t.Fatalf("expected rollback to v1, got %d", v)
	}
	if indexExists(t, store.DBPath, "idx_events_kind") {
		t.Fatal("idx_events_kind should be dropped on rollback to v1")
	}
	// Data preserved.
	got, _, err := store.Get("task-mig")
	if err != nil {
		t.Fatal(err)
	}
	if got.Objective != "schema" {
		t.Fatalf("data lost on rollback: %+v", got)
	}

	// Re-open upgrades again (idempotent forward migrate).
	_ = store.Close()
	store2, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store2.Close()
	v, err = store2.SchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 {
		t.Fatalf("re-open should upgrade to v2, got %d", v)
	}
}

func TestFreshStoreIsLatestSchema(t *testing.T) {
	root := t.TempDir()
	store, err := sqlitestore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	v, err := store.SchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 {
		t.Fatalf("fresh store version=%d", v)
	}
	if !indexExists(t, store.DBPath, "idx_events_kind") {
		t.Fatal("fresh store missing v2 index")
	}
}

func indexExists(t *testing.T, dbPath, name string) bool {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	err = db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type='index' AND name=?`, name).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n > 0
}
