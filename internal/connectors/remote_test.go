package connectors

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"testing"
	"time"
)

type fakeResolver []netip.Addr

func (r fakeResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return r, nil
}

type fakeConn struct {
	net.Conn
	peer string
}

func (c fakeConn) RemoteAddr() net.Addr { return &net.TCPAddr{IP: net.ParseIP(c.peer), Port: 443} }
func (c fakeConn) Close() error         { return nil }
func TestPublicDestinations(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.1.2.3", "169.254.169.254", "100.64.0.1", "192.168.1.1", "172.16.0.1", "0.0.0.0", "224.0.0.1", "198.18.0.1", "192.0.2.1", "::1", "fc00::1", "fe80::1", "::ffff:127.0.0.1", "64:ff9b::7f00:1", "2002:7f00:1::", "2001:db8::1"} {
		if PublicIP(netip.MustParseAddr(raw)) {
			t.Errorf("unsafe IP accepted: %s", raw)
		}
	}
	for _, raw := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"} {
		if !PublicIP(netip.MustParseAddr(raw)) {
			t.Errorf("public IP rejected: %s", raw)
		}
	}
	for _, raw := range []string{"http://example.com", "https://127.0.0.1", "https://10.0.0.1", "https://localhost", "https://x.localhost", "https://user:pass@example.com", "https://example.com:8443", "https://example.com/?target=x", "https://example.com/#x"} {
		if _, err := Endpoint(raw); err == nil {
			t.Errorf("unsafe endpoint accepted: %s", raw)
		}
	}
}
func TestDNSAndConnectedIPChecks(t *testing.T) {
	for _, ips := range []fakeResolver{{netip.MustParseAddr("127.0.0.1")}, {netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("10.0.0.1")}} {
		called := false
		dial := safeDial(ips, func(context.Context, string, string) (net.Conn, error) {
			called = true
			return nil, errors.New("unexpected dial")
		})
		if _, err := dial(context.Background(), "tcp", "example.com:443"); err == nil || called {
			t.Fatal("private or mixed DNS answer reached dial")
		}
	}
	ips := fakeResolver{netip.MustParseAddr("8.8.8.8")}
	dial := safeDial(ips, func(_ context.Context, _ string, address string) (net.Conn, error) {
		if address != "8.8.8.8:443" {
			t.Fatal("DNS not pinned")
		}
		return fakeConn{peer: "127.0.0.1"}, nil
	})
	if _, err := dial(context.Background(), "tcp", "example.com:443"); err == nil {
		t.Fatal("peer mismatch accepted")
	}
	dial = safeDial(ips, func(context.Context, string, string) (net.Conn, error) { return fakeConn{peer: "8.8.8.8"}, nil })
	if _, err := dial(context.Background(), "tcp", "example.com:443"); err != nil {
		t.Fatal(err)
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestRedirectsAndBudgets(t *testing.T) {
	client := SafeClient()
	transport := client.Transport.(*http.Transport)
	if transport.Proxy != nil || client.Timeout > 10*time.Second || transport.ResponseHeaderTimeout == 0 || transport.MaxResponseHeaderBytes == 0 {
		t.Fatal("egress budgets or proxy policy missing")
	}
	calls := 0
	client.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"https://127.0.0.1/secret"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	})
	if _, err := client.Get("https://example.com"); err == nil || calls != 1 {
		t.Fatal("redirect followed")
	}
}
