package httpmw

import (
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
