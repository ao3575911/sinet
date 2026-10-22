package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"context"
	"sinet/contracts"
	"sinet/internal/cli"
	"sinet/internal/client"
	"sinet/internal/connectors"
	"sinet/internal/githubconnector"
	"sinet/internal/githubpublic"
	"sinet/internal/operator"
)

type fixtureTransport func(*http.Request) (*http.Response, error)

func (f fixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func githubManifest() map[string]any {
	return map[string]any{"id": "github-public", "version": "1.0.0", "publisher": "alice", "name": "Public GitHub", "description": "Synthetic local authorization fixture; not a deployed service", "runtime": "remote_https", "endpoint_url": "https://gateway.example.com", "privacy_policy_url": "https://gateway.example.com/privacy", "operations": []any{map[string]any{"name": "repository_summary", "description": "Send owner ao3575911 and repository sinet to this gateway and GitHub for read-only public repository facts; no account data.", "risk": "external_disclosure", "parameters": json.RawMessage(contracts.GitHubParameters)}}}
}

func prepareGitHub(t *testing.T) (*fixture, string, string) {
	t.Helper()
	f := setup(t)
	a, b := f.login(t, "a@example.test"), f.login(t, "b@example.test")
	f.claim(t, a, "alice")
	f.claim(t, b, "bob")
	code, out := f.request("POST", "/v1/connectors/submissions", a, githubManifest(), "")
	expect(t, 200, code, out)
	code, out = f.request("POST", "/v1/connectors/github-public/install", a, map[string]string{"version": "1.0.0"}, "")
	expect(t, 403, code, out)
	// Operator attestation applies ONLY inside this disposable test schema. This is
	// synthetic ownership/review evidence, never a real connector approval.
	if err := operator.Run(context.Background(), f.db, []string{"review", "approve", "github-public", "1.0.0", "ownership:synthetic-fixture review:isolated-test-only", "--confirm"}, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{a, b} {
		code, out = f.request("POST", "/v1/connectors/github-public/install", token, map[string]string{"version": "1.0.0"}, "")
		expect(t, 200, code, out)
	}
	return f, a, b
}
func githubTool() map[string]any {
	return map[string]any{"connector": "github-public", "version": "1.0.0", "operation": "repository_summary", "arguments": json.RawMessage(githubpublic.Input)}
}
func grantGitHub(t *testing.T, f *fixture, token string) {
	t.Helper()
	code, out := f.request("POST", "/v1/connectors/github-public/grants", token, map[string]any{"version": "1.0.0", "operation": "repository_summary", "approve": true}, "")
	expect(t, 200, code, out)
}
func consentGitHub(t *testing.T, f *fixture, token string) string {
	t.Helper()
	code, out := f.request("POST", "/v1/agent/consents", token, map[string]any{"tool": githubTool(), "approve": true}, "")
	expect(t, 200, code, out)
	recipients := out["recipients"].([]any)
	if len(recipients) != 2 || recipients[0] != "https://gateway.example.com" || recipients[1] != "https://api.github.com" || out["expires_in"] != float64(300) || !strings.Contains(out["disclosure"].(string), "No SINet session") {
		t.Fatal("incomplete consent disclosure")
	}
	return out["consent"].(string)
}
func fixtureGitHubPipeline(t *testing.T, f *fixture, upstream http.RoundTripper) *atomic.Int32 {
	t.Helper()
	calls := &atomic.Int32{}
	gateway := githubconnector.NewHandlerWithTransport(fixtureTransport(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		if r.URL.String() != githubpublic.Upstream || r.Method != "GET" || r.Body != nil {
			t.Error("upstream request changed")
		}
		for key := range r.Header {
			if key != "Accept" && key != "User-Agent" && key != "X-Github-Api-Version" {
				t.Errorf("private upstream header %s", key)
			}
		}
		return upstream.RoundTrip(r)
	}))
	f.s.executeTool = connectors.Executor{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://gateway.example.com/repository_summary" {
			t.Error("gateway endpoint changed")
		}
		for key := range r.Header {
			if key != "Accept" && key != "Content-Type" {
				t.Errorf("sensitive gateway header %s", key)
			}
		}
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != githubpublic.Input {
			t.Error("private/changed request body")
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		w := httptest.NewRecorder()
		gateway.ServeHTTP(w, r)
		return w.Result(), nil
	})}.Execute
	return calls
}
func syntheticGitHub(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("../githubconnector/testdata/repository.json")
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
func upstreamJSON(body string, status int) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}
}

