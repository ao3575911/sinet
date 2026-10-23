//go:build darwin || windows

package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These tests deliberately use the real OS store, including in native CI.
func TestNativeCredentialLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	t.Setenv("SINET_CONFIG", path)
	t.Setenv("SINET_API", "https://api.example.test")
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Forget() })
	token := strings.Repeat("ab", 32)
	if err := c.Save(token); err != nil {
		t.Fatal(err)
	}
	first := c.session
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), token) || strings.Contains(string(raw), `"token"`) {
		t.Fatal("token persisted outside OS store")
	}
	reopened, err := New()
	if err != nil || !reopened.Authenticated() {
		t.Fatalf("reload: %v", err)
	}
	if err := c.Save(strings.Repeat("cd", 32)); err != nil {
		t.Fatal(err)
	}
	if _, err := readCredential(path, first); err == nil {
		t.Fatal("replaced credential remains")
	}
	// Changing metadata cannot rebind a token to another server.
	changed := c.session
	changed.API = "https://other.example.test"
	if _, err := readCredential(path, changed); err == nil {
		t.Fatal("origin rebinding accepted")
	}
	if _, err := readCredential(path+"-copy", c.session); err == nil {
		t.Fatal("profile copying accepted")
	}
	last := c.session
	if err := c.Forget(); err != nil {
		t.Fatal(err)
	}
	if _, err := readCredential(path, last); err == nil {
		t.Fatal("forgotten OS credential remains")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("session file remains")
	}
}

func TestNativeLegacyMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	t.Setenv("SINET_CONFIG", path)
	t.Setenv("SINET_API", "https://other.example.test")
	legacy := Session{API: "https://api.example.test", Token: strings.Repeat("ab", 32)}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := protectSessionFile(f); err != nil {
		f.Close()
		t.Fatal(err)
	}
	err = json.NewEncoder(f).Encode(legacy)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	c, err := New()
	if err != nil || c.Authenticated() {
		t.Fatalf("legacy origin binding: %v", err)
	}
	if c.session.Credential != "" {
		t.Fatal("migrated to wrong origin")
	}
	t.Setenv("SINET_API", legacy.API)
	c, err = New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Forget() })
	if !c.Authenticated() || c.session.Credential == "" {
		t.Fatal("legacy session not migrated")
	}
	raw, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(raw), legacy.Token) {
		t.Fatal("plaintext remains after migration")
	}
	// A downgrade must reject the new format rather than bypass native storage.
	var old struct{ API, Token string }
	if err := json.Unmarshal(raw, &old); err != nil || validToken(old.Token) {
		t.Fatal("unsafe downgrade")
	}
}

func TestNativeCredentialSymlinkDenied(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	// CI enables Windows Developer Mode so this is tested, never skipped.
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SINET_CONFIG", path)
	t.Setenv("SINET_API", "https://api.example.test")
	if _, err := New(); err == nil {
		t.Fatal("loaded symlink")
	}
	c := &Client{API: "https://api.example.test", Config: path}
	if err := c.Save(strings.Repeat("ab", 32)); err == nil {
		t.Fatal("saved through symlink")
	}
	raw, err := os.ReadFile(target)
	if err != nil || string(raw) != "untouched" {
		t.Fatal("symlink target modified")
	}
}
