package core_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/core"
)

type staticResolver struct{}

func (staticResolver) Resolve(area, preferred string) (core.ResolvedRouting, error) {
	return core.ResolvedRouting{
		PrimaryExecutor:  "backend",
		ChoreographyRule: "craft-backend",
		Consultants:      []string{"solutions-architect"},
		AllowedPaths:     []string{"internal/**"},
	}, nil
}

type failingBriefings struct{}

func (failingBriefings) WriteBriefing(*core.Contract) (string, error) {
	return "", errors.New("briefing write failed")
}

type failingEventStore struct {
	failAt  int
	appends int
	events  []core.Event
}

func (s *failingEventStore) Append(ev core.Event) error {
	s.appends++
	if s.failAt > 0 && s.appends == s.failAt {
		return errors.New("event append failed")
	}
	s.events = append(s.events, ev)
	return nil
}

func (s *failingEventStore) ListByTask(taskID string) ([]core.Event, error) {
	var out []core.Event
	for _, ev := range s.events {
		if ev.TaskID == taskID {
			out = append(out, ev)
		}
	}
	return out, nil
}

func (s *failingEventStore) ListRecent(limit int) ([]core.Event, error) {
	if limit <= 0 || len(s.events) <= limit {
		return append([]core.Event{}, s.events...), nil
	}
	return append([]core.Event{}, s.events[len(s.events)-limit:]...), nil
}

func (s *failingEventStore) DeleteByTask(taskID string) error {
	filtered := s.events[:0]
	for _, ev := range s.events {
		if ev.TaskID != taskID {
			filtered = append(filtered, ev)
		}
	}
	s.events = filtered
	return nil
}

type failOnNthSaveStore struct {
	inner  *fsstore.Store
	failAt int
	saves  int
}

func (s *failOnNthSaveStore) EnsureLayout() error {
	return s.inner.EnsureLayout()
}

func (s *failOnNthSaveStore) Save(c *core.Contract) (string, error) {
	s.saves++
	if s.failAt > 0 && s.saves == s.failAt {
		return "", errors.New("save failed")
	}
	return s.inner.Save(c)
}

func (s *failOnNthSaveStore) Get(taskID string) (*core.Contract, string, error) {
	return s.inner.Get(taskID)
}

func (s *failOnNthSaveStore) Peek(taskID string) (*core.Contract, string, error) {
	return s.inner.Peek(taskID)
}

func (s *failOnNthSaveStore) List(bucket string) ([]*core.Contract, error) {
	return s.inner.List(bucket)
}

func (s *failOnNthSaveStore) Delete(taskID string) error {
	return s.inner.Delete(taskID)
}

func TestCreateRollsBackWhenBriefingWriteFails(t *testing.T) {
	root := t.TempDir()
	store := fsstore.New(root)
	svc := &core.TaskService{
		Store:     store,
		Router:    staticResolver{},
		Briefings: failingBriefings{},
	}

	res, err := svc.Create("briefing rollback", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err == nil {
		t.Fatal("expected create error")
	}
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "STATE.BRIEFING_WRITE_FAILED" {
		t.Fatalf("unexpected error %#v", err)
	}
	assertTaskMissing(t, store, res.Contract.TaskID)
}

func TestCreateRollsBackWhenSecondEventAppendFails(t *testing.T) {
	root := t.TempDir()
	store := fsstore.New(root)
	events := &failingEventStore{failAt: 2}
	svc := &core.TaskService{
		Store:  store,
		Events: events,
		Router: staticResolver{},
	}

	res, err := svc.Create("event rollback", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err == nil {
		t.Fatal("expected create error")
	}
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "STATE.EVENT_APPEND_FAILED" {
		t.Fatalf("unexpected error %#v", err)
	}
	assertTaskMissing(t, store, res.Contract.TaskID)
	if len(events.events) != 0 {
		t.Fatalf("expected rollback to purge events, got %d", len(events.events))
	}
}

func TestSubmitConsultationRemovesFileWhenSaveFails(t *testing.T) {
	root := t.TempDir()
	baseStore := fsstore.New(root)
	store := &failOnNthSaveStore{inner: baseStore, failAt: 2}
	svc := &core.TaskService{
		Store:         store,
		Router:        staticResolver{},
		Consultations: baseStore,
	}

	created, err := svc.Create("consultation save failure", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.SubmitConsultation(created.Contract.TaskID, "solutions-architect", "summary", []string{"do x"}, nil)
	if err == nil {
		t.Fatal("expected consultation error")
	}
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "STATE.STORE_ERROR" {
		t.Fatalf("unexpected error %#v", err)
	}
	assertNoConsultationFiles(t, root, created.Contract.TaskID)
	saved, _, err := baseStore.Get(created.Contract.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Counters.Consultations != 0 {
		t.Fatalf("consultation counter=%d want 0", saved.Counters.Consultations)
	}
}

func TestSubmitConsultationRollsBackWhenEventAppendFails(t *testing.T) {
	root := t.TempDir()
	store := fsstore.New(root)
	svc := &core.TaskService{
		Store:         store,
		Router:        staticResolver{},
		Consultations: store,
	}

	created, err := svc.Create("consultation event failure", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	svc.Events = &failingEventStore{failAt: 1}

	_, _, err = svc.SubmitConsultation(created.Contract.TaskID, "solutions-architect", "summary", []string{"do x"}, nil)
	if err == nil {
		t.Fatal("expected consultation error")
	}
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "STATE.EVENT_APPEND_FAILED" {
		t.Fatalf("unexpected error %#v", err)
	}
	assertNoConsultationFiles(t, root, created.Contract.TaskID)
	saved, _, err := store.Get(created.Contract.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Counters.Consultations != 0 {
		t.Fatalf("consultation counter=%d want 0", saved.Counters.Consultations)
	}
}

func assertTaskMissing(t *testing.T, store core.StateStore, taskID string) {
	t.Helper()
	_, _, err := store.Get(taskID)
	if err == nil {
		t.Fatalf("task %s still persisted", taskID)
	}
	de, ok := err.(*core.DomainError)
	if !ok || de.Code != "EXECUTION.TASK_NOT_FOUND" {
		t.Fatalf("unexpected get error %#v", err)
	}
}

func assertNoConsultationFiles(t *testing.T, root, taskID string) {
	t.Helper()
	dir := filepath.Join(root, ".arah", "local", "execution", taskID, "consultations")
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no consultation files, got %d", len(entries))
	}
}
