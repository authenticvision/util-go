package logutil

import (
	"context"
	"log/slog"
)

type logContextKey struct{}

func WithLogContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, logContextKey{}, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(logContextKey{}).(*slog.Logger)
	if !ok {
		slog.Warn("logutil.FromContext: no logger in context")
		return slog.Default()
	}
	return log
}
