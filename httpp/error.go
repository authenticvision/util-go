package httpp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ClientMessage is a string sent to the client as-is. This prevents accidentally forwarding a
// non-constant string, e.g. an error message that contains confidential data, to clients.
type ClientMessage string

const DefaultMessage ClientMessage = ""

func Err(err error, statusCode int, clientMessage ClientMessage) error {
	return Error{
		err:           err,
		statusCode:    statusCode,
		clientMessage: clientMessage,
	}
}

type Error struct {
	err           error
	statusCode    int
	clientMessage ClientMessage
}

func (e Error) Error() string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "http status %d", e.statusCode)
	if e.clientMessage != "" {
		_, _ = fmt.Fprintf(&sb, ", %s", e.clientMessage)
	}
	if e.err != nil {
		_, _ = fmt.Fprintf(&sb, ": %v", e.err)
	}
	return sb.String()
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) StatusCode() int {
	return e.statusCode
}

func (e Error) StatusText() string {
	if e.clientMessage != "" {
		return string(e.clientMessage)
	}
	return http.StatusText(e.statusCode)
}

// WriteError writes err as HTTP response. Middlewares that act on errors use WriteError to forward
// errors to clients. It is invalid to write an error when an HTTP response was previously started.
func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		panic("httpp.WriteError called with nil error")
	}
	var httpErr Error
	if errors.As(err, &httpErr) {
		http.Error(w, httpErr.StatusText(), httpErr.StatusCode())
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
