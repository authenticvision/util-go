package logutil

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func NewHandler(format Format, level slog.Level) (slog.Handler, error) {
	return NewHandlerTo(os.Stderr, format, level)
}

func NewHandlerTo(w io.Writer, format Format, level slog.Level) (slog.Handler, error) {
	var handler slog.Handler
	switch format {
	case FormatText:
		color := false
		if f, ok := w.(*os.File); ok {
			color = isatty.IsTerminal(f.Fd())
		}
		wrapper := &termOutputWrapper{next: w}
		handler = tint.NewHandler(wrapper, &tint.Options{
			Level:       level,
			ReplaceAttr: wrapper.attrReplacer,
			NoColor:     !color,
		})
	case FormatJSON:
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: LevelAttrReplacer,
		})
	default:
		return nil, fmt.Errorf("unsupported log format: %s", format)
	}
	return &scopedErrorHandler{next: handler}, nil
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

// scopedErrorHandler extends log messages by scopedError attributes of an error under ErrKey.
type scopedErrorHandler struct {
	next slog.Handler
}

func (d *scopedErrorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.next.Enabled(ctx, level)
}

func (d *scopedErrorHandler) Handle(ctx context.Context, record slog.Record) error {
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == ErrKey {
			value := attr.Value
			if value.Kind() == slog.KindLogValuer {
				value = value.LogValuer().LogValue()
			}
			if err, ok := value.Any().(error); ok {
				attrs := Destructure(err)
				if len(attrs) > 0 {
					record = record.Clone()
					record.AddAttrs(attrs...)
				}
				return false // stop
			}
		}
		return true // next key
	})
	return d.next.Handle(ctx, record)
}

func (d *scopedErrorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &scopedErrorHandler{next: d.next.WithAttrs(attrs)}
}

func (d *scopedErrorHandler) WithGroup(name string) slog.Handler {
	return &scopedErrorHandler{next: d.next.WithGroup(name)}
}
