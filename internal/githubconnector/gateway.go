// Package githubconnector serves a fixed, read-only public GitHub operation.
// Deployment does not imply approval in the SINet connector registry.
package githubconnector

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"sinet/internal/connectors"
)

const upstream = "https://api.github.com/repos/ao3575911/sinet"
const publicURL = "https://github.com/ao3575911/sinet"
const route = "/repository_summary"
const maxResponse = 65536
const requestsPerMinute = 60

// Summary contains only explicitly approved public fields, never raw GitHub JSON.
type Summary struct {
	FullName    string  `json:"full_name"`
	HTMLURL     string  `json:"html_url"`
	Description *string `json:"description"`
	Language    *string `json:"language"`
	Stars       int64   `json:"stargazers_count"`
	Forks       int64   `json:"forks_count"`
	OpenIssues  int64   `json:"open_issues_count"`
	Archived    bool    `json:"archived"`
	UpdatedAt   string  `json:"updated_at"`
	Source      string  `json:"source"`
}

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
	return newHandler(connectors.SafeClient())
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

// object rejects duplicate keys and trailing values as well as malformed JSON.
// Raw fields allow upstream additions without ever forwarding them.
func object(data []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(data) {
		return nil, errors.New("invalid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errors.New("object required")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, errors.New("invalid key")
		}
		if _, exists := fields[key]; exists {
			return nil, errors.New("duplicate key")
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return nil, err
		}
		fields[key] = raw
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, errors.New("trailing data")
	}
	return fields, nil
}

func validInput(data []byte) bool {
	fields, err := object(data)
	if err != nil || len(fields) != 2 {
		return false
	}
	var owner, repository string
	return json.Unmarshal(fields["owner"], &owner) == nil &&
		json.Unmarshal(fields["repository"], &repository) == nil &&
		owner == "ao3575911" && repository == "sinet"
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
	if !validInput(body) {
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
	if err != nil || len(data) > maxResponse {
		fail(w, http.StatusBadGateway, "upstream_response_limit")
		return
	}
	summary, err := parseSummary(data)
	if err != nil {
		fail(w, http.StatusBadGateway, "invalid_upstream_response")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func safeText(text string, limit int) bool {
	if len(text) > limit {
		return false
	}
	for _, r := range text {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == unicode.ReplacementChar {
			return false
		}
	}
	return true
}

func parseSummary(data []byte) (Summary, error) {
	var summary Summary
	fields, err := object(data)
	if err != nil {
		return summary, err
	}
	var private bool
	values := map[string]any{
		"full_name": &summary.FullName, "html_url": &summary.HTMLURL,
		"description": &summary.Description, "language": &summary.Language,
		"stargazers_count": &summary.Stars, "forks_count": &summary.Forks,
		"open_issues_count": &summary.OpenIssues, "archived": &summary.Archived,
		"updated_at": &summary.UpdatedAt, "private": &private,
	}
	for name, destination := range values {
		raw, exists := fields[name]
		if !exists || (name != "description" && name != "language" && strings.TrimSpace(string(raw)) == "null") {
			return summary, errors.New("missing upstream field")
		}
		if err := json.Unmarshal(raw, destination); err != nil {
			return summary, err
		}
	}
	if private || summary.FullName != "ao3575911/sinet" || summary.HTMLURL != publicURL {
		return summary, errors.New("unexpected repository")
	}
	for _, count := range []int64{summary.Stars, summary.Forks, summary.OpenIssues} {
		if count < 0 || count > 9007199254740991 {
			return summary, errors.New("invalid count")
		}
	}
	if summary.Description != nil && !safeText(*summary.Description, 1024) {
		return summary, errors.New("invalid description")
	}
	if summary.Language != nil && !safeText(*summary.Language, 128) {
		return summary, errors.New("invalid language")
	}
	if len(summary.UpdatedAt) > 40 {
		return summary, errors.New("invalid timestamp")
	}
	if _, err := time.Parse(time.RFC3339, summary.UpdatedAt); err != nil {
		return summary, err
	}
	summary.Source = upstream
	return summary, nil
}
