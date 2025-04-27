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
	return NewHandler(c.Format, slog.Level(c.Level))
}

func (c Config) InstallForProcess() error {
	handler, err := c.NewHandler()
	if err != nil {
		return fmt.Errorf("failed to create log handler: %w", err)
	}
	slog.SetDefault(slog.New(handler))
	InstallGoLogShim()
	return nil
}
