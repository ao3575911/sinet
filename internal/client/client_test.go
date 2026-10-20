package client

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCredentialStorageOriginAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile", "session.json")
	t.Setenv("SINET_CONFIG", path)
	received := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Get("Authorization") != ""
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("SINET_API", server.URL)
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	token := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err = c.Save(token); err != nil {
		t.Fatal(err)
	}
	c, err = New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.Do("GET", "/test", nil); err != nil || !received {
		t.Fatal("saved credential not used for bound origin")
	}
	t.Setenv("SINET_API", "http://127.0.0.1:1")
	c, err = New()
	if err != nil {
		t.Fatal(err)
	}
	if c.Authenticated() {
		t.Fatal("credential reused for changed origin")
	}
	if runtime.GOOS != "windows" {
		if err = os.Chmod(path, 0644); err != nil {
			t.Fatal(err)
		}
		if _, err = New(); err == nil {
			t.Fatal("world-readable credential accepted")
		}
	}
}
func TestSymlinkAndInsecureOriginDenied(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SINET_CONFIG", path)
	t.Setenv("SINET_API", "http://remote.example")
	if _, err := New(); err == nil {
		t.Fatal("insecure remote origin accepted")
	}
	if runtime.GOOS != "windows" {
		if err := os.Symlink(target, path); err != nil {
			t.Fatal(err)
		}
		if _, err := New(); err == nil {
			t.Fatal("symlink session accepted")
		}
	}
}
