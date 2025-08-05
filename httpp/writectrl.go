package httpp

import "net/http"

// ResponseWriterUnwrapper models chained response writers for http.ResponseController
type ResponseWriterUnwrapper interface {
	Unwrap() http.ResponseWriter
}
