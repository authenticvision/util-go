package logutil

import (
	"log/slog"
	"strings"
)

type Level slog.Level

func (l *Level) UnmarshalText(text []byte) error {
	name := strings.ToUpper(string(text))
	return (*slog.Level)(l).UnmarshalText([]byte(name))
}

func (l *Level) String() string {
	return (*slog.Level)(l).String()
}

func (l *Level) CmdTypeDesc() string {
	return "level"
}
