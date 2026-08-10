package sqlitestore

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/core"
)

func TestDeleteRestoresFilesystemMirrorWhenSQLiteDeleteFails(t *testing.T) {
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatal(err)
	}

	contract := &core.Contract{
		Version:         "1.0",
		TaskID:          "task-delete-restore",
		Objective:       "keep mirror on sqlite delete failure",
		WorkClass:       core.WorkStandard,
		IntentType:      core.IntentExecution,
		State:           core.StateExecuting,
		PrimaryExecutor: "backend",
	}
	if _, err := store.Save(contract); err != nil {
		t.Fatal(err)
	}
	if _, err := store.fs.WriteBriefing(contract); err != nil {
		t.Fatal(err)
	}

	roDB, err := sql.Open("sqlite", "file:"+store.DBPath+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	oldDB := store.db
	store.db = roDB
	if err := oldDB.Close(); err != nil {
		_ = roDB.Close()
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	if err := store.Delete(contract.TaskID); err == nil {
		t.Fatal("expected delete failure")
	}

	if _, _, err := store.fs.Get(contract.TaskID); err != nil {
		t.Fatalf("filesystem mirror not restored: %v", err)
	}
	briefing := filepath.Join(root, ".arah", "local", "execution", contract.TaskID, "BRIEFING.md")
	if _, err := os.Stat(briefing); err != nil {
		t.Fatalf("briefing not restored: %v", err)
	}
}
