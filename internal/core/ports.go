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

// Event is an append-only runtime fact used for timelines and Evidence Graph input.
type Event struct {
	ID      string         `json:"id"`
	TaskID  string         `json:"task_id,omitempty"`
	Kind    string         `json:"kind"`
	At      string         `json:"at"`
	Payload map[string]any `json:"payload,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
}

// EventStore is the append-only outbound port for runtime events.
type EventStore interface {
	Append(ev Event) error
	ListByTask(taskID string) ([]Event, error)
	ListRecent(limit int) ([]Event, error)
}

// ChoreographyResolver selects primary_executor and participants for an area.
type ChoreographyResolver interface {
	Resolve(area, preferredExecutor string) (ResolvedRouting, error)
}
