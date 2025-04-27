package httplog

import (
	"bufio"
	"github.com/authenticvision/util-go/logutil"
	"github.com/google/uuid"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Middleware struct {
	log *slog.Logger
}

func (m *Middleware) Middleware(next http.Handler) http.Handler {
	return Wrap(m.log, next)
}

func NewMiddleware(log *slog.Logger) *Middleware {
	return &Middleware{log: log}
}

func Wrap(log *slog.Logger, next http.Handler) http.Handler {
	return &wrappedHandler{log: log, next: next}
}

type wrappedHandler struct {
	log  *slog.Logger
	next http.Handler
}

func (mid *wrappedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := mid.log.With(slog.String("request_id", uuid.NewString()))
	hook := interceptStatusCode(w)
	now := time.Now()
	mid.next.ServeHTTP(hook, r.WithContext(logutil.WithLogContext(r.Context(), log)))
	duration := time.Since(now)
	log.Info("HTTP request",
		slog.Duration("duration", duration),
		slog.Any("http", makeDatadogHttpValue(r, hook.StatusCode())),
		slog.Any("network", makeDatadogNetworkValue(r)),
	)
}

type ResponseWriterWithStatus interface {
	http.ResponseWriter
	StatusCode() int
}

func interceptStatusCode(w http.ResponseWriter) ResponseWriterWithStatus {
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
	statusCode int
}

func (hook *httpStatusHook) WriteHeader(statusCode int) {
	hook.statusCode = statusCode
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
