// Package ports re-exports arah-core ports for adapters that prefer this import path.
// Prefer importing github.com/sraphaz/arah-harness/internal/core directly.
package ports

import "github.com/sraphaz/arah-harness/internal/core"

// StateStore is an alias of core.StateStore.
type StateStore = core.StateStore

// ChoreographyResolver is an alias of core.ChoreographyResolver.
type ChoreographyResolver = core.ChoreographyResolver

// EventStore is an alias of core.EventStore.
type EventStore = core.EventStore
