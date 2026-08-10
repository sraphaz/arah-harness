package core

import (
	"fmt"
	"time"

	"github.com/sraphaz/arah-harness/internal/envelope"
)

// TaskService is the application use-case layer over StateStore + choreography.
type TaskService struct {
	Store  StateStore
	Events EventStore // optional; timeline when set
	Router ChoreographyResolver
}

func (s *TaskService) emit(taskID, kind string, payload map[string]any) {
	if s.Events == nil {
		return
	}
	_ = s.Events.Append(Event{
		ID:      fmt.Sprintf("ev-%d", time.Now().UnixNano()),
		TaskID:  taskID,
		Kind:    kind,
		At:      time.Now().UTC().Format(time.RFC3339Nano),
		TraceID: envelope.NewTraceID(),
		Payload: payload,
	})
}

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
	s.emit(c.TaskID, "task.created", map[string]any{
		"primary_executor": c.PrimaryExecutor,
		"state":            string(c.State),
		"area":             area,
	})
	s.emit(c.TaskID, "task.started", map[string]any{"state": string(c.State)})
	return c, path, nil
}

func (s *TaskService) Get(taskID string) (*Contract, string, error) {
	c, path, err := s.Store.Get(taskID)
	if err != nil {
		return nil, "", err
	}
	return c, path, nil
}

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
	s.emit(c.TaskID, "task.completed", map[string]any{
		"evidence": evidence,
		"state":    string(c.State),
	})
	return c, path, nil
}

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
	s.emit(c.TaskID, "task.blocked", map[string]any{
		"reason": reason,
		"state":  string(c.State),
	})
	return c, path, nil
}

func (s *TaskService) Timeline(taskID string) ([]Event, error) {
	if s.Events == nil {
		return nil, errf("STATE.EVENT_STORE_UNAVAILABLE", "event store not configured", nil)
	}
	return s.Events.ListByTask(taskID)
}

func wrapStore(err error) error {
	return errf("STATE.STORE_ERROR", err.Error(), nil)
}
