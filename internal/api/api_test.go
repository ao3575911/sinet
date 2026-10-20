package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"sinet/internal/store"
	"sinet/migrations"
)

type mailbox struct {
	mu    sync.Mutex
	codes map[string]string
}

func (m *mailbox) Send(_ context.Context, email, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[email] = code
	return nil
}
func (m *mailbox) code(email string) string { m.mu.Lock(); defer m.mu.Unlock(); return m.codes[email] }

type fixture struct {
	s    *Server
	db   *sql.DB
	mail *mailbox
}

func setup(t *testing.T) *fixture {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL required for PostgreSQL integration tests")
	}
	ctx := context.Background()
	admin, err := store.Open(ctx, raw)
	if err != nil {
		t.Fatal("test database unavailable")
	}
	schema := "test_" + random(8)
	if _, err = admin.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := store.Open(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	if err = migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	m := &mailbox{codes: map[string]string{}}
	s := New(db, m, nil, "http://127.0.0.1:8080")
	t.Cleanup(func() { db.Close(); _, _ = admin.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); admin.Close() })
	return &fixture{s, db, m}
}
func (f *fixture) request(method, path, token string, body any, key string) (int, map[string]any) {
	var b []byte
	if body != nil {
		b = jsonBytes(body)
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(b))
	r.RemoteAddr = "192.0.2.1:1000"
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
		if key == "" {
			key = random(16)
		}
		r.Header.Set("Idempotency-Key", key)
	}
	w := httptest.NewRecorder()
	f.s.ServeHTTP(w, r)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}
