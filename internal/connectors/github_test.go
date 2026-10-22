package connectors

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"sinet/contracts"
	"sinet/internal/githubpublic"
)

type githubTransport func(*http.Request) (*http.Response, error)

func (f githubTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func fixedGitHubManifest() Manifest {
	return Manifest{ID: "github-public", Version: "1.0.0", Publisher: "alice", Name: "GitHub public", Description: "Read-only", Runtime: "remote_https", EndpointURL: "https://gateway.example.com", PrivacyPolicyURL: "https://gateway.example.com/privacy", Operations: []Operation{{Name: "repository_summary", Description: "Read only", Risk: "external_disclosure", Parameters: json.RawMessage(contracts.GitHubParameters)}}}
}
func TestGitHubManifestCannotBroadenContract(t *testing.T) {
	for _, mutate := range []func(*Manifest){
		func(m *Manifest) { m.Runtime = "builtin"; m.EndpointURL = "" },
		func(m *Manifest) { m.Operations[0].Risk = "read" },
		func(m *Manifest) { m.Operations[0].Name = "anything" },
		func(m *Manifest) { m.Operations = append(m.Operations, m.Operations[0]) },
		func(m *Manifest) {
			m.Operations[0].Parameters = json.RawMessage(`{"type":"object","additionalProperties":false}`)
		},
	} {
		m := fixedGitHubManifest()
		mutate(&m)
		raw, _ := json.Marshal(m)
		if _, err := Parse(raw); err == nil {
			t.Fatal("broadened GitHub manifest accepted")
		}
	}
	m := fixedGitHubManifest()
	raw, _ := json.Marshal(m)
	if _, err := Parse(raw); err != nil {
		t.Fatal(err)
	}
}
func TestGitHubAdapterRevalidatesGatewayOutput(t *testing.T) {
	raw, err := os.ReadFile("../githubconnector/testdata/repository.json")
	if err != nil {
		t.Fatal(err)
	}
	summary, err := githubpublic.ParseUpstream(raw)
	if err != nil {
		t.Fatal(err)
	}
	valid, _ := json.Marshal(summary)
	for name, body := range map[string]string{
		"valid":       string(valid),
		"unknown":     strings.TrimSuffix(string(valid), "}") + `,"private":false}`,
		"duplicate":   strings.Replace(string(valid), `"full_name":`, `"source":"wrong","full_name":`, 1),
		"control":     strings.Replace(string(valid), "Synthetic", `\u001bSynthetic`, 1),
		"private_url": strings.Replace(string(valid), githubpublic.PublicURL, "https://127.0.0.1/private", 1),
		"oversized":   strings.Repeat("x", 65537),
		"malformed":   `{"output":`,
	} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			executor := Executor{Transport: githubTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}
			m := fixedGitHubManifest()
			out, err := executor.Execute(context.Background(), m, m.Operations[0], json.RawMessage(githubpublic.Input))
			if calls != 1 || (err == nil) != (name == "valid") {
				t.Fatalf("invalid result: %s err=%v", out, err)
			}
		})
	}
}
func TestGitHubAdapterDeniesInputsBeforeTransport(t *testing.T) {
	for _, args := range []string{`{"owner":"bad","owner":"ao3575911","repository":"sinet"}`, `{"owner":"ao3575911","repository":"private"}`, `{"owner":"ao3575911","repository":"sinet","token":"private"}`, strings.Repeat(" ", 513) + githubpublic.Input} {
		executor := Executor{Transport: githubTransport(func(*http.Request) (*http.Response, error) {
			t.Fatal("invalid input reached transport")
			return nil, nil
		})}
		m := fixedGitHubManifest()
		if _, err := executor.Execute(context.Background(), m, m.Operations[0], json.RawMessage(args)); err == nil {
			t.Fatal("invalid input accepted")
		}
	}
}
