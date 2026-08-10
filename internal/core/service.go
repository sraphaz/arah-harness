package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
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

func resultOf(c *Contract, path string, before, after mutateSnapshot, opts MutateOptions, idempotent bool) *MutationResult {
	diff := ""
	if !idempotent {
		diff = formatMutationDiff(before, after)
	}
	return &MutationResult{
		Contract:   c,
		Path:       path,
		Diff:       diff,
		Idempotent: idempotent,
		DryRun:     opts.DryRun || strings.HasPrefix(path, "dry-run"),
	}
}

// Create builds a contract, starts execution, and optionally persists it.
// With DryRun=true the planned contract is returned without Save/emit (path "dry-run").
func (s *TaskService) Create(objective, area string, wc WorkClass, intent IntentType, opts MutateOptions) (*MutationResult, error) {
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
		return nil, err
	}
	c, err := NewContract(objective, area, wc, intent, routing)
	if err != nil {
		return nil, err
	}
	before := mutateSnapshot{} // create is additive from an empty prior
	if err := c.Start(); err != nil {
		return nil, err
	}
	after := snapContract(c)
	if opts.DryRun {
		return resultOf(c, "dry-run", before, after, opts, false), nil
	}
	path, err := s.Store.Save(c)
	if err != nil {
		return nil, wrapStore(err)
	}
	if err := s.emit(c.TaskID, "task.created", map[string]any{
		"primary_executor": c.PrimaryExecutor,
		"state":            string(c.State),
		"area":             area,
	}); err != nil {
		return resultOf(c, path, before, after, opts, false), errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": c.TaskID, "kind": "task.created"})
	}
	if err := s.emit(c.TaskID, "task.started", map[string]any{"state": string(c.State)}); err != nil {
		return resultOf(c, path, before, after, opts, false), errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": c.TaskID, "kind": "task.started"})
	}
	return resultOf(c, path, before, after, opts, false), nil
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
// Re-completing an already-done task with the same evidence is idempotent (no Save/emit).
func (s *TaskService) Complete(taskID string, evidence []string, opts MutateOptions) (*MutationResult, error) {
	load := s.Store.Get
	if opts.DryRun {
		load = s.Store.Peek
	}
	c, path, err := load(taskID)
	if err != nil {
		return nil, err
	}
	before := snapContract(c)
	if c.State == StateDone && evidenceSatisfied(c, evidence) {
		outPath := path
		if opts.DryRun {
			outPath = "dry-run:" + path
		}
		return resultOf(c, outPath, before, before, opts, true), nil
	}
	// Work on a copy so DryRun cannot mutate the store's in-memory view via shared pointers.
	planned := *c
	planned.Execution.CompletionEvidence = append([]string{}, c.Execution.CompletionEvidence...)
	planned.Result.Evidence = append([]string{}, c.Result.Evidence...)
	planned.History = append([]HistoryEntry{}, c.History...)
	if err := planned.Complete(evidence); err != nil {
		return nil, err
	}
	after := snapContract(&planned)
	if opts.DryRun {
		return resultOf(&planned, "dry-run:"+path, before, after, opts, false), nil
	}
	path, err = s.Store.Save(&planned)
	if err != nil {
		return nil, wrapStore(err)
	}
	if err := s.emit(planned.TaskID, "task.completed", map[string]any{
		"evidence": evidence,
		"state":    string(planned.State),
	}); err != nil {
		return resultOf(&planned, path, before, after, opts, false), errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": planned.TaskID})
	}
	return resultOf(&planned, path, before, after, opts, false), nil
}

// Block records a concrete blocking reason and moves the task to blocked.
// Re-blocking with the same reason is idempotent (no Save/emit).
func (s *TaskService) Block(taskID, reason string, opts MutateOptions) (*MutationResult, error) {
	load := s.Store.Get
	if opts.DryRun {
		load = s.Store.Peek
	}
	c, path, err := load(taskID)
	if err != nil {
		return nil, err
	}
	reason = strings.TrimSpace(reason)
	before := snapContract(c)
	if c.State == StateBlocked && c.Result.BlockingReason != nil && *c.Result.BlockingReason == reason && reason != "" {
		outPath := path
		if opts.DryRun {
			outPath = "dry-run:" + path
		}
		return resultOf(c, outPath, before, before, opts, true), nil
	}
	planned := *c
	planned.History = append([]HistoryEntry{}, c.History...)
	if err := planned.Block(reason); err != nil {
		return nil, err
	}
	after := snapContract(&planned)
	if opts.DryRun {
		return resultOf(&planned, "dry-run:"+path, before, after, opts, false), nil
	}
	path, err = s.Store.Save(&planned)
	if err != nil {
		return nil, wrapStore(err)
	}
	if err := s.emit(planned.TaskID, "task.blocked", map[string]any{
		"reason": reason,
		"state":  string(planned.State),
	}); err != nil {
		return resultOf(&planned, path, before, after, opts, false), errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": planned.TaskID})
	}
	return resultOf(&planned, path, before, after, opts, false), nil
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
