package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecurityErrorsAreBoundedCodes(t *testing.T) {
	for _, code := range []string{"upstream_failed", "line\nforged success", "\x1b[2Jsecret", strings.Repeat("a", 129)} {
		t.Run(code[:min(len(code), 20)], func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(502)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
			}))
			defer server.Close()
			c := &Client{API: server.URL, HTTP: server.Client()}
			_, err := c.Do("GET", "/test", nil)
			want := "HTTP 502: request_failed"
			if code == "upstream_failed" {
				want = "HTTP 502: upstream_failed"
			}
			if err == nil || err.Error() != want {
				t.Fatalf("unsafe error: %v", err)
			}
		})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		_, _ = w.Write([]byte(strings.Repeat("x", 4097)))
	}))
	defer server.Close()
	c := &Client{API: server.URL, HTTP: server.Client()}
	_, err := c.Do("GET", "/test", nil)
	if err == nil || err.Error() != "API response exceeds client budget" {
		t.Fatal("unbounded error body", err)
	}
}
func TestSecurityInvalidSessionFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	t.Setenv("SINET_CONFIG", path)
	t.Setenv("SINET_API", "http://127.0.0.1:8080")
	for _, raw := range []string{strings.Repeat(" ", 4097), `{"api":"http://127.0.0.1:8080","token":"short"}`, `{"api":"http://127.0.0.1:8080","token":"` + strings.Repeat("z", 64) + `"}`} {
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := New(); err == nil {
			t.Fatal("invalid session accepted")
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := New(); err == nil {
		t.Fatal("nonregular session accepted")
	}
	c := &Client{API: "http://127.0.0.1:8080", Config: path}
	if err := c.Save(strings.Repeat("a", 64)); err == nil {
		t.Fatal("overwrote directory")
	}
	if c.Authenticated() {
		t.Fatal("failed session save changed authentication state")
	}
}

func TestSecurityResponsePreservesExactNumbers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"output":{"value":9007199254740993}}`))
	}))
	defer server.Close()
	c := &Client{API: server.URL, HTTP: server.Client()}
	out, err := c.Do("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out["output"].(map[string]any)["value"] != json.Number("9007199254740993") {
		t.Fatal("API result number rounded")
	}
}
