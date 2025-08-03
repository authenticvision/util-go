package httpmw

import (
	"bufio"
	"errors"
	"github.com/authenticvision/util-go/httpp"
	"github.com/authenticvision/util-go/logutil"
	"github.com/google/uuid"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// NewLogMiddleware creates a middleware for recording each request as log line.
// Errors are processed via logutil.Destructure and won't be forwarded.
func NewLogMiddleware(log *slog.Logger) Middleware {
	return &logMiddleware{log: log}
}

type logMiddleware struct {
	log *slog.Logger
}

func (m *logMiddleware) Middleware(next httpp.Handler) httpp.Handler {
	return &logHandler{log: m.log, next: next}
}

type logHandler struct {
	log  *slog.Logger
	next httpp.Handler
}

func (h *logHandler) ServeErrHTTP(w http.ResponseWriter, r *http.Request) error {
	hookedW := interceptStatusCode(w)
	log := h.log.With(slog.String("request_id", uuid.NewString()))
	ctx := logutil.WithLogContext(r.Context(), log)
	r = r.WithContext(ctx)

	start := time.Now()
	err := h.next.ServeErrHTTP(hookedW, r)
	duration := time.Since(start)

	log = log.With(
		slog.Duration("duration", duration),
		slog.Any("http", makeDatadogHttpValue(r, hookedW.StatusCode())),
		slog.Any("network", makeDatadogNetworkValue(r)))

	level := slog.LevelInfo
	if err != nil {
		httpp.WriteError(hookedW, err)

		var errLeveler slog.Leveler
		if errors.As(err, &errLeveler) {
			level = errLeveler.Level()
		} else {
			level = slog.LevelError
		}

		log = logutil.Destructure(err, log)
	}

	log.Log(ctx, level, "HTTP request")

	return nil
}

type httpStatusRecorder interface {
	http.ResponseWriter
	StatusCode() int
}

func interceptStatusCode(w http.ResponseWriter) httpStatusRecorder {
	hook := &httpStatusHook{ResponseWriter: w}
	if _, ok := w.(http.Hijacker); ok {
		// for WebSocket support
		return &httpStatusHookHijackable{httpStatusHook: hook}
	} else {
		return hook
	}
}

type httpStatusHook struct {
	http.ResponseWriter
	wroteHeader bool
	statusCode  int
}

func (hook *httpStatusHook) WriteHeader(statusCode int) {
	if !hook.wroteHeader {
		hook.wroteHeader = true
		hook.statusCode = statusCode
	}
	hook.ResponseWriter.WriteHeader(statusCode)
}

func (hook *httpStatusHook) StatusCode() int {
	if hook.statusCode != 0 {
		return hook.statusCode
	} else {
		// implicit behavior of Go's ResponseWriter
		return http.StatusOK
	}
}

type httpStatusHookHijackable struct {
	*httpStatusHook
}

func (hook *httpStatusHookHijackable) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return hook.ResponseWriter.(http.Hijacker).Hijack()
}
