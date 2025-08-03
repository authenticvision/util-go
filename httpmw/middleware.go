package httpmw

import "github.com/authenticvision/util-go/httpp"

type Middleware interface {
	Middleware(handler httpp.Handler) httpp.Handler
}

func Chain(handler httpp.Handler, middlewares ...Middleware) httpp.Handler {
	for _, middleware := range middlewares {
		handler = middleware.Middleware(handler)
	}
	return handler
}
