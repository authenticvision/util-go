package httpmw

import (
	"context"
	"errors"
	"github.com/authenticvision/util-go/httpp"
	"github.com/authenticvision/util-go/logutil"
	"github.com/google/uuid"
	"log/slog"
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
	hookedW := &httpStatusRecorder{ResponseWriter: w}
	log := h.log.With(slog.String("request_id", uuid.NewString()))
	ctx := logutil.WithLogContext(r.Context(), log)
	r = r.WithContext(ctx)

	start := time.Now()
	err := h.next.ServeErrHTTP(hookedW, r)
	duration := time.Since(start)
	if err != nil {
		httpp.WriteError(hookedW, err)
		// hookedW.statusCode is available now in case of errors
	}

	log = log.With(
		slog.Duration("duration", duration),
		slog.Any("http", makeDatadogHttpValue(r, hookedW.statusCode)),
		slog.Any("network", makeDatadogNetworkValue(r)))

	level := slog.LevelInfo
	if err != nil {
		var errLeveler slog.Leveler
		if errors.Is(err, context.Canceled) {
			log = log.With(slog.Bool("canceled", true))
			// Context cancellation happens when the browser closes/aborts a connection, which then
			// cascades to any running sub-requests on the server. This includes some error
			// scenarios, like a network-level timeout or I/O error. Regardless, log this cascade
			// of errors is intentional, hence always log it with info level.
		} else if errors.As(err, &errLeveler) {
			level = errLeveler.Level()
		} else {
			level = slog.LevelError
		}

		log = logutil.Destructure(err, log)
	}

	log.Log(ctx, level, "HTTP request")

	return nil
}

var _ interface {
	http.ResponseWriter
	httpp.ResponseWriterUnwrapper
} = &httpStatusRecorder{}

type httpStatusRecorder struct {
	http.ResponseWriter
	wroteHeader bool
	statusCode  int
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
	return hook.ResponseWriter.Write(b)
}
