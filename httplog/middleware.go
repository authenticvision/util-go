package httplog

import (
	"bufio"
	"github.com/authenticvision/util-go/logutil"
	"log/slog"
	"net"
	"net/http"
	"time"
)

func Middleware(log *slog.Logger, next http.Handler) http.Handler {
	return &middleware{
		log:  log,
		next: next,
	}
}

type middleware struct {
	log  *slog.Logger
	next http.Handler
}

type datadogLogHttpRequest struct {
	Host           string               `json:"host"`
	Proto          string               `json:"proto"`
	Method         string               `json:"method"`
	StatusCategory string               `json:"status_category"`
	StatusCode     int                  `json:"status_code"`
	URLDetails     datadogLogUrlDetails `json:"url_details"`
}

type datadogLogUrlDetails struct {
	Path string `json:"path"`
}

type datadogLogHttpClient struct {
	IP   string `json:"ip"`
	Port string `json:"port,omitempty"`
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	le := &logExtender{log: m.log}
	ctx := logutil.WithLogContext(withLogExtender(r.Context(), le), le.Log())
	hook := interceptStatusCode(w)

	now := time.Now()
	m.next.ServeHTTP(hook, r.Clone(ctx))
	duration := time.Since(now)

	le.Log().Info("HTTP request",
		slog.Duration("duration", duration),
		slog.Any("http", datadogLogHttpRequest{
			Host:           r.Host,
			Proto:          r.Proto,
			Method:         r.Method,
			StatusCategory: statusCategoryFromCode(hook.StatusCode()),
			StatusCode:     hook.StatusCode(),
			URLDetails:     datadogLogUrlDetails{Path: r.URL.Path},
		}),
		slog.Any("network", map[string]datadogLogHttpClient{
			"client": clientInfoFromString(r.RemoteAddr),
		}),
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
	return hook.statusCode
}

type httpStatusHookHijackable struct {
	*httpStatusHook
}

func (hook *httpStatusHookHijackable) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return hook.ResponseWriter.(http.Hijacker).Hijack()
}

func clientInfoFromString(ipAndPort string) datadogLogHttpClient {
	ip, port, err := net.SplitHostPort(ipAndPort)
	if err != nil {
		ip = ipAndPort
		port = ""
	}
	return datadogLogHttpClient{
		IP:   ip,
		Port: port,
	}
}

func statusCategoryFromCode(code int) string {
	if code >= 200 && code < 400 {
		return "OK"
	} else if code >= 400 && code < 500 {
		return "warning"
	} else {
		return "error"
	}
}
