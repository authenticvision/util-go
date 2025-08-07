package logutil

import (
	"errors"
	"github.com/authenticvision/util-go/generic"
	"log/slog"
	"strconv"
	"strings"
)

// RootScope attributes are created as top-level fields in log messages.
var RootScope = Scope{}

// NewScope returns a scope with attributes for later instantiation of a logger or an error.
// The scope's group name is used for grouping the scope's attributes via slog.Group.
func NewScope(group string, attrs ...slog.Attr) *Scope {
	if group == "" {
		panic("logutil.NewScope: a group name is required")
	}
	return &Scope{group: group, attrs: attrs}
}

// Scope collects attributes for later instantiation of a logger or an error.
type Scope struct {
	group string
	attrs []slog.Attr
}

// Sub returns a new Scope with the current scope's group, attributes, plus additional attributes.
func (s *Scope) Sub(attrs ...slog.Attr) *Scope {
	return &Scope{group: s.group, attrs: s.concat(attrs)}
}

// Add the given attributes to the current scope.
func (s *Scope) Add(attrs ...slog.Attr) {
	s.attrs = append(s.attrs, attrs...)
}

// Log returns a slog.Logger with the scope's current attributes, plus additional attributes.
func (s *Scope) Log(log *slog.Logger, attrs ...slog.Attr) *slog.Logger {
	sAttrs := generic.AnySlice(s.concat(attrs))
	if s.group != "" {
		return log.With(slog.Group(s.group, sAttrs...))
	} else {
		return log.With(sAttrs...)
	}
}

// New creates a new error with the given message and attributes.
// The error inherits the scope's current attributes.
func (s *Scope) New(msg string, attrs ...slog.Attr) error {
	return &scopedError{err: errors.New(msg), group: s.group, attrs: s.concat(attrs)}
}

// Err wraps an error to propagate the scope's current attributes, plus additional attributes.
// The inner error should not be nil.
func (s *Scope) Err(err error, msg string, attrs ...slog.Attr) error {
	return &scopedError{err: err, msg: msg, group: s.group, attrs: s.concat(attrs)}
}

func (s *Scope) concat(attrs []slog.Attr) []slog.Attr {
	return append(append([]slog.Attr{}, s.attrs...), attrs...)
}

type scopedError struct {
	err   error
	msg   string
	group string
	attrs []slog.Attr
}

// Error returns one of three formats, depending on whether msg and err are set:
//
//   - <msg> [with <attrib1=value> [<attrib2=value> ...]][: <err>]
//   - <err> [with <attrib1=value> [<attrib2=value> ...]]
//   - scoped error [with <attrib1=value> [<attrib2=value> ...]]
func (e scopedError) Error() string {
	var sb strings.Builder
	if e.msg != "" {
		sb.WriteString(e.msg)
	} else if e.err != nil {
		sb.WriteString(e.err.Error())
	} else {
		sb.WriteString("scoped error")
	}
	if len(e.attrs) > 0 {
		sb.WriteString(" with ")
		for i, attr := range e.attrs {
			if i > 0 {
				sb.WriteRune(' ')
			}
			sb.WriteString(attr.Key)
			sb.WriteRune('=')
			sb.WriteString(strconv.Quote(attr.Value.String()))
		}
	}
	if e.msg != "" && e.err != nil {
		sb.WriteString(": ")
		sb.WriteString(e.err.Error())
	}
	return sb.String()
}

func (e scopedError) Unwrap() error {
	return e.err
}

// Destructure recursively moves all attributes from err to log, and returns the new logger.
// The error chain is modified!
// Errors that unwrap to more than one structure error are unsupported/skipped.
func Destructure(err error, log *slog.Logger) *slog.Logger {
	var attrs []any
	curr := err
	for {
		var sErr *scopedError
		if errors.As(curr, &sErr) {
			sAttrs := generic.AnySlice(sErr.attrs)
			if sErr.group != "" {
				attrs = append(attrs, slog.Group(sErr.group, sAttrs...))
			} else {
				attrs = append(attrs, sAttrs...)
			}
			sErr.attrs = nil
			curr = sErr.Unwrap()
		} else {
			break
		}
	}
	attrs = append(attrs, Err(err))
	return log.With(attrs...)
}
