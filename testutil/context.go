package testutil

import (
	"context"
	"log/slog"
	"testing"

	"github.com/authenticvision/util-go/logutil"
)

type contextConfig struct {
	logLevel slog.Level
}

type ContextOption func(config *contextConfig)

func WithLogLevel(level slog.Level) ContextOption {
	return func(config *contextConfig) {
		config.logLevel = level
	}
}

var Trace = WithLogLevel(logutil.LevelTrace)

// Context provides a logger for t.Context(). It might eventually provide
// additional context values offered by util-go, so all code that uses util-go
// context-specific code should get a context from this method in tests.
func Context(t *testing.T, opts ...ContextOption) context.Context {
	conf := contextConfig{
		logLevel: slog.LevelDebug,
	}
	for _, opt := range opts {
		opt(&conf)
	}

	logHandler, err := logutil.NewHandlerTo(t.Output(), logutil.FormatText, conf.logLevel)
	if err != nil {
		t.Fatalf("failed to create log handler: %v", err)
		return nil // unreachable
	}

	log := slog.New(logHandler)
	ctx := logutil.WithLogContext(t.Context(), log)

	return ctx
}
