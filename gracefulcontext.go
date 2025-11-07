package util

import (
	"context"
	"time"
)

// DelayedCancel returns a context that gets canceled after a delay when its
// parent context is canceled. The returned context.CancelFunc immediately
// cancels the context and should be used like context.WithCancel's one.
func DelayedCancel(ctx context.Context, delay time.Duration) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	context.AfterFunc(ctx, func() {
		t := time.AfterFunc(delay, cancel)
		context.AfterFunc(newCtx, func() {
			t.Stop()
		})
	})
	return newCtx, cancel
}