func TestGitHubAuthorizationJourneyCountedUpstream(t *testing.T) {
	f, a, b := prepareGitHub(t)
	body := syntheticGitHub(t)
	calls := fixtureGitHubPipeline(t, f, fixtureTransport(func(*http.Request) (*http.Response, error) { return upstreamJSON(body, 200), nil }))
	// Private account data exists but is absent at both outbound boundaries.
	for _, token := range []string{a, b} {
		code, out := f.request("POST", "/v1/notes", token, map[string]string{"title": "private", "body": "private-note-sentinel"}, "")
		expect(t, 200, code, out)
	}
	invoke := func(token string, tool map[string]any, want int) {
		t.Helper()
		before := calls.Load()
		code, out := f.request("POST", "/v1/agent/messages", token, map[string]any{"tool": tool}, "")
		expect(t, want, code, out)
		if want != 200 && calls.Load() != before {
			t.Fatal("denied request reached upstream")
		}
	}
	invoke(a, githubTool(), 403) // missing grant
	code, out := f.request("POST", "/v1/agent/consents", a, map[string]any{"tool": githubTool(), "approve": true}, "")
	expect(t, 403, code, out)
	grantGitHub(t, f, a)
	grantGitHub(t, f, b)
	invoke(a, githubTool(), 403) // missing consent
	token := consentGitHub(t, f, a)
	tool := githubTool()
	tool["consent"] = token
	invoke(b, tool, 403) // both accounts installed/granted; consent alone crosses accounts
	for _, changed := range []struct {
		field  string
		value  any
		status int
	}{{"version", "9.9.9", 403}, {"operation", "write_repository", 400}, {"arguments", json.RawMessage(`{"owner":"ao3575911","repository":"other"}`), 400}, {"arguments", json.RawMessage(`{"owner":"ao3575911","repository":"sinet","email":"secret"}`), 400}} {
		invalid := githubTool()
		invalid["consent"] = token
		invalid[changed.field] = changed.value
		invoke(a, invalid, changed.status)
	}
	if calls.Load() != 0 {
		t.Fatal("denial phase called upstream")
	}
	invoke(a, tool, 200)
	if calls.Load() != 1 {
		t.Fatal("success did not invoke once")
	}
	invoke(a, tool, 403) // new idempotency key cannot replay one-use consent
	expired := consentGitHub(t, f, a)
	if _, err := f.db.Exec(`UPDATE consents SET expires_at=now()-interval '1 second' WHERE token_hash=$1`, hash(expired)); err != nil {
		t.Fatal(err)
	}
	tool["consent"] = expired
	invoke(a, tool, 403)
	pending := consentGitHub(t, f, a)
	code, out = f.request("DELETE", "/v1/connectors/github-public/grants/repository_summary", a, nil, "")
	expect(t, 200, code, out)
	tool["consent"] = pending
	invoke(a, tool, 403)
	grantGitHub(t, f, a)
	invoke(a, tool, 403) // revoke destroyed even unconsumed approvals
	tool["consent"] = consentGitHub(t, f, a)
	code, out = f.request("DELETE", "/v1/connectors/github-public/install", a, nil, "")
	expect(t, 200, code, out)
	invoke(a, tool, 403)
	code, out = f.request("GET", "/v1/audit/tool-calls", a, nil, "")
	expect(t, 200, code, out)
	successes := 0
	for _, item := range out["items"].([]any) {
		entry := item.(map[string]any)
		if entry["outcome"] == "success" {
			successes++
			t.Logf("synthetic audit id=%v connector=%v version=%v operation=%v allowed=%v outcome=%v occurred_at=%v", entry["id"], entry["connector"], entry["version"], entry["operation"], entry["allowed"], entry["outcome"], entry["occurred_at"])
			if entry["allowed"] != true {
				t.Fatal("success not authorized")
			}
		}
	}
	if successes != 1 {
		t.Fatal("missing audit success")
	}
	code, out = f.request("GET", "/v1/audit/tool-calls", b, nil, "")
	expect(t, 200, code, out)
	if len(out["items"].([]any)) != 1 || out["items"].([]any)[0].(map[string]any)["allowed"] != false {
		t.Fatal("audit crossed accounts")
	}
	code, out = f.request("GET", "/v1/agent/messages", b, nil, "")
	expect(t, 200, code, out)
	if len(out["items"].([]any)) != 0 {
		t.Fatal("tool result entered model history")
	}
	ownTool := githubTool()
	ownTool["consent"] = consentGitHub(t, f, b)
	invoke(b, ownTool, 200)
	if calls.Load() != 2 {
		t.Fatal("second account could not independently invoke")
	}
	t.Log("synthetic counted journey: 2 independently authorized upstream calls; all grant/consent/binding/expiry/replay/revoke/uninstall denials made zero additional calls")
}

