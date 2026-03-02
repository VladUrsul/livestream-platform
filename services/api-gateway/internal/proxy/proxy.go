package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// ServiceProxy forwards requests to a downstream microservice.
type ServiceProxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

// NewServiceProxy creates a reverse proxy pointing at targetURL.
func NewServiceProxy(targetURL string) (*ServiceProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL %q: %w", targetURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Custom error handler so proxy failures return clean JSON
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"service unavailable"}`))
	}

	return &ServiceProxy{target: target, proxy: proxy}, nil
}

// Forward is a Gin handler that proxies the request to the target service.
func (p *ServiceProxy) Forward() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Rewrite the host header to the target service
		c.Request.Host = p.target.Host

		// Remove the double-slash that can appear when stripping prefixes
		c.Request.URL.Host = p.target.Host
		c.Request.URL.Scheme = p.target.Scheme

		p.proxy.ServeHTTP(c.Writer, c.Request)
	}
}
