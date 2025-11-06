package httpp

import (
	"context"
	"fmt"
	"net/http"
)

// Handler mirrors http.Handler, but can additionally return an error.
// Errors are converted to HTTP status codes and can carry a public error message.
type Handler interface {
	ServeErrHTTP(http.ResponseWriter, *http.Request) error
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (f HandlerFunc) ServeErrHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// StdHandler universally serves as http.Handler and Handler.
type StdHandler interface {
	http.Handler
	Handler
}

// NeverErrors panics when a Handler returns a non-nil error.
func NeverErrors(handler Handler) StdHandler {
	return &noErrorHandler{next: handler}
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

// Adapt converts a stdlib http.Handler into a Handler. It will never return an error.
func Adapt(handler http.Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		handler.ServeHTTP(w, r)
		return nil
	})
}

// CollectErrors wraps a Handler in an http.Handler that stores errors
// returned from the wrapped handler in the Request context.
func CollectErrors(handler Handler) http.Handler {
	return &collectHandler{next: handler}
}

func WithErrorCollector(r *http.Request) (*http.Request, *error) {
	var err error
	ctx := context.WithValue(r.Context(), collectedErrorTag{}, &err)
	return r.WithContext(ctx), &err
}

var _ http.Handler = &collectHandler{}

type collectedErrorTag struct{}

type collectHandler struct {
	next Handler
}

func (h *collectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errPtr := r.Context().Value(collectedErrorTag{}).(*error)
	*errPtr = h.next.ServeErrHTTP(w, r)
}
