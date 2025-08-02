package httpmw

import (
	"fmt"
	"github.com/authenticvision/util-go/logutil"
	"log/slog"
	"net/http"
)

func NewPanicMiddleware() *PanicMiddleware {
	return &PanicMiddleware{}
}

type PanicMiddleware struct {
	next http.Handler
}

func (m *PanicMiddleware) Middleware(next http.Handler) http.Handler {
	return &panicHandler{next: next}
}

type panicHandler struct {
	next http.Handler
}

func (h *panicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			if err == http.ErrAbortHandler {
				panic(err)
			}

			var errAttr slog.Attr
			if e, ok := err.(error); ok {
				errAttr = logutil.Err(e)
			} else {
				errAttr = logutil.ErrColor(slog.String(logutil.KeyErr, fmt.Sprintf("panic: %v", err)))
			}

			log := logutil.FromContext(r.Context())
			log.Error("http handler panic", errAttr, logutil.Stack(3))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}()

	h.next.ServeHTTP(w, r)
}
