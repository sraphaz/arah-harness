package core_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
)

type failBriefing struct {
	inner core.BriefingWriter
}

func (f failBriefing) WriteBriefing(c *core.Contract) (string, error) {
	if f.inner != nil {
		_, _ = f.inner.WriteBriefing(c)
	}
	return "", fmt.Errorf("briefing boom")
}

type failEvents struct {
	inner    core.EventStore
	failOn   string
	appended []string
}

type deleteFailStore struct {
	inner     core.StateStore
	deleteErr error
}

func (s deleteFailStore) EnsureLayout() error { return s.inner.EnsureLayout() }
func (s deleteFailStore) Save(c *core.Contract) (string, error) {
	return s.inner.Save(c)
}
func (s deleteFailStore) Get(taskID string) (*core.Contract, string, error) {
	return s.inner.Get(taskID)
}
func (s deleteFailStore) Peek(taskID string) (*core.Contract, string, error) {
	return s.inner.Peek(taskID)
}
func (s deleteFailStore) List(bucket string) ([]*core.Contract, error) {
	return s.inner.List(bucket)
}
func (s deleteFailStore) Delete(taskID string) error { return s.deleteErr }

func (f *failEvents) Append(ev core.Event) error {
	f.appended = append(f.appended, ev.Kind)
	if f.failOn != "" && ev.Kind == f.failOn {
		return fmt.Errorf("append boom: %s", ev.Kind)
	}
	if f.inner != nil {
		return f.inner.Append(ev)
	}
	return nil
}

func (f *failEvents) ListByTask(taskID string) ([]core.Event, error) {
	if f.inner != nil {
		return f.inner.ListByTask(taskID)
	}
	return nil, nil
}

func (f *failEvents) ListRecent(limit int) ([]core.Event, error) {
	if f.inner != nil {
		return f.inner.ListRecent(limit)
	}
	return nil, nil
}

func sqliteSvc(t *testing.T) (string, *core.TaskService, *sqlitestore.Store) {
	t.Helper()
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte(`
version: 2
rules:
  - id: craft-backend
    paths: ["backend/**"]
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
	fs := fsstore.New(root)
	svc := &core.TaskService{
		Store: store, Events: store, Router: choreography.New(root),
		Briefings: fs, Consultations: fs,
	}
	return root, svc, store
}

func TestCreateRollsBackWhenBriefingFails(t *testing.T) {
	root, svc, _ := sqliteSvc(t)
	fs := fsstore.New(root)
	svc.Briefings = failBriefing{inner: fs}
	res, err := svc.Create("partial", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err == nil || res != nil {
		t.Fatalf("expected failure without result, got res=%v err=%v", res, err)
	}
	de := err.(*core.DomainError)
	if de.Code != "STATE.BRIEFING_WRITE_FAILED" {
		t.Fatalf("code=%s", de.Code)
	}
	list, err := svc.Store.List("active")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("zombie tasks: %d", len(list))
	}
}

func TestCreateRollsBackWhenStartedEventFails(t *testing.T) {
	_, svc, store := sqliteSvc(t)
	svc.Events = &failEvents{inner: store, failOn: "task.started"}
	res, err := svc.Create("partial-events", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err == nil || res != nil {
		t.Fatalf("expected failure without result, got res=%v err=%v", res, err)
	}
	list, _ := svc.Store.List("active")
	if len(list) != 0 {
		t.Fatalf("zombie tasks after event failure: %d", len(list))
	}
}

func TestCreateReturnsRollbackFailureWhenDeleteFails(t *testing.T) {
	_, svc, store := sqliteSvc(t)
	svc.Store = deleteFailStore{inner: store, deleteErr: fmt.Errorf("delete boom")}
	svc.Briefings = failBriefing{}

	res, err := svc.Create("rollback fails", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err == nil || res != nil {
		t.Fatalf("expected failure without result, got res=%v err=%v", res, err)
	}
	de, ok := err.(*core.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "STATE.STORE_ERROR" {
		t.Fatalf("code=%s", de.Code)
	}
	if got := de.Details["create_stage"]; got != "STATE.BRIEFING_WRITE_FAILED" {
		t.Fatalf("create_stage=%v", got)
	}
	if got := de.Details["rollback_error"]; got != "delete boom" {
		t.Fatalf("rollback_error=%v", got)
	}
	list, err := store.List("active")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected persisted zombie to be visible after rollback failure, got %d", len(list))
	}
}

func TestSubmitConsultationEmitsOrRollsBack(t *testing.T) {
	_, svc, store := sqliteSvc(t)
	created, err := svc.Create("consult me", "backend", core.WorkArchitectural, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Contract.TaskID

	fe := &failEvents{inner: store, failOn: "consultation.submitted"}
	svc.Events = fe
	_, path, err := svc.SubmitConsultation(id, "solutions-architect", "looks good", nil, nil)
	if err == nil {
		t.Fatal("expected emit failure")
	}
	if path != "" {
		t.Fatalf("path should be cleared on rollback, got %s", path)
	}
	got, _, err := svc.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Counters.Consultations != 0 {
		t.Fatalf("counter not rolled back: %d", got.Counters.Consultations)
	}

	svc.Events = store
	_, path, err = svc.SubmitConsultation(id, "solutions-architect", "approved", []string{"ship"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	evs, err := svc.Timeline(id)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ev := range evs {
		if ev.Kind == "consultation.submitted" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing consultation.submitted in %#v", evs)
	}
}
