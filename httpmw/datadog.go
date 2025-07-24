package httpmw

import (
	"net"
	"net/http"
)

type datadogLogHttpRequest struct {
	Host           string               `json:"host"`
	Proto          string               `json:"proto"`
	Method         string               `json:"method"`
	StatusCategory string               `json:"status_category"`
	StatusCode     int                  `json:"status_code"`
	URLDetails     datadogLogUrlDetails `json:"url_details"`
}

type datadogLogUrlDetails struct {
	Path string `json:"path"`
}

func makeDatadogHttpValue(r *http.Request, statusCode int) any {
	var category string
	if statusCode >= 200 && statusCode < 400 {
		category = "OK"
	} else if statusCode >= 400 && statusCode < 500 {
		category = "warning"
	} else {
		category = "error"
	}
	return datadogLogHttpRequest{
		Host:           r.Host,
		Proto:          r.Proto,
		Method:         r.Method,
		StatusCategory: category,
		StatusCode:     statusCode,
		URLDetails:     datadogLogUrlDetails{Path: r.URL.Path},
	}
}

type datadogLogPeer struct {
	IP   string `json:"ip"`
	Port string `json:"port,omitempty"`
}

func makeDatadogNetworkValue(r *http.Request) any {
	return map[string]datadogLogPeer{
		"client": datadogLogPeerFromAddress(r.RemoteAddr),
	}
}

func datadogLogPeerFromAddress(ipAndPort string) datadogLogPeer {
	ip, port, err := net.SplitHostPort(ipAndPort)
	if err != nil {
		ip = ipAndPort
		port = ""
	}
	return datadogLogPeer{
		IP:   ip,
		Port: port,
	}
}
