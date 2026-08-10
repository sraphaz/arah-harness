// Package ports re-exports arah-core ports for adapters that prefer this path.
// Prefer importing github.com/sraphaz/arah-harness/internal/core directly.
package ports

import "github.com/sraphaz/arah-harness/internal/core"

type StateStore = core.StateStore
type ChoreographyResolver = core.ChoreographyResolver
