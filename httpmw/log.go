package httpmw

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/authenticvision/util-go/httpmw/internal/ddlog"
	"github.com/authenticvision/util-go/httpp"
	"github.com/authenticvision/util-go/logutil"
	"github.com/google/uuid"
)

type User = logutil.UserValue

type accessLogTag struct{}

type accessLog struct {
	SuppressInfoLog bool
	User            *User
}

// WithRequestUser attaches the given user identity to the request's log, and
// additionally adds it to the request's top-level log scope for access logs.
func WithRequestUser(r *http.Request, user User) *http.Request {
	ctx := r.Context()
	if p, ok := ctx.Value(accessLogTag{}).(*accessLog); ok {
		p.User = &user
	}
	log := logutil.FromContext(ctx)
	log = log.With(slog.Any(logutil.UserKey, user))
	return r.WithContext(logutil.WithLogContext(ctx, log))
}

// NewLogMiddleware creates a middleware for recording each request as log line.
// Errors are processed via logutil.Destructure and won't be forwarded.
func NewLogMiddleware(log *slog.Logger) Middleware {
	return &logMiddleware{log: log}
}

// DisableAccessLog suppresses informational access log lines for the request.
// This only affects the application's internal access log.
func DisableAccessLog(r *http.Request) {
	if p, ok := r.Context().Value(accessLogTag{}).(*accessLog); ok {
		p.SuppressInfoLog = true
	}
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
	// public random ID for all log lines of this request, e.g. for use on error screens
	id := uuid.New()
	w.Header().Set("X-Request-Id", id.String())

	// attach logger and extendable scope to context
	var opts accessLog
	ctx := logutil.WithLogContext(r.Context(), ddlog.WithRequest(h.log, r, id))
	ctx = context.WithValue(ctx, accessLogTag{}, &opts)
	r = r.WithContext(ctx)

	// run request
	hookedW := &httpStatusRecorder{ResponseWriter: w}
	start := time.Now()
	err := h.next.ServeErrHTTP(hookedW, r)
	duration := time.Since(start)
	if err != nil {
		httpp.WriteError(hookedW, err)
		// hookedW.statusCode is available now in case of errors
	}

	// attach request+response telemetry
	log := h.log.With(slog.Duration("duration", duration))
	log = ddlog.WithResponse(log, r, id, hookedW)
	if user := opts.User; user != nil {
		log = log.With(slog.Any(logutil.UserKey, *user))
	}

	// attach request error, if any
	level := slog.LevelInfo
	if err != nil {
		var errLeveler slog.Leveler
		if errors.Is(err, context.Canceled) || clientConnDied(r, err) {
			log = log.With(slog.Bool("canceled", true))
			// Context cancellation happens when the browser closes/aborts a connection, which then
			// cascades to any running sub-requests on the server. This includes some error
			// scenarios, like a network-level timeout or I/O error. When such an event
			// occurs during a read or write, said operation may also directly return an error.
			// Regardless, this cascade of errors is intentional, or at least not
			// under our influence, hence always log it with info level.
		} else if errors.As(err, &errLeveler) {
			level = errLeveler.Level()
		} else {
			level = slog.LevelError
		}

		log = log.With(logutil.Err(err))
	}

	if !opts.SuppressInfoLog || level != slog.LevelInfo {
		log.Log(ctx, level, "HTTP request")
	}

	return nil
}

// clientConnDied returns true if err originates from a read/write to the HTTP client.
// The check against prevents accidentally hiding network errors to a backend/database/etc.
// Most commonly this means ECONNRESET ("connection reset by peer") and EPIPE ("broken pipe")
func clientConnDied(r *http.Request, err error) bool {
	var netErr *net.OpError
	if !errors.As(err, &netErr) {
		return false
	}
	if netErr.Op != "read" && netErr.Op != "write" {
		return false
	}
	localAddr, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok || netErr.Source == nil {
		return false
	}
	return netErr.Source.String() == localAddr.String()
}

var _ interface {
	http.ResponseWriter
	httpp.ResponseWriterUnwrapper
	ddlog.HttpStatusRecorder
} = &httpStatusRecorder{}

type httpStatusRecorder struct {
	http.ResponseWriter
	wroteHeader  bool
	statusCode   int
	bytesWritten uint64
}

func (hook *httpStatusRecorder) Unwrap() http.ResponseWriter {
	return hook.ResponseWriter
}

func (hook *httpStatusRecorder) WriteHeader(statusCode int) {
	if !hook.wroteHeader {
		hook.wroteHeader = true
		hook.statusCode = statusCode
	}
	hook.ResponseWriter.WriteHeader(statusCode)
}

func (hook *httpStatusRecorder) Write(b []byte) (int, error) {
	if !hook.wroteHeader {
		hook.WriteHeader(http.StatusOK)
	}
	n, err := hook.ResponseWriter.Write(b)
	hook.bytesWritten += uint64(n)
	return n, err
}

func (hook *httpStatusRecorder) StatusCode() int {
	return hook.statusCode
}

func (hook *httpStatusRecorder) BytesWritten() uint64 {
	return hook.bytesWritten
}
