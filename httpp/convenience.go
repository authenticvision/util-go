package httpp

import (
	"encoding/json"
	"net/http"
	"strconv"
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
	hdr.Set("Content-Length", strconv.Itoa(len(buf)))

	_, err = w.Write(buf)
	if err != nil {
		// This usually fails with an I/O error when the client disconnects unexpectedly.
		return ServerError(err, "JSON write failure")
	}

	return nil
}

func BadRequest(err error, public ClientMessage) error {
	return Err(err, http.StatusBadRequest, public)
}

func Unauthorized(public ClientMessage) error {
	return Err(nil, http.StatusUnauthorized, public)
}

func Forbidden(public ClientMessage) error {
	return Err(nil, http.StatusForbidden, public)
}

func Unprocessable(err error, public ClientMessage) error {
	return Err(err, http.StatusUnprocessableEntity, public)
}

func ServerError(err error, public ClientMessage) error {
	return Err(err, http.StatusInternalServerError, public)
}
