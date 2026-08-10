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

// Create builds a contract, starts execution, persists it, and records timeline events.
func (s *TaskService) Create(objective, area string, wc WorkClass, intent IntentType) (*Contract, string, error) {
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
func (s *TaskService) Complete(taskID string, evidence []string) (*Contract, string, error) {
	c, _, err := s.Store.Get(taskID)
	if err != nil {
		return nil, "", err
	}
	if err := c.Complete(evidence); err != nil {
		return nil, "", err
	}
	path, err := s.Store.Save(c)
	if err != nil {
		return nil, "", wrapStore(err)
	}
	if err := s.emit(c.TaskID, "task.completed", map[string]any{
		"evidence": evidence,
		"state":    string(c.State),
	}); err != nil {
		return c, path, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": c.TaskID})
	}
	return c, path, nil
}

// Block records a concrete blocking reason and moves the task to blocked.
func (s *TaskService) Block(taskID, reason string) (*Contract, string, error) {
	c, _, err := s.Store.Get(taskID)
	if err != nil {
		return nil, "", err
	}
	if err := c.Block(reason); err != nil {
		return nil, "", err
	}
	path, err := s.Store.Save(c)
	if err != nil {
		return nil, "", wrapStore(err)
	}
	if err := s.emit(c.TaskID, "task.blocked", map[string]any{
		"reason": reason,
		"state":  string(c.State),
	}); err != nil {
		return c, path, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": c.TaskID})
	}
	return c, path, nil
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
