package httpp

import (
	"fmt"
	"net/http"
)

type Handler interface {
	ServeErrHTTP(http.ResponseWriter, *http.Request) error
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (f HandlerFunc) ServeErrHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

type StdHandler interface {
	http.Handler
	Handler
}

// EmitErrors adds a stdlib-compatible ServeHTTP method for use with http.Handler.
// Any errors are silently forwarded to the client and not processed further.
func EmitErrors(handler Handler) StdHandler {
	return &emitHandler{next: handler}
}

// EmitErrorsFunc wraps a HandlerFunc with a stdlib-compatible ServeHTTP method.
// Any errors are silently forwarded to the client and not processed further.
func EmitErrorsFunc(handlerFunc HandlerFunc) StdHandler {
	return &emitHandler{next: handlerFunc}
}

// NeverErrors panics when a Handler returns a non-nil error.
func NeverErrors(handler Handler) StdHandler {
	return &noErrorHandler{next: handler}
}

var _ StdHandler = &emitHandler{}

type emitHandler struct {
	next Handler
}

func (h *emitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.ServeErrHTTP(w, r); err != nil {
		WriteError(w, err)
	}
}

func (h *emitHandler) ServeErrHTTP(w http.ResponseWriter, r *http.Request) error {
	return h.next.ServeErrHTTP(w, r)
}

var _ StdHandler = &noErrorHandler{}

type noErrorHandler struct {
	next Handler
}

func (h *noErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = h.ServeErrHTTP(w, r)
}

func (h *noErrorHandler) ServeErrHTTP(w http.ResponseWriter, r *http.Request) error {
	err := h.next.ServeErrHTTP(w, r)
	if err != nil {
		panic(fmt.Errorf("unexpected error from handler: %w", err))
	}
	return nil
}
