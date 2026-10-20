package api

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"sinet/internal/operator"
	"sinet/migrations"
)

func TestOperatorReviewModerationRetentionAndMigrations(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	var out bytes.Buffer
	if err := migrations.Apply(ctx, f.db); err != nil {
		t.Fatal("migrations are not repeatable", err)
	}
	a := f.login(t, "creator@example.test")
	f.claim(t, a, "creator")
	raw, err := os.ReadFile("../../connectors/examples/echo.json")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	m["id"] = "review-fixture"
	m["publisher"] = "creator"
	m["runtime"] = "remote_https"
	m["endpoint_url"] = "https://example.com/tools"
	m["privacy_policy_url"] = "https://example.com/privacy"
	status, body := f.request("POST", "/v1/connectors/submissions", a, m, "")
	expect(t, 200, status, body)
	if err = operator.Run(ctx, f.db, []string{"review", "approve", "review-fixture", "1.0.0", "missing proofs", "--confirm"}, &out); err == nil {
		t.Fatal("approval without ownership/review attestations accepted")
	}
	if err = operator.Run(ctx, f.db, []string{"review", "approve", "review-fixture", "1.0.0", "ownership: TEST FIXTURE ONLY; review: test docket", "--confirm"}, &out); err != nil {
		t.Fatal(err)
	}
	if err = operator.Run(ctx, f.db, []string{"review", "reject", "review-fixture", "1.0.0", "test revocation reference", "--confirm"}, &out); err != nil {
		t.Fatal(err)
	}
	status, body = f.request("POST", "/v1/connectors/review-fixture/install", a, map[string]string{"version": "1.0.0"}, "")
	expect(t, 403, status, body)
	status, body = f.request("POST", "/v1/rooms/lobby/join", a, map[string]any{}, "")
	expect(t, 200, status, body)
	status, body = f.request("POST", "/v1/rooms/lobby/messages", a, map[string]string{"content": "remove this fixture"}, "")
	expect(t, 200, status, body)
	id := body["id"]
	status, body = f.request("POST", "/v1/reports", a, map[string]any{"message_id": id, "reason": "test moderation"}, "")
	expect(t, 200, status, body)
	if err = operator.Run(ctx, f.db, []string{"reports", "resolve", body["id"].(string), "remove-message", "--confirm"}, &out); err != nil {
		t.Fatal(err)
	}
	status, body = f.request("GET", "/v1/rooms/lobby/messages", a, nil, "")
	expect(t, 200, status, body)
	if len(body["items"].([]any)) != 0 {
		t.Fatal("moderation failed")
	}
	if err = operator.Run(ctx, f.db, []string{"maintenance", "--confirm"}, &out); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec(`UPDATE schema_migrations SET checksum='tampered' WHERE name='0001_core.sql'`); err != nil {
		t.Fatal(err)
	}
	if err = migrations.Apply(ctx, f.db); err == nil {
		t.Fatal("migration checksum change accepted")
	}
}
