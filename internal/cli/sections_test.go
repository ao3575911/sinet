package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sinet/internal/client"
)

func TestEmailHandoffRequiresApprovalAndPreservesDraftText(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == "GET" {
			if r.URL.Path != "/v1/email/drafts/draft-id" {
				t.Error("unexpected draft request")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"recipient": "friend@example.test", "subject": "Subject &bcc=bad@example.test", "body": "private body"})
			return
		}
		var value map[string]string
		if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
			t.Error(err)
		}
		if value["subject"] != "--limit" || value["body"] != "--offset" || r.URL.RawQuery != "" {
			t.Error("draft content interpreted as pagination flags", value)
		}
		_, _ = w.Write([]byte(`{"status":"draft"}`))
	}))
	defer server.Close()
	var output bytes.Buffer
	app := &App{C: &client.Client{API: server.URL, HTTP: server.Client()}, In: bufio.NewReader(strings.NewReader("")), Out: &output}
	if err := app.run([]string{"email", "gmail", "draft-id"}); err == nil || calls != 0 {
		t.Fatal("Gmail link prepared without explicit approval")
	}
	if err := app.run([]string{"email", "gmail", "draft-id", "--approve"}); err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(result["url"].(string))
	if err != nil {
		t.Fatal(err)
	}
	if result["sent"] != false || u.Host != "mail.google.com" || u.Query().Get("su") != "Subject &bcc=bad@example.test" || u.Query().Has("bcc") || calls != 1 {
		t.Fatal("unsafe handoff or false send claim")
	}
	if err := app.run([]string{"email", "new", "friend@example.test", "--limit", "--offset"}); err != nil {
		t.Fatal(err)
	}
}

func TestWalletSignatureComesFromStdin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var value map[string]string
		if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
			t.Error(err)
		}
		if r.URL.Path != "/v1/wallets/verify" || value["nonce"] != "challenge-nonce" || value["signature"] != "0xtest" {
			t.Error("signature stdin flow incorrect", value)
		}
		_, _ = w.Write([]byte(`{"status":"ownership_verified"}`))
	}))
	defer server.Close()
	var output bytes.Buffer
	app := &App{C: &client.Client{API: server.URL, HTTP: server.Client()}, In: bufio.NewReader(strings.NewReader("0xtest\n")), Out: &output}
	if err := app.run([]string{"wallet", "verify", "challenge-nonce"}); err != nil {
		t.Fatal(err)
	}
}
