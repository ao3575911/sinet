package api

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/sha3"
	"sinet/internal/wallet"
)

func testWallet(t *testing.T) (*secp256k1.PrivateKey, string) {
	t.Helper()
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write(key.PubKey().SerializeUncompressed()[1:])
	return key, "0x" + hex.EncodeToString(hash.Sum(nil)[12:])
}

func signWallet(key *secp256k1.PrivateKey, message string) string {
	compact := ecdsa.SignCompact(key, wallet.MessageHash(message), false)
	raw := append(append([]byte{}, compact[1:]...), compact[0])
	return "0x" + hex.EncodeToString(raw)
}

func challengeWallet(t *testing.T, f *fixture, token, address string) map[string]any {
	t.Helper()
	status, out := f.request("POST", "/v1/wallets/challenge", token, map[string]string{"address": address}, "")
	expect(t, 200, status, out)
	return out
}

func verifyWalletProof(f *fixture, token string, challenge map[string]any, signature, key string) (int, map[string]any) {
	return f.request("POST", "/v1/wallets/verify", token, map[string]any{"nonce": challenge["nonce"], "signature": signature}, key)
}

func TestWalletOwnershipIsolationReplayAndUnlink(t *testing.T) {
	f := setup(t)
	a := f.login(t, "wallet-a@example.test")
	b := f.login(t, "wallet-b@example.test")
	f.claim(t, a, "walletalice")
	f.claim(t, b, "walletbob")
	key, address := testWallet(t)
	challenge := challengeWallet(t, f, a, address)
	signature := signWallet(key, challenge["message"].(string))
	status, out := verifyWalletProof(f, b, challenge, signature, "")
	expect(t, 403, status, out)
	idempotency := random(16)
	status, out = verifyWalletProof(f, a, challenge, signature, idempotency)
	expect(t, 200, status, out)
	if out["status"] != "ownership_verified" || out["token"] != nil {
		t.Fatal("proof changed authentication authority")
	}
	status, out = verifyWalletProof(f, a, challenge, signature, idempotency)
	expect(t, 200, status, out)
	status, out = verifyWalletProof(f, a, challenge, signature, "")
	expect(t, 403, status, out)
	status, out = f.request("GET", "/v1/wallets", a, nil, "")
	expect(t, 200, status, out)
	if len(out["items"].([]any)) != 1 || out["transactions_available"] != false || out["balances_available"] != false {
		t.Fatal("wallet state or capability claim incorrect")
	}
	status, out = f.request("GET", "/v1/wallets", b, nil, "")
	expect(t, 200, status, out)
	if len(out["items"].([]any)) != 0 {
		t.Fatal("wallet leaked across accounts")
	}
	other := challengeWallet(t, f, b, address)
	status, out = verifyWalletProof(f, b, other, signWallet(key, other["message"].(string)), "")
	expect(t, 409, status, out)
	status, out = f.request("DELETE", "/v1/wallets/"+address, b, map[string]bool{"confirm": true}, "")
	expect(t, 404, status, out)
	status, out = f.request("DELETE", "/v1/wallets/"+address, a, map[string]bool{"confirm": false}, "")
	expect(t, 400, status, out)
	status, out = f.request("DELETE", "/v1/wallets/"+address, a, map[string]bool{"confirm": true}, "")
	expect(t, 200, status, out)
	status, out = verifyWalletProof(f, a, challenge, signature, "")
	expect(t, 403, status, out)
	var count int
	if err := f.db.QueryRow(`SELECT count(*) FROM wallet_links`).Scan(&count); err != nil || count != 0 {
		t.Fatal("wallet link remained after unlink", err)
	}
}

