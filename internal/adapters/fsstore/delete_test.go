package fsstore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/core"
)

func TestDeleteRemovesMirrorAndArtifacts(t *testing.T) {
	root := t.TempDir()
	s := fsstore.New(root)
	c := &core.Contract{
		TaskID: "task-del", State: core.StateExecuting, PrimaryExecutor: "backend",
		Objective: "x", WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
	}
	if _, err := s.Save(c); err != nil {
		t.Fatal(err)
	}
	if _, err := s.WriteBriefing(c); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(c.TaskID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Get(c.TaskID); err == nil {
		t.Fatal("expected task gone after Delete")
	}
	if _, err := os.Stat(filepath.Join(root, ".arah", "local", "execution", c.TaskID)); !os.IsNotExist(err) {
		t.Fatalf("artifact dir should be gone, err=%v", err)
	}
	// Idempotent: second delete of missing paths is ok.
	if err := s.Delete(c.TaskID); err != nil {
		t.Fatal(err)
	}
}
