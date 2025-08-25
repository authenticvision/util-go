package logutil

import (
	"log/slog"

	"github.com/lmittmann/tint"
)

const ErrKey = "error"

func ErrColor(value slog.Attr) slog.Attr {
	const ansiRed = 9
	return tint.Attr(ansiRed, value)
}

func Err(err error) slog.Attr {
	return ErrColor(slog.Any(ErrKey, err))
}
