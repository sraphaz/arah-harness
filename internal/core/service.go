package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sraphaz/arah-harness/internal/envelope"
)

// TaskService is the application use-case layer over StateStore + choreography.
// CLI and MCP must call the same methods so decisions stay identical.
type TaskService struct {
	Store  StateStore
	Events EventStore // optional; timeline when set
	Router ChoreographyResolver
}

// MutateOptions controls plan → validate → apply behaviour (H-14).
type MutateOptions struct {
	// DryRun plans and validates without persisting state or emitting events.
	DryRun bool
}

func newEventID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("ev-%d-%s", time.Now().UnixNano(), hex.EncodeToString(b[:]))
}

func (s *TaskService) emit(taskID, kind string, payload map[string]any) error {
	if s.Events == nil {
		return nil
	}
	return s.Events.Append(Event{
		ID:      newEventID(),
		TaskID:  taskID,
		Kind:    kind,
		At:      time.Now().UTC().Format(time.RFC3339Nano),
		TraceID: envelope.NewTraceID(),
		Payload: payload,
	})
}

// Create builds a contract, starts execution, and optionally persists it.
// With DryRun=true the planned contract is returned without Save/emit (path "dry-run").
func (s *TaskService) Create(objective, area string, wc WorkClass, intent IntentType, opts MutateOptions) (*Contract, string, error) {
	if area == "" {
		area = "backend"
	}
	if wc == "" {
		wc = WorkStandard
	}
	if intent == "" {
		intent = IntentExecution
	}
	routing, err := s.Router.Resolve(area, "")
	if err != nil {
		return nil, "", err
	}
	c, err := NewContract(objective, area, wc, intent, routing)
	if err != nil {
		return nil, "", err
	}
	if err := c.Start(); err != nil {
		return nil, "", err
	}
	if opts.DryRun {
		return c, "dry-run", nil
	}
	path, err := s.Store.Save(c)
	if err != nil {
		return nil, "", wrapStore(err)
	}
	if err := s.emit(c.TaskID, "task.created", map[string]any{
		"primary_executor": c.PrimaryExecutor,
		"state":            string(c.State),
		"area":             area,
	}); err != nil {
		return c, path, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": c.TaskID, "kind": "task.created"})
	}
	if err := s.emit(c.TaskID, "task.started", map[string]any{"state": string(c.State)}); err != nil {
		return c, path, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": c.TaskID, "kind": "task.started"})
	}
	return c, path, nil
}

// Get loads a contract by task_id from the StateStore.
func (s *TaskService) Get(taskID string) (*Contract, string, error) {
	c, path, err := s.Store.Get(taskID)
	if err != nil {
		return nil, "", err
	}
	return c, path, nil
}

// Complete validates concrete evidence and transitions the task to done.
func (s *TaskService) Complete(taskID string, evidence []string, opts MutateOptions) (*Contract, string, error) {
	load := s.Store.Get
	if opts.DryRun {
		load = s.Store.Peek
	}
	c, path, err := load(taskID)
	if err != nil {
		return nil, "", err
	}
	// Work on a copy so DryRun cannot mutate the store's in-memory view via shared pointers.
	planned := *c
	planned.Execution.CompletionEvidence = append([]string{}, c.Execution.CompletionEvidence...)
	planned.Result.Evidence = append([]string{}, c.Result.Evidence...)
	planned.History = append([]HistoryEntry{}, c.History...)
	if err := planned.Complete(evidence); err != nil {
		return nil, "", err
	}
	if opts.DryRun {
		return &planned, "dry-run:" + path, nil
	}
	path, err = s.Store.Save(&planned)
	if err != nil {
		return nil, "", wrapStore(err)
	}
	if err := s.emit(planned.TaskID, "task.completed", map[string]any{
		"evidence": evidence,
		"state":    string(planned.State),
	}); err != nil {
		return &planned, path, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": planned.TaskID})
	}
	return &planned, path, nil
}

// Block records a concrete blocking reason and moves the task to blocked.
func (s *TaskService) Block(taskID, reason string, opts MutateOptions) (*Contract, string, error) {
	load := s.Store.Get
	if opts.DryRun {
		load = s.Store.Peek
	}
	c, path, err := load(taskID)
	if err != nil {
		return nil, "", err
	}
	planned := *c
	planned.History = append([]HistoryEntry{}, c.History...)
	if err := planned.Block(reason); err != nil {
		return nil, "", err
	}
	if opts.DryRun {
		return &planned, "dry-run:" + path, nil
	}
	path, err = s.Store.Save(&planned)
	if err != nil {
		return nil, "", wrapStore(err)
	}
	if err := s.emit(planned.TaskID, "task.blocked", map[string]any{
		"reason": reason,
		"state":  string(planned.State),
	}); err != nil {
		return &planned, path, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": planned.TaskID})
	}
	return &planned, path, nil
}

// Timeline returns append-only events for an existing task.
func (s *TaskService) Timeline(taskID string) ([]Event, error) {
	if s.Events == nil {
		return nil, errf("STATE.EVENT_STORE_UNAVAILABLE", "event store not configured", nil)
	}
	if _, _, err := s.Store.Get(taskID); err != nil {
		return nil, err
	}
	return s.Events.ListByTask(taskID)
}

func wrapStore(err error) error {
	return errf("STATE.STORE_ERROR", err.Error(), nil)
}
