// Package githubconnector serves a fixed, read-only public GitHub operation.
// Deployment does not imply approval in the SINet connector registry.
package githubconnector

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"sync"
	"time"

	"sinet/internal/connectors"
	"sinet/internal/githubpublic"
)

const upstream = githubpublic.Upstream
const route = "/repository_summary"
const maxResponse = 65536
const requestsPerMinute = 60

type gateway struct {
	client   *http.Client
	slots    chan struct{}
	mu       sync.Mutex
	window   time.Time
	requests int
}

// NewHandler uses the existing DNS-pinned, public-IP-only HTTPS transport.
// Proxy environment variables, redirects and caller headers are never forwarded.
func NewHandler() http.Handler {
	return NewHandlerWithTransport(nil)
}

// NewHandlerWithTransport supplies an isolated test upstream; nil retains safe HTTPS.
// The standalone executable always calls NewHandler.
func NewHandlerWithTransport(transport http.RoundTripper) http.Handler {
	client := connectors.SafeClient()
	if transport != nil {
		client.Transport = transport
	}
	return newHandler(client)
}

func newHandler(client *http.Client) *gateway {
	return &gateway{client: client, slots: make(chan struct{}, 4)}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func fail(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}

func (g *gateway) budget() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now()
	if g.window.IsZero() || now.Sub(g.window) >= time.Minute {
		g.window, g.requests = now, 0
	}
	if g.requests >= requestsPerMinute {
		return false
	}
	g.requests++
	return true
}

func (g *gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" && r.URL.RawQuery == "" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	if r.URL.EscapedPath() != route || r.URL.RawQuery != "" {
		fail(w, http.StatusNotFound, "unknown_operation")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		fail(w, http.StatusMethodNotAllowed, "post_required")
		return
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		fail(w, http.StatusUnsupportedMediaType, "json_required")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 512))
	if err != nil {
		fail(w, http.StatusRequestEntityTooLarge, "input_limit")
		return
	}
	if !githubpublic.ValidInput(body) {
		fail(w, http.StatusBadRequest, "invalid_repository_arguments")
		return
	}
	select {
	case g.slots <- struct{}{}:
		defer func() { <-g.slots }()
	default:
		fail(w, http.StatusServiceUnavailable, "gateway_busy")
		return
	}
	if !g.budget() {
		w.Header().Set("Retry-After", "60")
		fail(w, http.StatusTooManyRequests, "gateway_rate_limited")
		return
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, upstream, nil)
	if err != nil {
		fail(w, http.StatusInternalServerError, "request_failed")
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "SINet-github-public/1.0.0")
	response, err := g.client.Do(req)
	if err != nil {
		var networkError net.Error
		if errors.As(err, &networkError) && networkError.Timeout() {
			fail(w, http.StatusGatewayTimeout, "upstream_timeout")
		} else {
			fail(w, http.StatusBadGateway, "upstream_unavailable")
		}
		return
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		fail(w, http.StatusNotFound, "repository_not_found")
		return
	case http.StatusTooManyRequests:
		fail(w, http.StatusServiceUnavailable, "upstream_rate_limited")
		return
	case http.StatusForbidden:
		if response.Header.Get("X-RateLimit-Remaining") == "0" {
			fail(w, http.StatusServiceUnavailable, "upstream_rate_limited")
		} else {
			fail(w, http.StatusServiceUnavailable, "upstream_forbidden")
		}
		return
	default:
		fail(w, http.StatusBadGateway, "upstream_failure")
		return
	}
	media, _, err = mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		fail(w, http.StatusBadGateway, "invalid_upstream_response")
		return
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponse+1))
	if err != nil {
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			fail(w, http.StatusGatewayTimeout, "upstream_timeout")
		} else {
			fail(w, http.StatusBadGateway, "upstream_failure")
		}
		return
	}
	if len(data) > maxResponse {
		fail(w, http.StatusBadGateway, "upstream_response_limit")
		return
	}
	summary, err := githubpublic.ParseUpstream(data)
	if err != nil {
		fail(w, http.StatusBadGateway, "invalid_upstream_response")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
