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
	Store         StateStore
	Events        EventStore // optional; timeline when set
	Router        ChoreographyResolver
	Briefings     BriefingWriter     // optional; writes BRIEFING.md on create
	Consultations ConsultationWriter // optional; structured consultant opinions
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
	return s.emitCorrelated(&Contract{TaskID: taskID}, kind, payload)
}

func (s *TaskService) emitCorrelated(c *Contract, kind string, payload map[string]any) error {
	if s.Events == nil {
		return nil
	}
	ev := Event{
		ID:            newEventID(),
		TaskID:        c.TaskID,
		Kind:          kind,
		At:            time.Now().UTC().Format(time.RFC3339Nano),
		TraceID:       envelope.NewTraceID(),
		Payload:       payload,
		RunID:         runIDFor(c),
		CorrelationID: c.TaskID,
		AgentID:       c.PrimaryExecutor,
	}
	return s.Events.Append(ev)
}

func runIDFor(c *Contract) string {
	if c == nil || c.TaskID == "" {
		return ""
	}
	// Stable run id derived from task id (one run per execution contract in 0.5).
	return "run-" + strings.TrimPrefix(c.TaskID, "task-")
}

func (s *TaskService) mutationEvent(taskID, kind, fingerprint string, payload map[string]any) Event {
	return Event{
		ID:            mutationEventID(taskID, kind, fingerprint),
		TaskID:        taskID,
		Kind:          kind,
		At:            time.Now().UTC().Format(time.RFC3339Nano),
		TraceID:       envelope.NewTraceID(),
		Payload:       payload,
		RunID:         "run-" + strings.TrimPrefix(taskID, "task-"),
		CorrelationID: taskID,
	}
}

// persistTerminal saves the contract and appends the mutation event. When Store
// implements TerminalApplier and is the same object as Events, both happen in
// one adapter transaction.
func (s *TaskService) persistTerminal(c *Contract, ev Event) (string, error) {
	if at, ok := s.Store.(TerminalApplier); ok && s.Events != nil && sameConcrete(s.Store, s.Events) {
		path, err := at.ApplyTerminal(c, ev)
		if err != nil {
			return "", wrapStore(err)
		}
		return path, nil
	}
	path, err := s.Store.Save(c)
	if err != nil {
		return "", wrapStore(err)
	}
	if s.Events != nil {
		if err := s.Events.Append(ev); err != nil {
			return path, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": c.TaskID, "kind": ev.Kind})
		}
	}
	return path, nil
}

// ensureMutationEvent reconciles a missing terminal timeline row on idempotent retry.
func (s *TaskService) ensureMutationEvent(ev Event) error {
	if s.Events == nil {
		return nil
	}
	if err := s.Events.Append(ev); err != nil {
		return errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": ev.TaskID, "kind": ev.Kind})
	}
	return nil
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
	if s.Briefings != nil {
		if _, berr := s.Briefings.WriteBriefing(c); berr != nil {
			s.abortCreate(c.TaskID)
			return nil, errf("STATE.BRIEFING_WRITE_FAILED", berr.Error(), map[string]any{"task_id": c.TaskID})
		}
	}
	if err := s.emitCorrelated(c, "task.created", map[string]any{
		"primary_executor": c.PrimaryExecutor,
		"state":            string(c.State),
		"area":             area,
	}); err != nil {
		s.abortCreate(c.TaskID)
		return nil, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": c.TaskID, "kind": "task.created"})
	}
	if err := s.emitCorrelated(c, "task.started", map[string]any{"state": string(c.State)}); err != nil {
		s.abortCreate(c.TaskID)
		return nil, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": c.TaskID, "kind": "task.started"})
	}
	return resultOf(c, path, before, after, opts, false), nil
}

// abortCreate rolls back a partially persisted create so callers can retry safely.
func (s *TaskService) abortCreate(taskID string) {
	_ = s.Store.Delete(taskID)
}

// Context returns a budgeted progressive-disclosure view of a task.
func (s *TaskService) Context(taskID string, budget ContextBudget) (*TaskContext, error) {
	c, _, err := s.Store.Get(taskID)
	if err != nil {
		return nil, err
	}
	var events []Event
	if s.Events != nil {
		events, _ = s.Events.ListByTask(taskID)
	}
	briefing := ""
	if budget == BudgetFull {
		briefing = RenderBriefing(c)
	}
	return BuildTaskContext(c, events, budget, briefing), nil
}

// ExplainRoute returns the choreography decision for an area (model-callable harness API).
func (s *TaskService) ExplainRoute(area, preferred string) (map[string]any, error) {
	if area == "" {
		area = "backend"
	}
	r, err := s.Router.Resolve(area, preferred)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"area":              area,
		"preferred":         preferred,
		"primary_executor":  r.PrimaryExecutor,
		"choreography_rule": r.ChoreographyRule,
		"consultants":       r.Consultants,
		"reviewers":         r.Reviewers,
		"subordinates":      r.Subordinates,
		"allowed_paths":     r.AllowedPaths,
	}, nil
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
// Re-completing an already-done task with the same evidence is idempotent; a
// missing task.completed event is reconciled before success is returned.
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
	fp := evidenceFingerprint(evidence)
	payload := map[string]any{"evidence": evidence, "state": string(StateDone)}
	ev := s.mutationEvent(taskID, "task.completed", fp, payload)
	ev.AgentID = c.PrimaryExecutor
	if c.State == StateDone && evidenceSameSet(c, evidence) {
		outPath := path
		if opts.DryRun {
			outPath = "dry-run:" + path
		} else if err := s.ensureMutationEvent(ev); err != nil {
			return nil, err
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
	path, err = s.persistTerminal(&planned, ev)
	if err != nil {
		return resultOf(&planned, path, before, after, opts, false), err
	}
	return resultOf(&planned, path, before, after, opts, false), nil
}

// Block records a concrete blocking reason and moves the task to blocked.
// Re-blocking with the same reason is idempotent; a missing task.blocked event
// is reconciled before success is returned.
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
	payload := map[string]any{"reason": reason, "state": string(StateBlocked)}
	ev := s.mutationEvent(taskID, "task.blocked", reason, payload)
	ev.AgentID = c.PrimaryExecutor
	if c.State == StateBlocked && c.Result.BlockingReason != nil && *c.Result.BlockingReason == reason && reason != "" {
		outPath := path
		if opts.DryRun {
			outPath = "dry-run:" + path
		} else if err := s.ensureMutationEvent(ev); err != nil {
			return nil, err
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
	path, err = s.persistTerminal(&planned, ev)
	if err != nil {
		return resultOf(&planned, path, before, after, opts, false), err
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

func sameConcrete(a, b any) bool {
	return a != nil && b != nil && a == b
}
