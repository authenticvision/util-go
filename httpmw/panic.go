package httpmw

import (
	"fmt"
	"github.com/authenticvision/util-go/httpp"
	"github.com/authenticvision/util-go/logutil"
	"net/http"
)

// PanicError wraps a panic value and is returned when a handler panics.
type PanicError struct {
	Value any
}

func (e PanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.Value)
}

func (e PanicError) Unwrap() error {
	if err, ok := e.Value.(error); ok {
		return err
	}
	return nil
}

// NewPanicMiddleware logs handler panics and returns them as error via PanicError.
func NewPanicMiddleware() Middleware {
	return &panicMiddleware{}
}

type panicMiddleware struct {
	next httpp.Handler
}

func (m *panicMiddleware) Middleware(next httpp.Handler) httpp.Handler {
	return &panicHandler{next: next}
}

type panicHandler struct {
	next httpp.Handler
}

func (h *panicHandler) ServeErrHTTP(w http.ResponseWriter, r *http.Request) (result error) {
	defer func() {
		if rec := recover(); rec != nil {
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			err := PanicError{Value: rec}
			log := logutil.FromContext(r.Context())
			log.Error("http handler panic", logutil.Err(err), logutil.Stack(3))
			result = httpp.ServerError(err, httpp.DefaultMessage)
		}
	}()
	return h.next.ServeErrHTTP(w, r)
}
