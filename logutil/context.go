package logutil

import (
	"context"
	"log/slog"
	"net/http"
)

type logContextKey struct{}

func WithLogContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, logContextKey{}, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(logContextKey{}).(*slog.Logger)
	if !ok {
		slog.Warn("logctx.FromContext: no logger in context")
		return slog.Default()
	}
	return log
}

func LoggerMiddleware(next http.Handler, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := WithLogContext(r.Context(), log)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
