package httputil

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

type panicHandler struct {
	log  *slog.Logger
	next http.Handler
}

func PanicHandler(log *slog.Logger, next http.Handler) http.Handler {
	return &panicHandler{
		log:  log,
		next: next,
	}
}

func (h *panicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			if err == http.ErrAbortHandler {
				panic(err)
			}
			h.log.Error("http handler panic",
				slog.Any("error", err),
				slog.String("stack", string(debug.Stack())))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}()

	h.next.ServeHTTP(w, r)
}
