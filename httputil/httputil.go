package httputil

import (
	"context"
	"net/http"
)

type Middleware interface {
	Middleware(handler http.Handler) http.Handler
}

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware.Middleware(handler)
	}
	return handler
}

func RequestWithValue[T any](r *http.Request, value *T) *http.Request {
	var tag *T
	return r.WithContext(context.WithValue(r.Context(), tag, value))
}
