package httpmw

import (
	"github.com/authenticvision/util-go/logutil"
	"log/slog"
	"net/http"
	"runtime/debug"
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
			log := logutil.FromContext(r.Context())
			log.Error("http handler panic",
				slog.Any("error", err),
				slog.String("stack", string(debug.Stack())))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}()

	h.next.ServeHTTP(w, r)
}
