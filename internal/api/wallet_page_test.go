package api

import (
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestWalletSigningPagePolicy(t *testing.T) {
	server := New(nil, nil, nil, "https://sinet.example")
	nonces := map[string]bool{}
	for range 2 {
		w := httptest.NewRecorder()
		server.ServeHTTP(w, httptest.NewRequest("GET", "/wallet/link", nil))
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
			t.Fatal("wallet helper failed", w.Code)
		}
		csp := w.Header().Get("Content-Security-Policy")
		match := regexp.MustCompile(`script-src 'nonce-([0-9a-f]{32})'`).FindStringSubmatch(csp)
		if len(match) != 2 || nonces[match[1]] || !strings.Contains(w.Body.String(), `nonce="`+match[1]+`"`) || !strings.Contains(csp, "connect-src 'none'") || !strings.Contains(csp, "frame-ancestors 'none'") {
			t.Fatal("wallet page policy missing or nonce reused")
		}
		nonces[match[1]] = true
		if w.Header().Get("Referrer-Policy") != "no-referrer" || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("wallet link privacy headers missing")
		}
		if !strings.Contains(w.Body.String(), `personal_sign`) || strings.Contains(w.Body.String(), `eth_sendTransaction`) {
			t.Fatal("wallet authority expanded")
		}
	}
	w := httptest.NewRecorder()
	server.ServeHTTP(w, httptest.NewRequest("GET", "/wallet/link?message=private", nil))
	if w.Code != 400 {
		t.Fatal("challenge accepted in URL query")
	}
}
