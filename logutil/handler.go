package logutil

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func NewHandler(format Format, level slog.Level) (slog.Handler, error) {
	f := os.Stderr
	switch format {
	case FormatText:
		color := isatty.IsTerminal(f.Fd())
		w := &termOutputWrapper{next: f}
		return tint.NewHandler(w, &tint.Options{
			Level:       level,
			ReplaceAttr: w.attrReplacer,
			NoColor:     !color,
		}), nil
	case FormatJSON:
		return slog.NewJSONHandler(f, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: LevelAttrReplacer,
		}), nil
	}
	return nil, fmt.Errorf("unsupported log format: %s", format)
}

// MustNewHandler forwards to NewHandler and panics when it fails.
// It is still used by AVAS, can eventually go away for mainutil.
func MustNewHandler(format Format, level slog.Level) slog.Handler {
	h, err := NewHandler(format, level)
	if err != nil {
		panic(err)
	}
	return h
}

var attrLevelTerminal = map[slog.Level]struct {
	ansiColor uint8
	label     string
}{
	LevelTrace: {ansiColor: 13, label: "TRC"},
	LevelFatal: {ansiColor: 9, label: "FTL"},
}

// termOutputWrapper prints stack traces with line breaks after their log message
type termOutputWrapper struct {
	next  io.Writer
	stack *stackValue
}

func (w *termOutputWrapper) Write(p []byte) (int, error) {
	n, err := w.next.Write(p)
	if err != nil {
		return n, err
	}
	if len(p) == 0 {
		return n, nil
	}
	if p[len(p)-1] == '\n' && w.stack != nil {
		if err := w.stack.WriteTableTo(w.next); err != nil {
			return n, fmt.Errorf("write stack trace: %w", err)
		}
		w.stack = nil
	}
	return n, nil
}

func (w *termOutputWrapper) attrReplacer(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.LevelKey:
		if level, ok := a.Value.Any().(slog.Level); ok {
			if term, ok := attrLevelTerminal[level]; ok {
				a.Value = slog.StringValue(term.label)
				a = tint.Attr(term.ansiColor, a)
			}
		}
	case StackKey:
		if stack, ok := a.Value.Any().(stackValue); ok {
			w.stack = &stack
			return slog.Attr{} // drop from message
		}
	}
	return a
}
