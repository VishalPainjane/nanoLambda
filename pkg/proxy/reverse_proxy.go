package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewReverseProxy creates a proxy that forwards requests to the container
func NewReverseProxy(targetAddr string) *httputil.ReverseProxy {
	targetURL := &url.URL{
		Scheme: "http",
		Host:   targetAddr,
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Update the Director to rewrite the path
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Rewrite path to /invoke
		req.URL.Path = "/invoke"
		// Ensure Host header is set correctly (some servers require it)
		req.Host = targetAddr
	}
	
	// Error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Printf("Proxy error: %v\n", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	return proxy
}
