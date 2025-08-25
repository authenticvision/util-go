package httpp

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/authenticvision/util-go/logutil"
)

func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func JSON(w http.ResponseWriter, resp any) error {
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

	_, err = w.Write(buf)
	if err != nil {
		// This usually fails with an I/O error when the client disconnects unexpectedly.
		return logutil.Severity(ServerError(err, "JSON write failure"), slog.LevelWarn)
	}

	return nil
}

func BadRequest(err error, public ClientMessage) error {
	return logutil.Severity(Err(err, http.StatusBadRequest, public), slog.LevelWarn)
}

func Unauthorized(public ClientMessage) error {
	return logutil.Severity(Err(nil, http.StatusUnauthorized, public), slog.LevelWarn)
}

func Forbidden(public ClientMessage) error {
	return logutil.Severity(Err(nil, http.StatusForbidden, public), slog.LevelWarn)
}

func Unprocessable(err error, public ClientMessage) error {
	return logutil.Severity(Err(err, http.StatusUnprocessableEntity, public), slog.LevelWarn)
}

func ServerError(err error, public ClientMessage) error {
	return Err(err, http.StatusInternalServerError, public)
}
