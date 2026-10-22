package api

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"io"
	"sinet/internal/connectors"
	"sinet/internal/operator"
)

func waitSecurityCondition(t *testing.T, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !check() {
		if time.Now().After(deadline) {
			t.Fatal("database condition did not become true")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestSecuritySessionExpiryAfterLockWait(t *testing.T) {
	f := setup(t)
	token := f.login(t, "expiry@example.test")
	var account string
	if err := f.db.QueryRow(`SELECT account_id FROM sessions WHERE token_hash=$1`, hash(token)).Scan(&account); err != nil {
		t.Fatal(err)
	}
	lock, err := f.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err := lock.Exec(`SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, account); err != nil {
		t.Fatal(err)
	}
	done := make(chan int, 1)
	go func() {
		code, _ := f.request("POST", "/v1/notes", token, map[string]string{"title": "must not persist", "body": "expired session"}, "")
		done <- code
	}()
	waitSecurityCondition(t, func() bool {
		var waiting bool
		if err := f.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_locks WHERE locktype='advisory' AND NOT granted)`).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		return waiting
	})
	if _, err := f.db.Exec(`UPDATE sessions SET expires_at=clock_timestamp()+interval '100 milliseconds' WHERE token_hash=$1`, hash(token)); err != nil {
		t.Fatal(err)
	}
	waitSecurityCondition(t, func() bool {
		var expired bool
		if err := f.db.QueryRow(`SELECT expires_at<=clock_timestamp() FROM sessions WHERE token_hash=$1`, hash(token)).Scan(&expired); err != nil {
			t.Fatal(err)
		}
		return expired
	})
	if err := lock.Commit(); err != nil {
		t.Fatal(err)
	}
	if code := <-done; code != 401 {
		t.Fatalf("expired queued session accepted: HTTP %d", code)
	}
	var count int
	if err := f.db.QueryRow(`SELECT count(*) FROM notes`).Scan(&count); err != nil || count != 0 {
		t.Fatal("expired session wrote data", err)
	}
}

func TestSecurityConsentExactNumbers(t *testing.T) {
	f := setup(t)
	token := f.login(t, "numbers@example.test")
	f.claim(t, token, "alice")
	manifest := githubManifest()
	manifest["id"] = "number-fixture"
	manifest["operations"] = []any{map[string]any{"name": "inspect_number", "description": "Isolated numeric consent fixture", "risk": "external_disclosure", "parameters": json.RawMessage(`{"type":"object","additionalProperties":false,"required":["value"],"properties":{"value":{"type":"integer"}}}`)}}
	code, out := f.request("POST", "/v1/connectors/submissions", token, manifest, "")
	expect(t, 200, code, out)
	if err := operator.Run(context.Background(), f.db, []string{"review", "approve", "number-fixture", "1.0.0", "ownership:synthetic review:isolated fixture", "--confirm"}, io.Discard); err != nil {
		t.Fatal(err)
	}
	code, out = f.request("POST", "/v1/connectors/number-fixture/install", token, map[string]string{"version": "1.0.0"}, "")
	expect(t, 200, code, out)
	code, out = f.request("POST", "/v1/connectors/number-fixture/grants", token, map[string]any{"version": "1.0.0", "operation": "inspect_number", "approve": true}, "")
	expect(t, 200, code, out)
	var calls atomic.Int32
	f.s.executeTool = func(_ context.Context, _ connectors.Manifest, _ connectors.Operation, args json.RawMessage) (json.RawMessage, error) {
		calls.Add(1)
		return args, nil
	}
	tool := map[string]any{"connector": "number-fixture", "version": "1.0.0", "operation": "inspect_number", "arguments": json.RawMessage(`{"value":9007199254740992}`)}
	code, out = f.request("POST", "/v1/agent/consents", token, map[string]any{"tool": tool, "approve": true}, "")
	expect(t, 200, code, out)
	tool["consent"] = out["consent"]
	tool["arguments"] = json.RawMessage(`{"value":9007199254740993}`)
	code, out = f.request("POST", "/v1/agent/messages", token, map[string]any{"tool": tool}, "")
	expect(t, 403, code, out)
	if calls.Load() != 0 {
		t.Fatal("changed number executed using another number's consent")
	}
	tool["arguments"] = json.RawMessage(`{ "value" : 9007199254740992 }`)
	code, out = f.request("POST", "/v1/agent/messages", token, map[string]any{"tool": tool}, "")
	expect(t, 200, code, out)
	if calls.Load() != 1 {
		t.Fatal("original exact number did not execute once")
	}
}

func TestSecurityConcurrentConnectorOwnership(t *testing.T) {
	f := setup(t)
	a, b := f.login(t, "a@example.test"), f.login(t, "b@example.test")
	f.claim(t, a, "alice")
	f.claim(t, b, "bob")
	lock, err := f.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err = lock.Exec(`LOCK TABLE connectors IN SHARE MODE`); err != nil {
		t.Fatal(err)
	}
	done := make(chan int, 2)
	for i, token := range []string{a, b} {
		m := githubManifest()
		m["version"] = []string{"1.0.0", "1.0.1"}[i]
		m["publisher"] = []string{"alice", "bob"}[i]
		go func(token string, m map[string]any) {
			code, _ := f.request("POST", "/v1/connectors/submissions", token, m, "")
			done <- code
		}(token, m)
	}
	waitSecurityCondition(t, func() bool {
		var count int
		if err := f.db.QueryRow(`SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND cardinality(pg_blocking_pids(pid))>0`).Scan(&count); err != nil {
			t.Fatal(err)
		}
		return count >= 2
	})
	if err := lock.Commit(); err != nil {
		t.Fatal(err)
	}
	first, second := <-done, <-done
	if !((first == 200 && second == 409) || (first == 409 && second == 200)) {
		t.Fatalf("multiple publishers claimed ID: %d %d", first, second)
	}
	var owners int
	if err := f.db.QueryRow(`SELECT count(DISTINCT publisher_id) FROM connectors WHERE id='github-public'`).Scan(&owners); err != nil || owners != 1 {
		t.Fatal("connector namespace ownership split", err)
	}
}

func TestSecurityAmbiguousRequestsDenied(t *testing.T) {
	f := setup(t)
	token := f.login(t, "json@example.test")
	for _, raw := range []string{`{"title":"first","title":"second"}`, `{"title":"x","body":"a","body":"b"}`} {
		code, out := f.request("POST", "/v1/notes", token, json.RawMessage(raw), "")
		expect(t, 400, code, out)
	}
}

func TestSecurityDatabaseFailuresAreNotAuthenticationOrRateLimits(t *testing.T) {
	t.Run("session lookup", func(t *testing.T) {
		f := setup(t)
		token := f.login(t, "lookup@example.test")
		if _, err := f.db.Exec(`ALTER TABLE sessions RENAME TO unavailable_sessions`); err != nil {
			t.Fatal(err)
		}
		code, out := f.request("GET", "/v1/me", token, nil, "")
		expect(t, 503, code, out)
		if out["error"] != "database_unavailable" {
			t.Fatal("misclassified session lookup failure")
		}
	})
	t.Run("rate counter", func(t *testing.T) {
		f := setup(t)
		token := f.login(t, "rate@example.test")
		if _, err := f.db.Exec(`ALTER TABLE rate_limits RENAME TO unavailable_limits`); err != nil {
			t.Fatal(err)
		}
		code, out := f.request("POST", "/v1/notes", token, map[string]string{"title": "x"}, "")
		expect(t, 503, code, out)
		if out["error"] != "database_unavailable" {
			t.Fatal("misclassified rate storage failure")
		}
	})
}