func TestGitHubConcurrentRevokeSerializesWithInvocation(t *testing.T) {
	f, a, _ := prepareGitHub(t)
	grantGitHub(t, f, a)
	body := syntheticGitHub(t)
	started, release := make(chan struct{}), make(chan struct{})
	var released atomic.Bool
	t.Cleanup(func() {
		if released.CompareAndSwap(false, true) {
			close(release)
		}
	})
	calls := fixtureGitHubPipeline(t, f, fixtureTransport(func(*http.Request) (*http.Response, error) {
		close(started)
		<-release
		return upstreamJSON(body, 200), nil
	}))
	tool := githubTool()
	tool["consent"] = consentGitHub(t, f, a)
	finished := make(chan int, 1)
	go func() {
		code, _ := f.request("POST", "/v1/agent/messages", a, map[string]any{"tool": tool}, "")
		finished <- code
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("invocation never reached fixture")
	}
	revoked := make(chan int, 1)
	go func() {
		code, _ := f.request("DELETE", "/v1/connectors/github-public/grants/repository_summary", a, nil, "")
		revoked <- code
	}()
	// Observe the database lock waiter, rather than relying on scheduling sleeps.
	deadline := time.Now().Add(5 * time.Second)
	for {
		var waiting bool
		err := f.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_locks WHERE locktype='advisory' AND NOT granted)`).Scan(&waiting)
		if err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("revoke did not wait for authorized invocation")
		}
		time.Sleep(5 * time.Millisecond)
	}
	select {
	case <-revoked:
		t.Fatal("revoke returned before active call finished")
	default:
	}
	if released.CompareAndSwap(false, true) {
		close(release)
	}
	if <-finished != 200 || <-revoked != 200 {
		t.Fatal("serialized call/revoke failed")
	}
	code, out := f.request("POST", "/v1/agent/messages", a, map[string]any{"tool": tool}, "")
	expect(t, 403, code, out)
	if calls.Load() != 1 {
		t.Fatal("post-revoke upstream call")
	}
	t.Log("revoke waits behind the account transaction lock; an earlier authorized call completes, all later calls are denied")
}

func TestGitHubCLIThroughPersonalAgent(t *testing.T) {
	f, a, _ := prepareGitHub(t)
	grantGitHub(t, f, a)
	body := syntheticGitHub(t)
	calls := fixtureGitHubPipeline(t, f, fixtureTransport(func(*http.Request) (*http.Response, error) { return upstreamJSON(body, 200), nil }))
	server := httptest.NewServer(f.s)
	defer server.Close()
	t.Setenv("SINET_API", server.URL)
	t.Setenv("SINET_CONFIG", t.TempDir()+"/session.json")
	c, err := client.New()
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Save(a); err != nil {
		t.Fatal(err)
	}
	input, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	var output bytes.Buffer
	err = cli.Run(c, []string{"agent", "github"}, input, &output)
	if err == nil || calls.Load() != 0 || !strings.Contains(output.String(), "api.github.com") || !strings.Contains(output.String(), "gateway.example.com") {
		t.Fatal("preview did not disclose both recipients or contacted upstream")
	}
	output.Reset()
	if err = cli.Run(c, []string{"agent", "github", "--approve"}, input, &output); err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if json.Unmarshal(output.Bytes(), &result) != nil || result["mode"] != "deterministic_tool" || calls.Load() != 1 {
		t.Fatal("CLI agent result missing")
	}
	if !strings.Contains(output.String(), "Untrusted public GitHub facts") {
		t.Fatal("missing untrusted-data label")
	}
	t.Log("actual CLI preview and approved invocation crossed API, adapter and gateway using synthetic upstream facts")
}

func TestGitHubFailureJourneyIsAuditedAndConsumesConsent(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
		err    error
		want   int
		code   string
	}{
		{"rate_limit", 429, `{"message":"secret"}`, nil, 503, "github_upstream_rate_limited"},
		{"timeout", 0, "", context.DeadlineExceeded, 504, "github_upstream_timeout"},
		{"upstream_error", 500, `{"secret":"do not disclose"}`, nil, 502, "github_upstream_failure"},
		{"private", 200, `{"private":true}`, nil, 502, "github_invalid_upstream_response"},
		{"oversized", 200, strings.Repeat("x", 65537), nil, 502, "github_upstream_response_limit"},
		{"hostile", 200, `{"description":"\u001b[2J"}`, nil, 502, "github_invalid_upstream_response"},
	} {
		t.Run(test.name, func(t *testing.T) {
			f, a, _ := prepareGitHub(t)
			grantGitHub(t, f, a)
			calls := fixtureGitHubPipeline(t, f, fixtureTransport(func(*http.Request) (*http.Response, error) {
				if test.err != nil {
					return nil, test.err
				}
				return upstreamJSON(test.body, test.status), nil
			}))
			tool := githubTool()
			tool["consent"] = consentGitHub(t, f, a)
			code, out := f.request("POST", "/v1/agent/messages", a, map[string]any{"tool": tool}, "")
			expect(t, test.want, code, out)
			if out["error"] != test.code || calls.Load() != 1 || strings.Contains(fmt.Sprint(out), "secret") {
				t.Fatalf("unsafe or untruthful failure %v", out)
			}
			code, out = f.request("POST", "/v1/agent/messages", a, map[string]any{"tool": tool}, "")
			expect(t, 403, code, out)
			if calls.Load() != 1 {
				t.Fatal("failed consent was replayed")
			}
			code, out = f.request("GET", "/v1/audit/tool-calls", a, nil, "")
			expect(t, 200, code, out)
			if len(out["items"].([]any)) != 2 || out["items"].([]any)[1].(map[string]any)["outcome"] != "execution_failed" {
				t.Fatal("failure audit missing")
			}
		})
	}
}

func TestGitHubConsentExpiresWhileWaitingForAccountLock(t *testing.T) {
	f, a, _ := prepareGitHub(t)
	grantGitHub(t, f, a)
	token := consentGitHub(t, f, a)
	var account string
	if err := f.db.QueryRow(`SELECT account_id FROM sessions WHERE token_hash=$1`, hash(a)).Scan(&account); err != nil {
		t.Fatal(err)
	}
	lock, err := f.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err = lock.Exec(`SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, account); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec(`UPDATE consents SET expires_at=clock_timestamp()+interval '200 milliseconds' WHERE token_hash=$1`, hash(token)); err != nil {
		t.Fatal(err)
	}
	body := syntheticGitHub(t)
	calls := fixtureGitHubPipeline(t, f, fixtureTransport(func(*http.Request) (*http.Response, error) { return upstreamJSON(body, 200), nil }))
	tool := githubTool()
	tool["consent"] = token
	done := make(chan int, 1)
	go func() {
		code, _ := f.request("POST", "/v1/agent/messages", a, map[string]any{"tool": tool}, "")
		done <- code
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var waiting bool
		if err = f.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_locks WHERE locktype='advisory' AND NOT granted)`).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("call never waited on account lock")
		}
		time.Sleep(time.Millisecond)
	}
	for {
		var expired bool
		if err = f.db.QueryRow(`SELECT expires_at<=clock_timestamp() FROM consents WHERE token_hash=$1`, hash(token)).Scan(&expired); err != nil {
			t.Fatal(err)
		}
		if expired {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("consent never expired")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err = lock.Commit(); err != nil {
		t.Fatal(err)
	}
	if <-done != 403 || calls.Load() != 0 {
		t.Fatal("consent expired in lock queue reached upstream")
	}
}

// A child process uses the actual CLI with a POSIX terminal. It connects only to
// this test's API; neither the gateway nor GitHub is a live external service.
func TestGitHubPTYCLIHelper(t *testing.T) {
	if os.Getenv("SINET_GITHUB_PTY_HELPER") != "1" {
		return
	}
	c, err := client.New()
	if err == nil {
		err = cli.Run(c, nil, os.Stdin, os.Stdout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
func TestGitHubTUIThroughPersonalAgent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX PTY only")
	}
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 required for POSIX PTY")
	}
	f, a, _ := prepareGitHub(t)
	grantGitHub(t, f, a)
	body := syntheticGitHub(t)
	calls := fixtureGitHubPipeline(t, f, fixtureTransport(func(*http.Request) (*http.Response, error) { return upstreamJSON(body, 200), nil }))
	server := httptest.NewServer(f.s)
	defer server.Close()
	t.Setenv("SINET_API", server.URL)
	t.Setenv("SINET_CONFIG", t.TempDir()+"/session.json")
	c, err := client.New()
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Save(a); err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	const driver = `import os,pty,select,subprocess,sys,time
master,slave=pty.openpty()
env=dict(os.environ,SINET_GITHUB_PTY_HELPER='1')
proc=subprocess.Popen([sys.argv[1],'-test.run=^TestGitHubPTYCLIHelper$'],stdin=slave,stdout=slave,stderr=slave,env=env)
os.close(slave)
seen=b''
def expect(text):
 global seen
 target=text.encode();deadline=time.monotonic()+15
 while target not in seen:
  if time.monotonic()>deadline:raise AssertionError('PTY timed out: '+text)
  ready,_,_=select.select([master],[],[],0.1)
  if ready:seen+=os.read(master,8192)
 _,seen=seen.split(target,1)
def send(text):os.write(master,text.encode())
try:
 expect('Your network');send('5');expect('GitHub public facts and audit');send('8')
 expect('GitHub public: ao3575911/sinet');send('4')
 expect('gateway.example.com');expect('api.github.com');expect('Type invoke')
 send('cancel\n');expect('Your network')
 send('5');expect('GitHub public facts and audit');send('8')
 expect('GitHub public: ao3575911/sinet');send('4')
 expect('gateway.example.com');expect('api.github.com');expect('Type invoke')
 send('invoke\n');expect('Untrusted public GitHub facts');expect('Your network');send('q')
 proc.wait(timeout=5);assert proc.returncode==0
 print('PASS synthetic GitHub PTY: both recipients, cancel, exact one-call approval, validated facts')
finally:
 if proc.poll() is None:proc.kill();proc.wait()
 os.close(master)
`
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	result, err := exec.CommandContext(ctx, python, "-c", driver, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("GitHub PTY failed: %v %s", err, result)
	}
	if calls.Load() != 1 {
		t.Fatal("cancel or invocation count incorrect")
	}
	t.Log(strings.TrimSpace(string(result)))
}