func expect(t *testing.T, want, got int, out map[string]any) {
	t.Helper()
	if got != want {
		t.Fatalf("expected HTTP %d got %d error=%v", want, got, out["error"])
	}
}
func (f *fixture) login(t *testing.T, email string) string {
	t.Helper()
	status, out := f.request("POST", "/v1/auth/request-otp", "", map[string]string{"email": email}, "")
	expect(t, 200, status, out)
	status, out = f.request("POST", "/v1/auth/verify-otp", "", map[string]string{"email": email, "code": f.mail.code(email)}, "")
	expect(t, 200, status, out)
	return out["token"].(string)
}
func (f *fixture) claim(t *testing.T, token, handle string) {
	t.Helper()
	status, out := f.request("POST", "/v1/handles/claim", token, map[string]string{"handle": handle}, "")
	expect(t, 200, status, out)
}
func TestIdentityOTPAndSessions(t *testing.T) {
	f := setup(t)
	email := "alice@example.test"
	token := f.login(t, email)
	status, out := f.request("POST", "/v1/auth/verify-otp", "", map[string]string{"email": email, "code": f.mail.code(email)}, "")
	expect(t, 401, status, out)
	status, out = f.request("GET", "/v1/me", token, nil, "")
	expect(t, 200, status, out)
	if out["agent_id"] == nil || out["model_available"] != false {
		t.Fatal("missing private agent or false model availability")
	}
	status, out = f.request("POST", "/v1/auth/request-otp", "", map[string]string{"email": "expired@example.test"}, "")
	expect(t, 200, status, out)
	if _, err := f.db.Exec(`UPDATE email_otps SET expires_at=now()-interval '1 second' WHERE email=$1`, "expired@example.test"); err != nil {
		t.Fatal(err)
	}
	status, out = f.request("POST", "/v1/auth/verify-otp", "", map[string]string{"email": "expired@example.test", "code": f.mail.code("expired@example.test")}, "")
	expect(t, 401, status, out)
	status, out = f.request("POST", "/v1/auth/request-otp", "", map[string]string{"email": "brute@example.test"}, "")
	expect(t, 200, status, out)
	for range 5 {
		status, out = f.request("POST", "/v1/auth/verify-otp", "", map[string]string{"email": "brute@example.test", "code": "wrong"}, "")
		expect(t, 401, status, out)
	}
	status, out = f.request("POST", "/v1/auth/verify-otp", "", map[string]string{"email": "brute@example.test", "code": f.mail.code("brute@example.test")}, "")
	expect(t, 401, status, out)
	var stored string
	if err := f.db.QueryRow(`SELECT code_hash FROM email_otps WHERE email=$1`, email).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == f.mail.code(email) || len(stored) != 64 {
		t.Fatal("OTP not hashed")
	}
	for range 3 {
		status, out = f.request("POST", "/v1/auth/request-otp", "", map[string]string{"email": email}, "")
	}
	expect(t, 429, status, out)
	f.claim(t, token, "Alice")
	status, out = f.request("POST", "/v1/handles/claim", token, map[string]string{"handle": "another"}, "")
	expect(t, 409, status, out)
	status, out = f.request("POST", "/v1/auth/logout", token, map[string]any{}, "")
	expect(t, 200, status, out)
	status, out = f.request("GET", "/v1/me", token, nil, "")
	expect(t, 401, status, out)
}
func TestConcurrentHandleClaimAndPublicProfile(t *testing.T) {
	f := setup(t)
	a := f.login(t, "a@example.test")
	b := f.login(t, "b@example.test")
	var wg sync.WaitGroup
	statuses := make(chan int, 2)
	for _, token := range []string{a, b} {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			status, _ := f.request("POST", "/v1/handles/claim", token, map[string]string{"handle": "Mixed_Case"}, "")
			statuses <- status
		}(token)
	}
	wg.Wait()
	close(statuses)
	winners, losers := 0, 0
	for status := range statuses {
		if status == 200 {
			winners++
		} else if status == 409 {
			losers++
		} else {
			t.Fatalf("unexpected claim status %d", status)
		}
	}
	if winners != 1 || losers != 1 {
		t.Fatal("claim not atomic")
	}
	status, out := f.request("GET", "/v1/users/mixed_case", "", nil, "")
	expect(t, 200, status, out)
	if out["email"] != nil || out["agent_id"] != nil || out["account_id"] != nil {
		t.Fatal("public profile exposes private fields")
	}
	r := httptest.NewRequest("GET", "/u/mixed_case", nil)
	w := httptest.NewRecorder()
	f.s.ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "og:title") {
		t.Fatal("web profile missing")
	}
	for _, h := range []string{"root", "adm_in", "Αlice", "ab", "a-b"} {
		status, out = f.request("GET", "/v1/handles/"+h+"/availability", a, nil, "")
		expect(t, 400, status, out)
	}
}
func TestSocialNotesSettingsAndIdempotency(t *testing.T) {
	f := setup(t)
	a := f.login(t, "a@example.test")
	b := f.login(t, "b@example.test")
	f.claim(t, a, "alice")
	f.claim(t, b, "bob")
	status, out := f.request("GET", "/v1/rooms/lobby/messages", a, nil, "")
	expect(t, 403, status, out)
	for _, token := range []string{a, b} {
		status, out = f.request("POST", "/v1/rooms/lobby/join", token, map[string]any{}, "")
		expect(t, 200, status, out)
	}
	key := random(16)
	body := map[string]string{"content": "hello"}
	status, out = f.request("POST", "/v1/rooms/lobby/messages", a, body, key)
	expect(t, 200, status, out)
	id := out["id"]
	status, out = f.request("POST", "/v1/rooms/lobby/messages", a, body, key)
	expect(t, 200, status, out)
	if out["id"] != id {
		t.Fatal("idempotency replay duplicated write")
	}
	status, out = f.request("POST", "/v1/rooms/lobby/messages", a, map[string]string{"content": "different"}, key)
	expect(t, 409, status, out)
	status, out = f.request("GET", "/v1/rooms/lobby/messages?limit=1", b, nil, "")
	expect(t, 200, status, out)
	if len(out["items"].([]any)) != 1 {
		t.Fatal("message missing")
	}
	status, out = f.request("POST", "/v1/reports", b, map[string]any{"message_id": id, "reason": "review this"}, "")
	expect(t, 200, status, out)
	status, out = f.request("POST", "/v1/blocks", b, map[string]string{"handle": "alice"}, "")
	expect(t, 200, status, out)
	status, out = f.request("GET", "/v1/rooms/lobby/messages", b, nil, "")
	expect(t, 200, status, out)
	if len(out["items"].([]any)) != 0 {
		t.Fatal("blocked messages visible")
	}
	status, out = f.request("GET", "/v1/users?q=alice", b, nil, "")
	expect(t, 200, status, out)
	if len(out["items"].([]any)) != 0 {
		t.Fatal("blocked user in directory")
	}
	status, out = f.request("POST", "/v1/notes", a, map[string]string{"title": "private", "body": "secret note"}, "")
	expect(t, 200, status, out)
	note := out["id"].(string)
	for _, method := range []string{"GET", "PATCH", "DELETE"} {
		status, out = f.request(method, "/v1/notes/"+note, b, map[string]string{"title": "bad", "body": "bad"}, "")
		expect(t, 404, status, out)
	}
	status, out = f.request("PATCH", "/v1/notes/"+note, a, map[string]string{"title": "edited", "body": "persistent"}, "")
	expect(t, 200, status, out)
	status, out = f.request("GET", "/v1/notes/"+note, a, nil, "")
	expect(t, 200, status, out)
	if out["body"] != "persistent" {
		t.Fatal("note edit not persisted")
	}
	status, out = f.request("PATCH", "/v1/me", a, map[string]any{"theme": "dark", "public": false, "notifications": true, "bio": "<script>test</script>"}, "")
	expect(t, 200, status, out)
	status, out = f.request("GET", "/v1/users/alice", "", nil, "")
	expect(t, 404, status, out)
	status, out = f.request("GET", "/v1/me/export", a, nil, "")
	expect(t, 200, status, out)
	if out["notes"] == nil {
		t.Fatal("export missing notes")
	}
	status, out = f.request("POST", "/v1/notes", a, map[string]string{"title": "second", "body": "export page"}, "")
	expect(t, 200, status, out)
	status, out = f.request("GET", "/v1/me/export?limit=1", a, nil, "")
	expect(t, 200, status, out)
	if out["pagination"].(map[string]any)["next_offset"] != float64(1) {
		t.Fatal("export page cursor missing")
	}
	first := out["notes"].(map[string]any)["items"].([]any)[0].(map[string]any)["id"]
	status, out = f.request("GET", "/v1/me/export?limit=1&offset=1", a, nil, "")
	expect(t, 200, status, out)
	if first == out["notes"].(map[string]any)["items"].([]any)[0].(map[string]any)["id"] {
		t.Fatal("export page repeated")
	}

	status, out = f.request("POST", "/v1/me/delete", a, map[string]string{"confirm": "delete my account"}, "")
	expect(t, 200, status, out)
	status, out = f.request("GET", "/v1/me", a, nil, "")
	expect(t, 401, status, out)
	status, out = f.request("GET", "/v1/handles/alice/availability", b, nil, "")
	expect(t, 200, status, out)
	if out["available"] != false {
		t.Fatal("deleted handle reassigned")
	}
}
func TestRequestBoundaries(t *testing.T) {
	f := setup(t)
	status, out := f.request("GET", "/v1/notes", "", nil, "")
	expect(t, 401, status, out)
	status, out = f.request("POST", "/v1/auth/request-otp", "", map[string]string{"email": "bad\r\nBcc: x"}, "")
	expect(t, 400, status, out)
	token := f.login(t, "a@example.test")
	for _, path := range []string{"/v1/notes?limit=999", "/v1/notes?offset=-1", "/v1/notes/not-a-uuid"} {
		status, out = f.request("GET", path, token, nil, "")
		expect(t, 400, status, out)
	}
	r := httptest.NewRequest(http.MethodPost, "/v1/notes", strings.NewReader(fmt.Sprintf(`{"body":"%s"}`, strings.Repeat("x", 140000))))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	f.s.ServeHTTP(w, r)
	if w.Code != 413 {
		t.Fatalf("oversized body status %d", w.Code)
	}
}

func TestConcurrentOTPRedemptionAndSessionExpiry(t *testing.T) {
	f := setup(t)
	email := "race@example.test"
	status, out := f.request("POST", "/v1/auth/request-otp", "", map[string]string{"email": email}, "")
	expect(t, 200, status, out)
	var wg sync.WaitGroup
	statuses := make(chan int, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, _ := f.request("POST", "/v1/auth/verify-otp", "", map[string]string{"email": email, "code": f.mail.code(email)}, "")
			statuses <- status
		}()
	}
	wg.Wait()
	close(statuses)
	successful, denied := 0, 0
	for status := range statuses {
		if status == 200 {
			successful++
		} else if status == 401 {
			denied++
		} else {
			t.Fatalf("concurrent verification status %d", status)
		}
	}
	if successful != 1 || denied != 1 {
		t.Fatal("OTP was not one use under concurrency")
	}
	token := f.login(t, "expiry@example.test")
	if _, err := f.db.Exec(`UPDATE sessions SET expires_at=now()-interval '1 second' WHERE token_hash=$1`, hash(token)); err != nil {
		t.Fatal(err)
	}
	status, out = f.request("GET", "/v1/me", token, nil, "")
	expect(t, 401, status, out)
}