func TestWalletProofBoundsAndOrigin(t *testing.T) {
	f := setup(t)
	a := f.login(t, "wallet-bounds@example.test")
	f.claim(t, a, "walletbounds")
	key, address := testWallet(t)
	challenge := challengeWallet(t, f, a, address)
	wrong := strings.Replace(challenge["message"].(string), "127.0.0.1:8080", "phishing.example", -1)
	status, out := verifyWalletProof(f, a, challenge, signWallet(key, wrong), "")
	expect(t, 403, status, out)
	f.s.PublicURL = "https://other.example"
	status, out = verifyWalletProof(f, a, challenge, signWallet(key, challenge["message"].(string)), "")
	expect(t, 403, status, out)
	f.s.PublicURL = "http://127.0.0.1:8080"
	if _, err := f.db.Exec(`UPDATE wallet_challenges SET expires_at=now()-interval '1 second'`); err != nil {
		t.Fatal(err)
	}
	status, out = verifyWalletProof(f, a, challenge, signWallet(key, challenge["message"].(string)), "")
	expect(t, 403, status, out)
	old := challengeWallet(t, f, a, address)
	current := challengeWallet(t, f, a, address)
	status, out = verifyWalletProof(f, a, old, signWallet(key, old["message"].(string)), "")
	expect(t, 403, status, out)
	for range 5 {
		status, out = verifyWalletProof(f, a, current, "invalid", "")
		expect(t, 403, status, out)
	}
	status, out = verifyWalletProof(f, a, current, signWallet(key, current["message"].(string)), "")
	expect(t, 403, status, out)
	var attempts int
	var consumed bool
	if err := f.db.QueryRow(`SELECT attempts,consumed FROM wallet_challenges`).Scan(&attempts, &consumed); err != nil || attempts != 5 || !consumed {
		t.Fatal("proof attempt budget failed", err)
	}
}

func TestWalletConcurrentProofAndChallengeQuota(t *testing.T) {
	f := setup(t)
	a := f.login(t, "wallet-race@example.test")
	f.claim(t, a, "walletrace")
	key, address := testWallet(t)
	challenge := challengeWallet(t, f, a, address)
	signature := signWallet(key, challenge["message"].(string))
	var wg sync.WaitGroup
	statuses := make(chan int, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, _ := verifyWalletProof(f, a, challenge, signature, "")
			statuses <- status
		}()
	}
	wg.Wait()
	close(statuses)
	counts := map[int]int{}
	for status := range statuses {
		counts[status]++
	}
	if counts[200] != 1 || counts[403] != 1 {
		t.Fatalf("concurrent proof replay accepted: %v", counts)
	}
	for range 4 {
		challengeWallet(t, f, a, address)
	}
	status, out := f.request("POST", "/v1/wallets/challenge", a, map[string]string{"address": address}, "")
	expect(t, 429, status, out)
	status, out = f.request("POST", "/v1/wallets/challenge", a, map[string]any{"address": address, "chain_id": 10}, "")
	expect(t, 400, status, out)
}

func TestWalletLinkLimitAndActivation(t *testing.T) {
	f := setup(t)
	a := f.login(t, "wallet-limit@example.test")
	_, address := testWallet(t)
	status, out := f.request("POST", "/v1/wallets/challenge", a, map[string]string{"address": address}, "")
	expect(t, 403, status, out)
	f.claim(t, a, "walletlimit")
	var account string
	if err := f.db.QueryRow(`SELECT id FROM accounts WHERE email='wallet-limit@example.test'`).Scan(&account); err != nil {
		t.Fatal(err)
	}
	for index := 1; index <= 5; index++ {
		if _, err := f.db.Exec(`INSERT INTO wallet_links(account_id,chain_id,address,proof_hash) VALUES($1,1,$2,$3)`, account, fmt.Sprintf("0x%040x", index), strings.Repeat("a", 64)); err != nil {
			t.Fatal(err)
		}
	}
	status, out = f.request("POST", "/v1/wallets/challenge", a, map[string]string{"address": address}, "")
	expect(t, 429, status, out)
}

