package util

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_DelayedCancel(t *testing.T) {
	r := require.New(t)
	baseCtx, baseCancel := context.WithCancel(context.Background())
	baseCancel()
	ctx, cancel := DelayedCancel(baseCtx, 100*time.Millisecond)
	defer cancel()
	time.Sleep(50 * time.Millisecond)
	r.Error(baseCtx.Err(), context.Canceled)
	r.NoError(ctx.Err())
	time.Sleep(100 * time.Millisecond)
	r.Error(ctx.Err(), context.Canceled)
}
