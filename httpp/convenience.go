package httpp

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/authenticvision/util-go/logutil"
)

func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func JSONStatus(w http.ResponseWriter, resp any, statusCode int) error {
	buf, err := json.Marshal(resp)
	if err != nil {
		// A custom serializer likely returned an error.
		return ServerError(err, "encoding failure")
	}

	hdr := w.Header()
	hdr.Set("Content-Type", "application/json")
	if !PrefersVariableContentLength(w) {
		hdr.Set("Content-Length", strconv.Itoa(len(buf)))
	}

	w.WriteHeader(statusCode)

	_, err = w.Write(buf)
	if err != nil {
		// This usually fails with an I/O error when the client disconnects unexpectedly.
		return logutil.Severity(ServerError(err, "JSON write failure"), slog.LevelWarn)
	}

	return nil
}

func JSON(w http.ResponseWriter, resp any) error {
	return JSONStatus(w, resp, http.StatusOK)
}

func TextStatus(w http.ResponseWriter, resp string, statusCode int) error {
	if !strings.HasSuffix(resp, "\n") {
		resp += "\n"
	}

	hdr := w.Header()
	hdr.Set("Content-Type", "text/plain; charset=utf-8")
	if !PrefersVariableContentLength(w) {
		hdr.Set("Content-Length", strconv.Itoa(len(resp)))
	}

	w.WriteHeader(statusCode)

	_, err := w.Write([]byte(resp))
	if err != nil {
		// This usually fails with an I/O error when the client disconnects unexpectedly.
		return logutil.Severity(ServerError(err, "Text write failure"), slog.LevelWarn)
	}

	return nil
}

func Text(w http.ResponseWriter, resp string) error {
	return TextStatus(w, resp, http.StatusOK)
}

func BadRequest(err error, public PublicMessage) error {
	return logutil.Severity(Err(err, http.StatusBadRequest, public), slog.LevelWarn)
}

func Unauthorized(public PublicMessage) error {
	return logutil.Severity(Err(nil, http.StatusUnauthorized, public), slog.LevelWarn)
}

func Forbidden(public PublicMessage) error {
	return logutil.Severity(Err(nil, http.StatusForbidden, public), slog.LevelWarn)
}

func NotFound(public PublicMessage) error {
	return logutil.Severity(Err(nil, http.StatusNotFound, public), slog.LevelWarn)
}

func Unprocessable(err error, public PublicMessage) error {
	return logutil.Severity(Err(err, http.StatusUnprocessableEntity, public), slog.LevelWarn)
}

func ServerError(err error, public PublicMessage) error {
	return Err(err, http.StatusInternalServerError, public)
}
