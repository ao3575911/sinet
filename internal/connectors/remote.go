package connectors

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

var denied = func() []netip.Prefix {
	raw := []string{"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24", "192.88.99.0/24", "192.168.0.0/16", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4", "::/96", "::ffff:0:0/96", "64:ff9b::/96", "64:ff9b:1::/48", "100::/64", "2001::/23", "2001:db8::/32", "2002::/16", "fc00::/7", "fe80::/10", "ff00::/8"}
	out := []netip.Prefix{}
	for _, v := range raw {
		out = append(out, netip.MustParsePrefix(v))
	}
	return out
}()

func PublicIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || ip.Zone() != "" || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, p := range denied {
		if p.Contains(ip) {
			return false
		}
	}
	return true
}
func Endpoint(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Port() != "" && u.Port() != "443") || len(u.Path) > 256 || strings.Contains(u.Path, "..") || strings.Contains(u.Hostname(), "%") {
		return nil, errors.New("invalid fixed HTTPS endpoint")
	}
	if addr, err := netip.ParseAddr(u.Hostname()); err == nil && !PublicIP(addr) {
		return nil, errors.New("private endpoint denied")
	}
	if !strings.Contains(u.Hostname(), ".") || strings.HasSuffix(strings.ToLower(u.Hostname()), ".localhost") {
		return nil, errors.New("non-public hostname denied")
	}
	return u, nil
}

type resolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

func safeDial(resolve resolver, dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil || port != "443" {
			return nil, errors.New("egress port denied")
		}
		ips, err := resolve.LookupNetIP(ctx, "ip", host)
		if err != nil || len(ips) == 0 {
			return nil, errors.New("DNS resolution failed")
		}
		for _, ip := range ips {
			if !PublicIP(ip) {
				return nil, errors.New("non-public DNS answer denied")
			}
		}
		chosen := ips[0].Unmap()
		conn, err := dial(ctx, "tcp", net.JoinHostPort(chosen.String(), "443"))
		if err != nil {
			return nil, errors.New("connector connection failed")
		}
		peer, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		actual, parseErr := netip.ParseAddr(peer)
		if err != nil || parseErr != nil || !PublicIP(actual) || actual.Unmap() != chosen {
			conn.Close()
			return nil, errors.New("connected IP mismatch")
		}
		return conn, nil
	}
}
func SafeClient() *http.Client {
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	transport := &http.Transport{Proxy: nil, DialContext: safeDial(net.DefaultResolver, dialer.DialContext), TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}, TLSHandshakeTimeout: 3 * time.Second, ResponseHeaderTimeout: 5 * time.Second, DisableKeepAlives: true, MaxResponseHeaderBytes: 16384}
	return &http.Client{Transport: transport, Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("connector redirects denied") }}
}
func Execute(ctx context.Context, m Manifest, op Operation, args json.RawMessage) (json.RawMessage, error) {
	if m.Runtime == "builtin" {
		if m.ID != "echo" || m.Version != "1.0.0" || op.Name != "repeat" {
			return nil, errors.New("builtin not implemented")
		}
		var input struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(args, &input); err != nil {
			return nil, err
		}
		output, _ := json.Marshal(map[string]string{"text": input.Text})
		return output, nil
	}
	u, err := Endpoint(m.EndpointURL)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/" + op.Name
	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(args))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := SafeClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("remote connector unavailable or destination denied")
	}
	defer resp.Body.Close()
	defer client.CloseIdleConnections()
	if resp.StatusCode != 200 {
		return nil, errors.New("remote connector returned failure")
	}
	out, err := io.ReadAll(io.LimitReader(resp.Body, 65537))
	if err != nil || len(out) > 65536 {
		return nil, errors.New("connector output limit")
	}
	if !json.Valid(out) {
		return nil, errors.New("connector response must be JSON")
	}
	return out, nil
}
