package logutil

import (
	"fmt"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"log/slog"
	"os"
	"strings"
)

type Format string

func (f *Format) UnmarshalText(text []byte) error {
	format := Format(strings.ToUpper(string(text)))
	switch format {
	case FormatText, FormatJSON:
		*f = format
		return nil
	}
	return fmt.Errorf("invalid log format text: %q", text)
}

func (f *Format) String() string {
	return string(*f)
}

func (f *Format) CmdTypeDesc() string {
	return "format"
}

const (
	FormatText Format = "TEXT"
	FormatJSON Format = "JSON"
)

func NewHandler(format Format, level Level) (slog.Handler, error) {
	out := os.Stderr
	switch format {
	case FormatText:
		return tint.NewHandler(out, &tint.Options{
			Level:   slog.Level(level),
			NoColor: !isatty.IsTerminal(out.Fd()),
		}), nil
	case FormatJSON:
		return slog.NewJSONHandler(out, &slog.HandlerOptions{
			Level: slog.Level(level),
		}), nil
	}
	return nil, fmt.Errorf("unsupported log format: %s", format)
}
