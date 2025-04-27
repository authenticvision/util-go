package logutil

import (
	"log/slog"
	"strings"
)

type Level slog.Level

func (l *Level) UnmarshalText(text []byte) error {
	name := strings.ToUpper(string(text))
	if level, ok := NameLevels[name]; ok {
		*l = Level(level)
		return nil
	} else {
		return (*slog.Level)(l).UnmarshalText([]byte(name))
	}
}

func (l *Level) String() string {
	if name, ok := LevelNames[slog.Level(*l)]; ok {
		return name
	} else {
		return (*slog.Level)(l).String()
	}
}

func (l *Level) CmdTypeDesc() string {
	return "level"
}

var (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

var LevelNames = map[slog.Level]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

var NameLevels = func() map[string]slog.Level {
	m := make(map[string]slog.Level, len(LevelNames))
	for l, n := range LevelNames {
		m[n] = l
	}
	if len(LevelNames) != len(m) {
		panic("duplicate level value or name")
	}
	return m
}()

func LevelAttrReplacer(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		if level, isLevel := a.Value.Any().(slog.Level); isLevel {
			if name, ok := LevelNames[level]; ok {
				a.Value = slog.StringValue(name)
			}
		}
	}
	return a
}
