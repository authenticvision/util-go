package logutil

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"runtime"
	"strings"

	"github.com/lmittmann/tint"
)

const (
	KeyErr   = "error"
	KeyStack = "stack"
)

func ErrColor(value slog.Attr) slog.Attr {
	const ansiRed = 9
	return tint.Attr(ansiRed, value)
}

func Err(err error) slog.Attr {
	return ErrColor(slog.Any(KeyErr, err))
}

type stackValue struct {
	pcs []uintptr
}

// WriteTableTo writes a concise stack trace.
func (s stackValue) WriteTableTo(w io.Writer) error {
	colFrame := digitsInBase10(len(s.pcs))
	colLoc := 0
	frames := runtime.CallersFrames(s.pcs)
	for {
		frame, more := frames.Next()
		width := len(frame.File) + 1 /* ':' */ + digitsInBase10(frame.Line)
		if colLoc < width {
			colLoc = width
		}
		if !more {
			break
		}
	}
	frames = runtime.CallersFrames(s.pcs)
	for i := len(s.pcs); ; i-- {
		frame, more := frames.Next()
		loc := fmt.Sprintf("%s:%d", frame.File, frame.Line)
		_, err := fmt.Fprintf(w, "\t#%-*d %-*s %s +0x%x\n",
			colFrame, i, colLoc, loc, frame.Function, frame.PC-frame.Entry+1)
		if err != nil {
			return fmt.Errorf("write stack frame %d: %w", i, err)
		}
		if !more {
			break
		}
	}
	return nil
}

// WriteStandardTo attempts to mimic the format of debug.Stack() for standard tooling.
// This omits function parameters and the point from which this Goroutine was created, because that
// information is not exposed through standard runtime interfaces.
func (s stackValue) WriteStandardTo(w io.Writer) error {
	_, _ = fmt.Fprintln(w, "goroutine:") // just in case any tooling eats the first line
	frames := runtime.CallersFrames(s.pcs)
	for {
		frame, more := frames.Next()
		_, err := fmt.Fprintf(w, "%s()\n\t%s:%d +0x%x\n",
			frame.Function, frame.File, frame.Line, frame.PC-frame.Entry+1)
		if err != nil {
			return fmt.Errorf("write stack frame: %w", err)
		}
		if !more {
			break
		}
	}
	return nil
}

// MarshalJSON is implemented specifically for slog.JSONHandler, which always calls json.Marshal on
// slog attributes of kind Any. This would otherwise log an empty object (no public values).
func (s stackValue) MarshalJSON() ([]byte, error) {
	w := &strings.Builder{}
	if err := s.WriteStandardTo(w); err != nil {
		return nil, fmt.Errorf("write stack trace: %w", err)
	} else if buf, err := json.Marshal(w.String()); err != nil {
		return nil, fmt.Errorf("marshal stack trace: %w", err)
	} else {
		return buf, nil
	}
}

func Stack(skip int) slog.Attr {
	return slog.Any(KeyStack, stackValue{pcs: fullStack(skip)})
}

func fullStack(skip int) []uintptr {
	depth := 32
	for {
		pc := make([]uintptr, depth)
		n := runtime.Callers(skip+2, pc)
		if n < len(pc) {
			return pc[:n-1] // skips return to goexit
		}
		depth *= 2
	}
}

// digitsInBase10 returns the number of digits in the base-10 representation of n.
func digitsInBase10(n int) int {
	return int(math.Log10(float64(n))) + 1
}
