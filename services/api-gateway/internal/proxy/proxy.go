package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

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
	p := httputil.NewSingleHostReverseProxy(target)
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"service unavailable"}`))
	}
	return &ServiceProxy{target: target, proxy: p}, nil
}

// Forward is a Gin handler that proxies the request to the target service.
func (p *ServiceProxy) Forward() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Host = p.target.Host
		c.Request.URL.Host = p.target.Host
		c.Request.URL.Scheme = p.target.Scheme
		p.proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// ForwardWS routes /ws/notifications to notifProxy, everything else to p (chat).
func (p *ServiceProxy) ForwardWS(notifProxy *ServiceProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/ws/notifications") {
			c.Request.Host = notifProxy.target.Host
			c.Request.URL.Host = notifProxy.target.Host
			c.Request.URL.Scheme = notifProxy.target.Scheme
			notifProxy.proxy.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.Request.Host = p.target.Host
		c.Request.URL.Host = p.target.Host
		c.Request.URL.Scheme = p.target.Scheme
		p.proxy.ServeHTTP(c.Writer, c.Request)
	}
}
