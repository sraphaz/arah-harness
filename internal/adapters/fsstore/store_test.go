package fsstore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/core"
)

func TestTaskLifecycle(t *testing.T) {
	root := t.TempDir()
	agents := filepath.Join(root, ".agents")
	if err := os.MkdirAll(agents, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agents, "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	svc := &core.TaskService{
		Store:  fsstore.New(root),
		Router: choreography.New(root),
	}
	c, path, err := svc.Create("implement foo", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if c.State != core.StateExecuting {
		t.Fatalf("state=%s", c.State)
	}
	if c.PrimaryExecutor != "backend" {
		t.Fatalf("executor=%s", c.PrimaryExecutor)
	}
	if path == "" {
		t.Fatal("empty path")
	}

	if _, _, err := svc.Complete(c.TaskID, nil, core.MutateOptions{}); err == nil {
		t.Fatal("expected evidence error")
	}

	c2, _, err := svc.Complete(c.TaskID, []string{"internal/core/domain.go updated"}, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if c2.State != core.StateDone {
		t.Fatalf("state=%s", c2.State)
	}

	got, _, err := svc.Get(c.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != core.StateDone {
		t.Fatalf("persisted state=%s", got.State)
	}
}

func TestBlockPersists(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644)
	svc := &core.TaskService{Store: fsstore.New(root), Router: choreography.New(root)}
	c, _, err := svc.Create("blocked work", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	c2, path, err := svc.Block(c.TaskID, "missing credential X", core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if c2.State != core.StateBlocked {
		t.Fatalf("state=%s", c2.State)
	}
	if !filepath.IsAbs(path) || filepath.Base(filepath.Dir(path)) != "blocked" {
		t.Fatalf("expected blocked bucket path, got %s", path)
	}
}

func TestSaveKeepsPriorBucketOnWriteFailure(t *testing.T) {
	root := t.TempDir()
	store := fsstore.New(root)
	c := &core.Contract{
		Version: "1.0", TaskID: "task-keep-active", Objective: "keep",
		WorkClass: core.WorkStandard, IntentType: core.IntentExecution,
		State: core.StateExecuting, PrimaryExecutor: "backend",
	}
	if _, err := store.Save(c); err != nil {
		t.Fatal(err)
	}
	active := filepath.Join(root, ".arah", "local", "execution", "active", c.TaskID+".yaml")
	if _, err := os.Stat(active); err != nil {
		t.Fatal("active missing before failed save")
	}

	completed := filepath.Join(root, ".arah", "local", "execution", "completed")
	if err := os.Chmod(completed, 0o555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(completed, 0o755) }()

	c.State = core.StateDone
	if _, err := store.Save(c); err == nil {
		t.Fatal("expected write failure into read-only completed bucket")
	}
	got, path, err := store.Get(c.TaskID)
	if err != nil {
		t.Fatalf("task lost after failed Save: %v", err)
	}
	if got.State != core.StateExecuting {
		t.Fatalf("expected prior executing contract, got %s at %s", got.State, path)
	}
	if filepath.Base(filepath.Dir(path)) != "active" {
		t.Fatalf("expected active bucket, got %s", path)
	}
}
