// Package ddlog generates log attributes for Datadog's standard attributes:
// https://docs.datadoghq.com/standard-attributes/?product=log
package ddlog

import (
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"net"
	"net/http"
	"strconv"
)

type HttpStatusRecorder interface {
	StatusCode() int
	BytesWritten() uint64
}

// WithRequest attaches fields that are known at request time, before the request is answered.
func WithRequest(log *slog.Logger, r *http.Request, id uuid.UUID) *slog.Logger {
	return log.With(
		slog.Any("http", HTTPRequest{
			ID:         id.String(),
			Version:    versionFromRequest(r),
			URLDetails: urlDetailsFromRequest(r),
		}),
		slog.Any("network", Network{
			Client: peerFromAddress(r.RemoteAddr),
		}),
	)
}

// WithResponse attaches request and additionally response information to a logger.
// It must not be used on a logger created through WithRequest.
func WithResponse(log *slog.Logger, r *http.Request, id uuid.UUID, recorder HttpStatusRecorder) *slog.Logger {
	statusCode := recorder.StatusCode()
	statusCategory := "error"
	switch {
	case statusCode >= 200 && statusCode < 400:
		statusCategory = "OK"
	case statusCode >= 400 && statusCode < 500:
		statusCategory = "warning"
	}

	return log.With(
		slog.Any("http", HTTPRequest{
			ID:             id.String(),
			Version:        versionFromRequest(r),
			StatusCode:     statusCode,
			StatusCategory: statusCategory,
			URLDetails:     urlDetailsFromRequest(r),
		}),
		slog.Any("network", Network{
			// LATER: BytesRead, eventually, though that requires intercepting the request's body
			BytesWritten: recorder.BytesWritten(),
			Client:       peerFromAddress(r.RemoteAddr),
		}),
	)
}

type HTTPRequest struct {
	ID             string     `json:"request_id,omitempty"`
	Version        string     `json:"version,omitempty"` // e.g., "1.1", "2"
	StatusCategory string     `json:"status_category,omitempty"`
	StatusCode     int        `json:"status_code,omitempty"`
	URLDetails     URLDetails `json:"url_details"`
}

type URLDetails struct {
	Host        string            `json:"host,omitempty"`
	Port        int               `json:"port,omitempty"`
	Path        string            `json:"path,omitempty"`
	QueryString map[string]string `json:"queryString,omitempty"` // sic, this really wants camelCase
}

func versionFromRequest(r *http.Request) string {
	if r.ProtoMajor == 1 || r.ProtoMinor != 0 {
		return fmt.Sprintf("%d.%d", r.ProtoMajor, r.ProtoMinor)
	} else {
		return strconv.Itoa(r.ProtoMajor)
	}
}

func urlDetailsFromRequest(r *http.Request) (result URLDetails) {
	host, port, err := net.SplitHostPort(r.Host)
	if err == nil {
		result.Host = host
		if portNum, err := strconv.Atoi(port); err == nil && portNum != 0 {
			result.Port = portNum
		}
	}

	result.Path = r.URL.Path
	query := r.URL.Query()
	if len(query) != 0 {
		result.QueryString = make(map[string]string, len(query))
		for k, v := range query {
			if len(v) != 0 {
				// In case of duplicates, only the first value is logged. This is consistent with
				// the implementation of http.Value.Get, which the application is likely to use.
				result.QueryString[k] = v[0]
			}
		}
	}

	return
}

type Network struct {
	// LATER: see above: BytesRead uint64 `json:"bytes_read,omitempty"`
	BytesWritten uint64      `json:"bytes_written,omitempty"`
	Client       NetworkPeer `json:"client"`
}

type NetworkPeer struct {
	IP   string `json:"ip"`
	Port string `json:"port,omitempty"`
}

func peerFromAddress(ipAndPort string) NetworkPeer {
	ip, port, err := net.SplitHostPort(ipAndPort)
	if err != nil {
		ip = ipAndPort
		port = ""
	}
	return NetworkPeer{
		IP:   ip,
		Port: port,
	}
}
