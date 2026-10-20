package client

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Session struct {
	API   string `json:"api"`
	Token string `json:"token"`
}
type Client struct {
	API, Config string
	session     Session
	HTTP        *http.Client
}

func New() (*Client, error) {
	path := os.Getenv("SINET_CONFIG")
	if path == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(base, "SINet", "session.json")
	}
	c := &Client{API: os.Getenv("SINET_API"), Config: path, HTTP: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("API redirect denied") }}}
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("session file must not be a symlink")
		}
		if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
			return nil, errors.New("session file permissions must be 0600")
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if json.Unmarshal(raw, &c.session) != nil {
			return nil, errors.New("invalid session file")
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if c.API == "" {
		c.API = c.session.API
	}
	if c.API == "" {
		c.API = "http://127.0.0.1:8080"
	}
	c.API = strings.TrimRight(c.API, "/")
	u, err := url.Parse(c.API)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"))) {
		return nil, errors.New("API must be HTTPS, or loopback HTTP")
	}
	return c, nil
}
func (c *Client) Authenticated() bool { return c.session.Token != "" && c.session.API == c.API }
func (c *Client) Save(token string) error {
	if len(token) != 64 {
		return errors.New("invalid session token")
	}
	dir := filepath.Dir(c.Config)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if info, err := os.Lstat(c.Config); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("session file must not be a symlink")
	}
	f, err := os.CreateTemp(dir, ".session-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	c.session = Session{c.API, token}
	if err = json.NewEncoder(f).Encode(c.session); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), c.Config)
}
func (c *Client) Forget() error {
	c.session = Session{}
	err := os.Remove(c.Config)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
func (c *Client) Do(method, path string, body any) (map[string]any, error) {
	var raw []byte
	var err error
	if body != nil {
		raw, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, c.API+path, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Authenticated() {
		req.Header.Set("Authorization", "Bearer "+c.session.Token)
	}
	if method != "GET" {
		key := make([]byte, 16)
		if _, err = rand.Read(key); err != nil {
			return nil, err
		}
		req.Header.Set("Idempotency-Key", hex.EncodeToString(key))
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, errors.New("API unavailable; check SINET_API and server health")
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024*1024+1))
	if err != nil || len(data) > 32*1024*1024 {
		return nil, errors.New("API response exceeds client budget")
	}
	var out map[string]any
	if json.Unmarshal(data, &out) != nil {
		return nil, errors.New("API returned invalid JSON")
	}
	if resp.StatusCode >= 400 {
		code, _ := out["error"].(string)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, code)
	}
	return out, nil
}
