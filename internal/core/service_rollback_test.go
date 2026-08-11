package core_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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
	inner     core.EventStore
	failOn    string
	appended  []string
}

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

type failDeleteStore struct {
	core.StateStore
	err error
}

func (f failDeleteStore) Delete(taskID string) error {
	return f.err
}

func TestCreateSurfacesAbortFailure(t *testing.T) {
	_, svc, store := sqliteSvc(t)
	svc.Briefings = failBriefing{}
	svc.Store = failDeleteStore{StateStore: store, err: fmt.Errorf("delete refused")}
	res, err := svc.Create("abort-fail", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err == nil || res != nil {
		t.Fatalf("expected abort failure, got res=%v err=%v", res, err)
	}
	de := err.(*core.DomainError)
	if de.Code != "STATE.CREATE_ABORT_FAILED" {
		t.Fatalf("code=%s msg=%s", de.Code, de.Message)
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
	for _, ev := range evs {
		if ev.Kind == "consultation.submitted" && ev.AgentID != "solutions-architect" {
			t.Fatalf("consultation event must attribute consultant, agent_id=%s", ev.AgentID)
		}
	}
}

func TestSubmitConsultationConcurrentRespectsLimit(t *testing.T) {
	_, svc, _ := sqliteSvc(t)
	created, err := svc.Create("concurrent consult", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// WorkStandard MaxConsultations=1
	id := created.Contract.TaskID
	var (
		wg       sync.WaitGroup
		okCount  atomic.Int32
		errCount atomic.Int32
	)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _, err := svc.SubmitConsultation(id, "solutions-architect", fmt.Sprintf("opinion %d", n), nil, nil)
			if err == nil {
				okCount.Add(1)
			} else {
				errCount.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if okCount.Load() != 1 {
		t.Fatalf("expected exactly 1 success, got ok=%d err=%d", okCount.Load(), errCount.Load())
	}
	got, _, err := svc.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Counters.Consultations != 1 {
		t.Fatalf("counter=%d", got.Counters.Consultations)
	}
}

type boomWriteConsult struct{}

func (boomWriteConsult) WriteConsultation(string, *core.ConsultationResult) (string, error) {
	return "", fmt.Errorf("write boom")
}
func (boomWriteConsult) RemoveConsultation(string) error { return nil }

type flakyReleaseStore struct {
	*sqlitestore.Store
}

func (f *flakyReleaseStore) ReleaseConsultationSlot(taskID string) (*core.Contract, error) {
	return nil, fmt.Errorf("release boom")
}

type boomRemoveConsult struct {
	inner core.ConsultationWriter
}

func (b boomRemoveConsult) WriteConsultation(taskID string, result *core.ConsultationResult) (string, error) {
	return b.inner.WriteConsultation(taskID, result)
}
func (b boomRemoveConsult) RemoveConsultation(string) error {
	return fmt.Errorf("remove boom")
}

func TestSubmitConsultationSurfacesCompensationFailures(t *testing.T) {
	root, svc, store := sqliteSvc(t)
	created, err := svc.Create("comp fail", "backend", core.WorkArchitectural, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Contract.TaskID

	svc.Store = &flakyReleaseStore{Store: store}
	svc.Consultations = boomWriteConsult{}
	_, _, err = svc.SubmitConsultation(id, "solutions-architect", "x", nil, nil)
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "STATE.CONSULTATION_ROLLBACK_FAILED" {
		t.Fatalf("write+rollback path: %#v", err)
	}

	// Fresh service for emit+remove compensation path.
	fs := fsstore.New(root)
	svc2 := &core.TaskService{
		Store: store, Events: &failEvents{inner: store, failOn: "consultation.submitted"},
		Router: choreography.New(root), Briefings: fs, Consultations: boomRemoveConsult{inner: fs},
	}
	_, _, err = svc2.SubmitConsultation(id, "solutions-architect", "y", nil, nil)
	de, ok = err.(*core.DomainError)
	if !ok || de.Code != "STATE.CONSULTATION_ROLLBACK_FAILED" {
		t.Fatalf("emit+remove path: %#v", err)
	}
}
