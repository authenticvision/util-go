package httpp

import "net/http"

// ResponseWriterUnwrapper models chained response writers for http.ResponseController
type ResponseWriterUnwrapper interface {
	Unwrap() http.ResponseWriter
}

type CompressingWriter interface {
	DisableCompression()
	IsStreamingCompression() bool
}

// DisableCompression opts out of opportunistic output compression.
// Has no effect if no compression middleware is active or headers have already been sent.
func DisableCompression(w http.ResponseWriter) {
	for {
		switch t := w.(type) {
		case CompressingWriter:
			t.DisableCompression()
			return
		case ResponseWriterUnwrapper:
			w = t.Unwrap()
		default:
			return
		}
	}
}

// PrefersVariableContentLength checks whether a writer is more efficient when the Content-Length
// is omitted. This enables content to be freely rewritten, e.g. with a streaming compression
// algorithm that has unpredictable output length.
func PrefersVariableContentLength(w http.ResponseWriter) bool {
	for {
		if t, ok := w.(CompressingWriter); ok && t.IsStreamingCompression() {
			return true
			// otherwise try our luck with the next writer
		}
		if t, ok := w.(ResponseWriterUnwrapper); ok {
			w = t.Unwrap()
		} else {
			return false
		}
	}
}
