package logutil

import (
	"fmt"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"log/slog"
	"os"
)

func NewHandler(format Format, level Level) (slog.Handler, error) {
	out := os.Stderr
	switch format {
	case FormatText:
		return tint.NewHandler(out, &tint.Options{
			Level:       slog.Level(level),
			ReplaceAttr: LevelAttrReplacer,
			NoColor:     !isatty.IsTerminal(out.Fd()),
		}), nil
	case FormatJSON:
		return slog.NewJSONHandler(out, &slog.HandlerOptions{
			Level:       slog.Level(level),
			ReplaceAttr: LevelAttrReplacer,
		}), nil
	}
	return nil, fmt.Errorf("unsupported log format: %s", format)
}

func MustNewHandler(format Format, level Level) slog.Handler {
	h, err := NewHandler(format, level)
	if err != nil {
		panic(err)
	}
	return h
}
