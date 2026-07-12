package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"proxy-gateway/logger"
)

func NewMakeProxy(target *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			HandleDirector(req, target)
		},
		ModifyResponse: HandleModifyResponse,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Logger.Error("HTTP proxy error", "method", r.Method, "url", r.URL.String(), "error", err)
			w.WriteHeader(http.StatusBadGateway)
			io.WriteString(w, fmt.Sprintf("Proxy error: %v", err))
		},
	}
}
