package githubconnector

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type brokenBody struct{ err error }

func (b brokenBody) Read([]byte) (int, error) { return 0, b.err }
func (b brokenBody) Close() error             { return nil }
func TestSecurityInterruptedBodyFailures(t *testing.T) {
	for _, test := range []struct {
		err    error
		status int
		code   string
	}{{context.DeadlineExceeded, 504, "upstream_timeout"}, {io.ErrUnexpectedEOF, 502, "upstream_failure"}} {
		client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: brokenBody{test.err}}, nil
		})}
		result := request(newHandler(client), input)
		if result.Code != test.status || !strings.Contains(result.Body.String(), test.code) {
			t.Fatalf("misreported interrupted upstream: %d %s", result.Code, result.Body.String())
		}
	}
}
