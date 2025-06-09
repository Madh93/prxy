// Package prxy provides the core implementation of the reverse proxy server.
//
// It encapsulates the logic for creating an HTTP server that uses a
// reverse proxy to forward requests to a designated target URL. The key
// feature is its ability to route all outgoing traffic through a specified
// external HTTP proxy. The package handles the setup of the server, transport,
// and request rewriting, as well as managing the server's lifecycle.

package prxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/Madh93/prxy/internal/config"
	"github.com/Madh93/prxy/internal/logging"
)

// Prxy holds all the dependencies for the HTTP server.
type Prxy struct {
	logger *logging.Logger
	server *http.Server
}

// New creates and configures a new Prxy instance.
func New(cfg *config.Config, logger *logging.Logger) (*Prxy, error) {
	// 0. Ensure to parse URLs
	parsedTargetURL, err := url.Parse(cfg.Target)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL %q: %w", cfg.Target, err)
	}
	parsedProxyURL, err := url.Parse(cfg.Proxy)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL %q: %w", cfg.Proxy, err)
	}

	// 1. Creates Reverse Proxy Handler
	reverseProxyHandler := httputil.NewSingleHostReverseProxy(parsedTargetURL)

	// 1.1 Use the outbound HTTP Proxy for the transport
	transport := &http.Transport{
		Proxy: http.ProxyURL(parsedProxyURL),
	}
	reverseProxyHandler.Transport = transport

	// 1.2 Ensure the Host header is rewritten to the target's host.
	originalDirector := reverseProxyHandler.Director
	reverseProxyHandler.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = parsedTargetURL.Host
	}

	// 1.3 Custom error handler for better logging and response.
	reverseProxyHandler.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		logger.Error("Reverse proxy error", "url", req.URL.String(), "error", err)
		http.Error(rw, "Proxy Error: "+err.Error(), http.StatusBadGateway)
	}

	// 2. Creates HTTP httpServer
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler: reverseProxyHandler,
	}

	// Create main Prxy struct.
	prxy := &Prxy{
		logger: logger,
		server: httpServer,
	}

	return prxy, nil
}

// Run starts the HTTP server and blocks until it exits.
func (s Prxy) Run() error {
	// This method always returns a non-nil error. When Shutdown() is called,
	// it returns http.ErrServerClosed.
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s Prxy) Shutdown(ctx context.Context) error {
	s.logger.Debug("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}

// Addr returns the network address the server is listening on.
// Returns an empty string if the server is not running.
func (s Prxy) Addr() string {
	return s.server.Addr
}
