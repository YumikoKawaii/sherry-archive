package tracking

import "context"

// Enricher is an optional hook called after events are persisted.
// Implementations update real-time analytics stores (e.g. Redis).
// Errors are silently dropped — analytics must never block ingestion.
type Enricher interface {
	ProcessEvents(ctx context.Context, events []EventRow)
}
