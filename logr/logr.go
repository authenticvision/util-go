// Package logr contains convenience functions for accessing a http.Request's logger.
package logr

import (
	"context"
	"github.com/authenticvision/util-go/logutil"
	"log/slog"
	"net/http"
)

func Enabled(r *http.Request, level slog.Level) bool {
	ctx := r.Context()
	log := logutil.FromContext(ctx)
	return log.Enabled(ctx, level)
}

func Log(r *http.Request, level slog.Level, msg string, attrs ...slog.Attr) {
	ctx := r.Context()
	log := logutil.FromContext(ctx)
	log.LogAttrs(ctx, level, msg, attrs...)
}

func Trace(r *http.Request, msg string, attrs ...slog.Attr) {
	Log(r, logutil.LevelTrace, msg, attrs...)
}

func Debug(r *http.Request, msg string, attrs ...slog.Attr) {
	Log(r, slog.LevelDebug, msg, attrs...)
}

func Info(r *http.Request, msg string, attrs ...slog.Attr) {
	Log(r, slog.LevelInfo, msg, attrs...)
}

func Warn(r *http.Request, msg string, attrs ...slog.Attr) {
	Log(r, slog.LevelWarn, msg, attrs...)
}

func Error(r *http.Request, msg string, attrs ...slog.Attr) {
	Log(r, slog.LevelError, msg, attrs...)
}

func New(r *http.Request, attrs ...any) *Logger {
	ctx := r.Context()
	log := logutil.FromContext(ctx).With(attrs...)
	return &Logger{ctx: ctx, log: log}
}

func FromScope(r *http.Request, scope *logutil.Scope, attrs ...slog.Attr) *Logger {
	ctx := r.Context()
	log := scope.Log(logutil.FromContext(ctx), attrs...)
	return &Logger{ctx: ctx, log: log}
}

type Logger struct {
	ctx context.Context
	log *slog.Logger
}

func (l *Logger) With(attrs ...any) *Logger {
	return &Logger{ctx: l.ctx, log: l.log.With(attrs...)}
}

func (l *Logger) Enabled(level slog.Level) bool {
	return l.log.Enabled(l.ctx, level)
}

func (l *Logger) Log(level slog.Level, msg string, attrs ...slog.Attr) {
	l.log.LogAttrs(l.ctx, level, msg, attrs...)
}

func (l *Logger) Trace(msg string, attrs ...slog.Attr) {
	l.Log(logutil.LevelTrace, msg, attrs...)
}

func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	l.Log(slog.LevelDebug, msg, attrs...)
}

func (l *Logger) Info(msg string, attrs ...slog.Attr) {
	l.Log(slog.LevelInfo, msg, attrs...)
}

func (l *Logger) Warn(msg string, attrs ...slog.Attr) {
	l.Log(slog.LevelWarn, msg, attrs...)
}

func (l *Logger) Error(msg string, attrs ...slog.Attr) {
	l.Log(slog.LevelError, msg, attrs...)
}
