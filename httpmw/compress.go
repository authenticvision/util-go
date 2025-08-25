package httpmw

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/authenticvision/util-go/httpp"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

// NewCompressionMiddleware opportunistically compresses the response with zstd or gzip.
// The fastest possible compression mode is used.
func NewCompressionMiddleware() Middleware {
	return &compressMiddleware{
		// earlier codecs are preferred
		codecs: []compressCodec{
			newCompressCodec("zstd", func() (*zstd.Encoder, error) {
				return zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
			}),
			newCompressCodec("gzip", func() (*gzip.Writer, error) {
				return gzip.NewWriterLevel(nil, gzip.BestSpeed)
			}),
		},
	}
}

type resettableWriter interface {
	Write(p []byte) (int, error)
	Close() error
	Flush() error
	Reset(w io.Writer)
}

func newCompressCodec[T resettableWriter](name string, create func() (T, error)) compressCodec {
	pool := &sync.Pool{}
	pool.New = func() any {
		encoder, err := create()
		if err != nil {
			panic(fmt.Errorf("httpmw: failed to create %s encoder: %w", name, err))
		}
		return encoder
	}
	return compressCodec{name: name, pool: pool}
}

type compressCodec struct {
	name string
	pool *sync.Pool
}

type compressMiddleware struct {
	codecs []compressCodec
}

func (m *compressMiddleware) Middleware(next httpp.Handler) httpp.Handler {
	return &compressHandler{compressMiddleware: m, next: next}
}

type compressHandler struct {
	*compressMiddleware
	next httpp.Handler
}

func (h *compressHandler) ServeErrHTTP(w http.ResponseWriter, r *http.Request) error {
	accepted := parseAcceptedEncodings(r)
	for _, codec := range h.codecs {
		if accepted.Contains(codec.name) {
			return h.serveEncoded(w, r, codec)
		}
	}
	return h.next.ServeErrHTTP(w, r)
}

func (h *compressHandler) serveEncoded(w http.ResponseWriter, r *http.Request, codec compressCodec) (result error) {
	cw := &compressWriter{
		ResponseWriter: w,
		compressCodec:  codec,
	}
	defer func() {
		err := cw.Close()
		result = errors.Join(result, err)
	}()
	return h.next.ServeErrHTTP(cw, r)
}

var _ interface {
	http.ResponseWriter
	httpp.ResponseWriterUnwrapper
	httpp.CompressingWriter
} = &compressWriter{}

type compressWriter struct {
	http.ResponseWriter
	compressCodec
	encoder     resettableWriter
	optOut      bool
	wroteHeader bool
}

func (w *compressWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *compressWriter) IsStreamingCompression() bool {
	return true
}

func (w *compressWriter) DisableCompression() {
	w.optOut = true
}

func (w *compressWriter) WriteHeader(statusCode int) {
	hdr := w.Header()
	switch {
	case w.wroteHeader:
		// programmer error, fall through to next WriteHeader to let runtime report it

	case w.optOut:
		// DisableCompression was called

	case hdr.Get("Content-Encoding") != "":
		// already compressed

	case hdr.Get("Content-Length") != "":
		// respect the handler's wish to send a fixed-length response as-is

	default:
		hdr.Set("Content-Encoding", w.name)
		addVary(hdr, "Accept-Encoding")
		w.encoder = w.pool.Get().(resettableWriter)
		w.encoder.Reset(w.ResponseWriter)
	}

	w.ResponseWriter.WriteHeader(statusCode)
	w.wroteHeader = true
}

func (w *compressWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.encoder != nil {
		n, err := w.encoder.Write(b)
		if err != nil {
			return 0, fmt.Errorf("%s encode: %w", w.name, err)
		}
		return n, nil
	} else {
		return w.ResponseWriter.Write(b)
	}
}

func (w *compressWriter) FlushError() error {
	if w.encoder != nil {
		if err := w.encoder.Flush(); err != nil {
			return fmt.Errorf("%s flush: %w", w.name, err)
		}
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *compressWriter) Close() error {
	if w.encoder == nil {
		return nil
	}

	err := w.encoder.Close()
	if err != nil {
		err = fmt.Errorf("%s encoder close: %w", w.name, err)
	}

	w.encoder.Reset(io.Discard) // to allow GC of w.ResponseWriter
	w.pool.Put(w.encoder)
	w.encoder = nil

	return err
}

type acceptedEncodings map[string]struct{}

func (m acceptedEncodings) Contains(name string) bool {
	_, ok := m[name]
	return ok
}

func parseAcceptedEncodings(r *http.Request) acceptedEncodings {
	accepted := r.Header.Get("Accept-Encoding")
	if accepted == "" {
		return nil
	}
	codecs := make(map[string]struct{}, 4)
	for _, name := range strings.Split(accepted, ",") {
		name = strings.TrimSpace(name)
		name, _, _ = strings.Cut(name, ";") // remove parameters
		if name != "" {
			codecs[name] = struct{}{}
		}
	}
	return codecs
}

func addVary(hdr http.Header, value string) {
	if vary := hdr.Get("Vary"); vary != "" {
		hdr.Set("Vary", vary+", "+value)
	} else {
		hdr.Set("Vary", value)
	}
}
