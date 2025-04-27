package logutil

import (
	logpkg "log"
	"log/slog"
	"strings"
)

type LogWriterShim struct {
	l *slog.Logger
}

func (l *LogWriterShim) Write(p []byte) (n int, err error) {
	// http.Server.ErrorLog needs slog integration still :/
	s := string(p[:len(p)-1]) // strip newline
	level := slog.LevelInfo
	if strings.HasPrefix(s, "http: TLS") && strings.HasSuffix(s, ": EOF") {
		// libavas will establish TCP connections to multiple AVAS instances, or the same AVAS
		// instance over IPv4 and IPv6, and use whatever connected the quickest and completes a
		// session init. It'll just close() the remainder, which Go writes a log message for. :/
		// See also: https://github.com/golang/go/issues/26918
		level = LevelTrace
	}
	l.l.Log(nil, level, s)
	return len(p), nil
}

func InstallGoLogShim() {
	logpkg.SetOutput(&LogWriterShim{l: slog.With("module", "stdlib_log")})
}
