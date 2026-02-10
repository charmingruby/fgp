// Package timeutil hosts internal timing helpers shared across packages.
//
// Example:
//
//	timeutil.Sleep(ctx, 10*time.Millisecond)
package timeutil

import (
	"context"
	"time"
)

// Sleep waits for the provided duration or until the context is done. It
// returns true when the full duration elapsed, or false when the context was
// canceled.
//
// Example:
//
//	ok := Sleep(ctx, time.Second)
//	if !ok {
//		return ctx.Err()
//	}
func Sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
