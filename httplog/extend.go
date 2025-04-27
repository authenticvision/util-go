package httplog

import (
	"context"
	"log/slog"
)

// LogExtender enables HTTP handlers to inject fields into the HTTP log.
type LogExtender interface {
	Log() *slog.Logger
	Extend(log *slog.Logger)
}

type logExtenderTagType int

var logExtenderTag logExtenderTagType

func withLogExtender(c context.Context, le *logExtender) context.Context {
	return context.WithValue(c, logExtenderTag, le)
}

func FromContext(c context.Context) LogExtender {
	if result := c.Value(logExtenderTag); result != nil {
		return result.(LogExtender)
	} else {
		return nil
	}
}

type logExtender struct {
	log *slog.Logger
}

func (ext *logExtender) Log() *slog.Logger {
	return ext.log
}

func (ext *logExtender) Extend(log *slog.Logger) {
	ext.log = log
}
