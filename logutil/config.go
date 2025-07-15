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
	Level  Level  `usage:"TRACE, DEBUG, INFO, WARN, or ERROR" default:"INFO"`
	Format Format `usage:"TEXT or JSON" default:"TEXT"`
}

func (c Config) NewHandler() (slog.Handler, error) {
	return NewHandler(c.Format, c.Level)
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

type ConfigEmbed struct {
	Log Config
}

func (c ConfigEmbed) LogConfig() Config {
	return c.Log
}
