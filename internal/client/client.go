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
	"regexp"
	"runtime"
	"sinet/internal/strictjson"
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
		if !info.Mode().IsRegular() {
			return nil, errors.New("session file must be a regular file, not a symlink or device")
		}
		if info.Size() > 4096 {
			return nil, errors.New("session file exceeds 4096 bytes")
		}
		if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
			return nil, errors.New("session file permissions must be 0600")
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		opened, err := f.Stat()
		if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
			f.Close()
			return nil, errors.New("session file changed while opening")
		}
		raw, err := io.ReadAll(io.LimitReader(f, 4097))
		f.Close()
		if err != nil {
			return nil, errors.New("session file could not be read")
		}
		if len(raw) > 4096 || json.Unmarshal(raw, &c.session) != nil || !validToken(c.session.Token) {
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
	if !validToken(token) {
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
	next := Session{c.API, token}
	if err = json.NewEncoder(f).Encode(next); err != nil {
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
	if err = os.Rename(f.Name(), c.Config); err != nil {
		return err
	}
	c.session = next
	return nil
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
	budget := int64(32 * 1024 * 1024)
	failed := resp.StatusCode < 200 || resp.StatusCode >= 300
	if failed {
		budget = 4096
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, budget+1))
	if err != nil {
		return nil, errors.New("API response interrupted; check server health")
	}
	if int64(len(data)) > budget {
		return nil, errors.New("API response exceeds client budget")
	}
	var out map[string]any
	if strictjson.Decode(data, &out) != nil {
		return nil, errors.New("API returned invalid JSON")
	}
	if failed {
		code, _ := out["error"].(string)
		if !errorCode.MatchString(code) {
			code = "request_failed"
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, code)
	}
	if out == nil {
		return nil, errors.New("API returned an invalid response object")
	}
	return out, nil
}

var errorCode = regexp.MustCompile(`^[a-z][a-z0-9_]{0,127}$`)

func validToken(token string) bool {
	if len(token) != 64 {
		return false
	}
	_, err := hex.DecodeString(token)
	return err == nil
}
