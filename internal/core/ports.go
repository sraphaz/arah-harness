package core

// StateStore persists execution contracts (hot state). Hexagonal outbound port.
type StateStore interface {
	EnsureLayout() error
	Save(c *Contract) (path string, err error)
	Get(taskID string) (*Contract, string, error)
	List(bucket string) ([]*Contract, error)
}

// ChoreographyResolver selects primary_executor and participants for an area.
type ChoreographyResolver interface {
	Resolve(area, preferredExecutor string) (ResolvedRouting, error)
}
