package logutil

import (
	"fmt"
	"log/slog"
)

var DefaultConfig = Config{
	Level:  Level(slog.LevelInfo),
	Format: FormatText,
}

type Config struct {
	Level  Level  `usage:"TRACE, DEBUG, INFO, WARN, or ERROR"`
	Format Format `usage:"TEXT or JSON"`
}

func (c Config) NewHandler() (slog.Handler, error) {
	return NewHandler(c.Format, c.Level)
}
