package httpp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// PublicMessage is a string sent to the client as-is. This prevents accidentally forwarding a
// non-constant string, e.g. an error message that contains confidential data, to clients.
type PublicMessage string

const DefaultMessage PublicMessage = ""

func Err(err error, statusCode int, msg PublicMessage) error {
	return Error{
		err:           err,
		statusCode:    statusCode,
		msg:        msg,
	}
}

type Error struct {
	err           error
	statusCode    int
	msg        PublicMessage
}

func (e Error) Error() string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "http status %d", e.statusCode)
	if e.msg != "" {
		_, _ = fmt.Fprintf(&sb, ", %s", e.msg)
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
	if e.msg != "" {
		return string(e.msg)
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
