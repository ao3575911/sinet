package model

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConfiguredAdapterAndFailureBoundaries(t *testing.T) {
	mode := "ok"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer fixture-only" {
			t.Error("provider contract invalid")
		}
		switch mode {
		case "ok":
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"real adapter fixture response"}}]}`))
		case "large":
			_, _ = w.Write([]byte(strings.Repeat("x", 32769)))
		case "failure":
			w.WriteHeader(500)
			_, _ = w.Write([]byte("upstream secret must never escape"))
		case "redirect":
			http.Redirect(w, r, "http://127.0.0.1:1", 302)
		}
	}))
	defer server.Close()
	p, err := New(server.URL+"/v1", "fixture-only", "test", true)
	if err != nil {
		t.Fatal(err)
	}
	if result, err := p.Chat(context.Background(), []Message{{"user", "hello"}}); err != nil || result != "real adapter fixture response" {
		t.Fatal("model adapter failed")
	}
	for _, v := range []string{"large", "failure", "redirect"} {
		mode = v
		_, err = p.Chat(context.Background(), nil)
		if err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatal("unsafe provider failure")
		}
	}
	if _, err = New("http://remote.example", "key", "test", false); err == nil {
		t.Fatal("insecure model endpoint accepted")
	}
	p, err = New("", "", "", false)
	if err != nil || p != nil {
		t.Fatal("no model behavior incorrect")
	}
}
