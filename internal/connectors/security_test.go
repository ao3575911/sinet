package connectors

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sinet/internal/githubpublic"
	"strings"
	"testing"
)

type brokenGatewayBody struct{ err error }

func (b brokenGatewayBody) Read([]byte) (int, error) { return 0, b.err }
func (b brokenGatewayBody) Close() error             { return nil }
func TestSecurityGatewayBodyFailures(t *testing.T) {
	for _, test := range []struct {
		err    error
		status int
		code   string
	}{{context.DeadlineExceeded, 504, "github_gateway_timeout"}, {io.ErrUnexpectedEOF, 502, "github_invalid_gateway_response"}} {
		executor := Executor{Transport: githubTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: brokenGatewayBody{test.err}}, nil
		})}
		m := fixedGitHubManifest()
		_, err := executor.Execute(context.Background(), m, m.Operations[0], json.RawMessage(githubpublic.Input))
		var failure *RemoteFailure
		if !errors.As(err, &failure) || failure.Status != test.status || failure.Code != test.code {
			t.Fatalf("misreported gateway error %v", err)
		}
	}
}
func TestSecurityDuplicateManifestAndArguments(t *testing.T) {
	m := fixedGitHubManifest()
	raw, _ := json.Marshal(m)
	raw = []byte(strings.Replace(string(raw), `"id":`, `"id":"evil", "id":`, 1))
	if _, err := Parse(raw); err == nil {
		t.Fatal("duplicate manifest key accepted")
	}
	op := Operation{Parameters: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"number":{"type":"integer"}}}`)}
	if err := op.Validate(json.RawMessage(`{"number":1,"number":2}`)); err == nil {
		t.Fatal("duplicate arguments accepted")
	}
}

func TestSecuritySchemaResourceLimits(t *testing.T) {
	for _, raw := range []string{
		`{"type":"object","additionalProperties":false,"$defs":{"x":{"type":"object"}},"$ref":"#/$defs/x"}`,
		`{"type":"object","additionalProperties":false,"properties":{"a":{"type":"number","minimum":1e-999999999}}}`,
		`{"type":"object","additionalProperties":false,"allOf":[` + strings.Repeat(`{"type":"object"},`, 300) + `{"type":"object"}]}`,
	} {
		op := Operation{Parameters: json.RawMessage(raw)}
		if err := op.Validate(json.RawMessage(`{}`)); err == nil {
			t.Fatal("unbounded schema accepted")
		}
	}
}
