package core

// StateStore persists execution contracts (hot state). Hexagonal outbound port.
type StateStore interface {
	EnsureLayout() error
	Save(c *Contract) (path string, err error)
	Get(taskID string) (*Contract, string, error)
	// Peek loads a contract without side effects (no freshness reconcile writes).
	Peek(taskID string) (*Contract, string, error)
	List(bucket string) ([]*Contract, error)
}

// StateDeleter removes a task from persistent state for create-time rollback.
type StateDeleter interface {
	Delete(taskID string) error
}

// Event is an append-only runtime fact used for timelines and Evidence Graph input.
type Event struct {
	ID            string         `json:"id"`
	TaskID        string         `json:"task_id,omitempty"`
	Kind          string         `json:"kind"`
	At            string         `json:"at"`
	Payload       map[string]any `json:"payload,omitempty"`
	TraceID       string         `json:"trace_id,omitempty"`
	RunID         string         `json:"run_id,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	AgentID       string         `json:"agent_id,omitempty"`
	SessionID     string         `json:"session_id,omitempty"`
}

// BriefingWriter persists the deterministic executor briefing artifact (H-13).
type BriefingWriter interface {
	WriteBriefing(c *Contract) (path string, err error)
}

// EventStore is the append-only outbound port for runtime events.
type EventStore interface {
	Append(ev Event) error
	ListByTask(taskID string) ([]Event, error)
	ListRecent(limit int) ([]Event, error)
}

// EventPurger removes all events for a task when create rollback is required.
// It is only safe for brand-new tasks that have no prior timeline.
type EventPurger interface {
	DeleteByTask(taskID string) error
}

// TerminalApplier persists a terminal contract and its mutation event atomically
// when the adapter can (e.g. SQLite transaction). Optional capability on StateStore.
type TerminalApplier interface {
	ApplyTerminal(c *Contract, ev Event) (path string, err error)
}

// ChoreographyResolver selects primary_executor and participants for an area.
type ChoreographyResolver interface {
	Resolve(area, preferredExecutor string) (ResolvedRouting, error)
}
