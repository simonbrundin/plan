package usage

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Global usage tracker
var (
	globalTracker     *Tracker
	globalTrackerOnce sync.Once
)

// SetGlobalTracker sets the global usage tracker
func SetGlobalTracker(t *Tracker) {
	globalTrackerOnce.Do(func() {
		globalTracker = t
	})
}

// GetGlobalTracker returns the global usage tracker
func GetGlobalTracker() *Tracker {
	return globalTracker
}

// RecordLLMUsage is a convenience function to record LLM usage globally
func RecordLLMUsage(ctx context.Context, usage TokenUsage) error {
	if globalTracker == nil {
		return nil
	}
	return globalTracker.Record(ctx, usage)
}

// InitializeGlobalTracker creates a global tracker from a database pool
func InitializeGlobalTracker(pool *pgxpool.Pool) *Tracker {
	tracker := NewTracker(pool)
	SetGlobalTracker(tracker)
	return tracker
}
