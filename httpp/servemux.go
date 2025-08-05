package httpp

import (
	"net/http"
	urlpkg "net/url"
	"strings"
)

var _ Handler = &ServeMux{}

func NewServeMux() *ServeMux {
	return &ServeMux{next: http.NewServeMux()}
}

type ServeMux struct {
	next *http.ServeMux
}

func (mux *ServeMux) Handle(pattern string, handler Handler) {
	if h, ok := handler.(http.Handler); ok {
		mux.next.Handle(pattern, h)
	} else {
		mux.next.Handle(pattern, EmitErrors(handler))
	}
}

func (mux *ServeMux) HandleFunc(pattern string, handlerFunc HandlerFunc) {
	mux.Handle(pattern, handlerFunc)
}

func (mux *ServeMux) HandleStd(pattern string, handler http.Handler) {
	mux.next.Handle(pattern, handler)
}

func (mux *ServeMux) ServeErrHTTP(w http.ResponseWriter, r *http.Request) error {
	if r.RequestURI == "*" {
		// reject OPTIONS requests, copied from http.ServeMux.ServeHTTP
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	handlerStd, _ := mux.next.Handler(r)
	handler, ok := handlerStd.(Handler)
	if !ok {
		// for HandleStd handlers or generic handlers of http.ServeMux, e.g. NotFoundHandler
		handlerStd.ServeHTTP(w, r)
		return nil
	}
	return handler.ServeErrHTTP(w, r)
}

// StripPrefix is copied from Go 1.24.5's http.StripPrefix, with error forwarding added.
func StripPrefix(prefix string, h Handler) Handler {
	if prefix == "" {
		return h
	}
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		rp := strings.TrimPrefix(r.URL.RawPath, prefix)
		if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(urlpkg.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			r2.URL.RawPath = rp
			return h.ServeErrHTTP(w, r2)
		} else {
			http.NotFound(w, r)
			return nil
		}
	})
}