func TestEmailDraftPrivacyValidationAndExportDeletion(t *testing.T) {
	f := setup(t)
	a := f.login(t, "email-a@example.test")
	b := f.login(t, "email-b@example.test")
	f.claim(t, a, "emailalice")
	f.claim(t, b, "emailbob")
	for _, path := range []string{"/v1/wallets", "/v1/email", "/v1/email/drafts", "/v1/sections"} {
		status, out := f.request("GET", path, "", nil, "")
		expect(t, 401, status, out)
	}
	status, out := f.request("GET", "/v1/email", a, nil, "")
	expect(t, 200, status, out)
	if out["sign_in_email"] != "email-a@example.test" || out["mailbox_sync_status"] != "not_connected" || out["alias"].(map[string]any)["address"] != nil {
		t.Fatal("fictional mailbox provisioning")
	}
	draft := map[string]string{"recipient": "Friend@example.test", "subject": "Private draft", "body": "First line\r\nSecond line"}
	idempotency := random(16)
	status, out = f.request("POST", "/v1/email/drafts", a, draft, idempotency)
	expect(t, 200, status, out)
	id := out["id"].(string)
	if out["status"] != "draft" || out["body"] != "First line\nSecond line" || out["recipient"] != "Friend@example.test" {
		t.Fatal("draft state/normalization incorrect")
	}
	status, out = f.request("POST", "/v1/email/drafts", a, draft, idempotency)
	expect(t, 200, status, out)
	if out["id"] != id {
		t.Fatal("draft idempotency duplicated creation")
	}
	for _, method := range []string{"GET", "PATCH", "DELETE"} {
		status, out = f.request(method, "/v1/email/drafts/"+id, b, draft, "")
		expect(t, 404, status, out)
	}
	status, out = f.request("GET", "/v1/email/drafts", b, nil, "")
	expect(t, 200, status, out)
	if len(out["items"].([]any)) != 0 {
		t.Fatal("draft listed for another account")
	}
	for _, invalid := range []map[string]string{
		{"recipient": "one@example.test\r\nBcc: two@example.test", "subject": "x", "body": "x"},
		{"recipient": "one@example.test", "subject": "bad\nheader", "body": "x"},
		{"recipient": "one@example.test", "subject": "x", "body": "hidden\u202etext"},
		{"recipient": "one@example.test", "subject": strings.Repeat("x", 201), "body": "x"},
		{"recipient": "one@example.test", "subject": "x", "body": strings.Repeat("x", 65537)},
	} {
		status, out = f.request("POST", "/v1/email/drafts", a, invalid, "")
		expect(t, 400, status, out)
	}
	status, out = f.request("PATCH", "/v1/email/drafts/"+id, a, map[string]string{"recipient": "", "subject": "Edited", "body": "Still a draft"}, "")
	expect(t, 200, status, out)
	key, address := testWallet(t)
	challenge := challengeWallet(t, f, a, address)
	status, out = verifyWalletProof(f, a, challenge, signWallet(key, challenge["message"].(string)), "")
	expect(t, 200, status, out)
	status, out = f.request("GET", "/v1/users/emailalice", b, nil, "")
	expect(t, 200, status, out)
	if out["wallets"] != nil || out["email_drafts"] != nil || out["email"] != nil {
		t.Fatal("private sections leaked into public profile")
	}
	status, out = f.request("GET", "/v1/me/export", a, nil, "")
	expect(t, 200, status, out)
	for _, field := range []string{"wallets", "email_drafts"} {
		if len(out[field].(map[string]any)["items"].([]any)) != 1 {
			t.Fatal("new private data missing from export", field)
		}
	}
	var account string
	if err := f.db.QueryRow(`SELECT id FROM accounts WHERE email='email-a@example.test'`).Scan(&account); err != nil {
		t.Fatal(err)
	}
	status, out = f.request("POST", "/v1/me/delete", a, map[string]string{"confirm": "delete my account"}, "")
	expect(t, 200, status, out)
	for _, table := range []string{"wallet_links", "wallet_challenges", "email_drafts"} {
		var count int
		if err := f.db.QueryRow(`SELECT count(*) FROM `+table+` WHERE account_id=$1`, account).Scan(&count); err != nil || count != 0 {
			t.Fatal("private data retained after account deletion", table, err)
		}
	}
}

func TestEmailDraftPaginationLimitAndSectionReadiness(t *testing.T) {
	f := setup(t)
	a := f.login(t, "email-limit@example.test")
	var account string
	if err := f.db.QueryRow(`SELECT id FROM accounts WHERE email='email-limit@example.test'`).Scan(&account); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`INSERT INTO email_drafts(account_id,subject,body) SELECT $1,'draft '||n,'private' FROM generate_series(1,100) n`, account); err != nil {
		t.Fatal(err)
	}
	status, out := f.request("POST", "/v1/email/drafts", a, map[string]string{"subject": "extra", "body": "extra"}, "")
	expect(t, 429, status, out)
	status, out = f.request("GET", "/v1/email/drafts?limit=1&offset=1", a, nil, "")
	expect(t, 200, status, out)
	if len(out["items"].([]any)) != 1 {
		t.Fatal("draft pagination ignored")
	}
	id := out["items"].([]any)[0].(map[string]any)["id"].(string)
	status, out = f.request("DELETE", "/v1/email/drafts/"+id, a, nil, "")
	expect(t, 200, status, out)
	status, out = f.request("GET", "/v1/sections", a, nil, "")
	expect(t, 200, status, out)
	if len(out["items"].([]any)) != 7 {
		t.Fatal("service sections missing")
	}
	for _, id := range []string{"business", "identity", "storage", "compute", "mobile"} {
		status, out = f.request("GET", "/v1/sections/"+id, a, nil, "")
		expect(t, 200, status, out)
		if out["status"] != "planned" {
			t.Fatal("unimplemented section claims availability")
		}
	}
}
