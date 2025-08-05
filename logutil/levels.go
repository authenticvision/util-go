package logutil

import (
	"github.com/BooleanCat/go-functional/v2/it/op"
	"github.com/spf13/pflag"
	"log/slog"
	"strings"
)

var _ pflag.Value = op.Ref(Level(0))

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

func (l *Level) Set(s string) error {
	return l.UnmarshalText([]byte(s))
}

func (l *Level) Type() string {
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
		if level, ok := a.Value.Any().(slog.Level); ok {
			if label, ok := LevelNames[level]; ok {
				a.Value = slog.StringValue(label)
			}
		}
	}
	return a
}

// Severity attaches a log level to an error chain.
func Severity(err error, level slog.Level) error {
	if err == nil {
		panic("logutil.Severity: err must not be nil")
	}
	return &severityError{err: err, level: level}
}

var _ slog.Leveler = &severityError{}

type severityError struct {
	err   error
	level slog.Level
}

func (e severityError) Error() string {
	return e.err.Error()
}

func (e severityError) Unwrap() error {
	return e.err
}

func (e severityError) Level() slog.Level {
	return e.level
}
