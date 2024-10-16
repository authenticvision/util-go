package httplog

// TODO: move to slog eventually

import (
	"github.com/authenticvision/httputil-go"
	"go.uber.org/zap"
	"net/http"
)

type Middleware struct {
	Logger *zap.Logger
}

func NewMiddleware(log *zap.Logger) *Middleware {
	return &Middleware{Logger: log}
}

func (s *Middleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := s.Logger.With(zap.String("url", r.URL.String()))
		handler.ServeHTTP(w, httputil.RequestWithValue(r, log))
	})
}

func FromRequest(r *http.Request) (logger *zap.Logger) {
	return r.Context().Value(logger).(*zap.Logger)
}
