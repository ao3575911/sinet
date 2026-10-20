package wallet

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"
)

type vector struct{ Address, Message, Signature, Digest string }

func loadVector(t *testing.T) vector {
	t.Helper()
	data, err := os.ReadFile("testdata/personal_sign.json")
	if err != nil {
		t.Fatal(err)
	}
	var result vector
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestIndependentPersonalSignVector(t *testing.T) {
	v := loadVector(t)
	if "0x"+hex.EncodeToString(MessageHash(v.Message)) != v.Digest {
		t.Fatal("EIP-191 digest differs from ethers vector")
	}
	if err := Verify(v.Address, v.Message, v.Signature); err != nil {
		t.Fatal(err)
	}
	raw, _ := hex.DecodeString(v.Signature[2:])
	raw[64] -= 27
	if err := Verify(v.Address, v.Message, "0x"+hex.EncodeToString(raw)); err != nil {
		t.Fatal("v=0/1 signature rejected", err)
	}
	for _, test := range []struct{ address, message, signature string }{
		{v.Address, v.Message + "changed", v.Signature},
		{"0xde709f2102306220921060314715629080e2fb77", v.Message, v.Signature},
		{v.Address, v.Message, "0x" + strings.Repeat("0", 130)},
		{v.Address, v.Message, v.Signature[:130] + "02"},
		{v.Address, v.Message, v.Signature[:130] + "ff"},
		{v.Address, v.Message, "0x" + strings.Repeat("f", 130)},
		{v.Address, strings.Repeat("x", 2049), v.Signature},
	} {
		if Verify(test.address, test.message, test.signature) == nil {
			t.Fatal("invalid proof accepted")
		}
	}
}

func TestSignatureMalleabilityDenied(t *testing.T) {
	v := loadVector(t)
	raw, _ := hex.DecodeString(v.Signature[2:])
	order, _ := new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	s := new(big.Int).SetBytes(raw[32:64])
	order.Sub(order, s).FillBytes(raw[32:64])
	raw[64] = 27 + ((raw[64] - 27) ^ 1)
	if Verify(v.Address, v.Message, "0x"+hex.EncodeToString(raw)) == nil {
		t.Fatal("high-s signature accepted")
	}
}

func TestEIP55AndInvalidAddresses(t *testing.T) {
	for _, address := range []string{"0x52908400098527886E0F7030069857D2E4169EE7", "0x8617E340B3D01FA5F11F306F4090FD50E238070D", "0xde709f2102306220921060314715629080e2fb77", loadVector(t).Address} {
		lower, err := Address(address)
		if err != nil || Checksum(lower) != address {
			t.Fatalf("EIP-55 vector %s: %v", address, err)
		}
	}
	for _, address := range []string{"0x" + strings.Repeat("0", 40), "0x123", "not an address", "0x7e5F4552091A69125d5DfCb7b8C2659029395Bdf"} {
		if _, err := Address(address); err == nil {
			t.Fatalf("invalid address accepted %s", address)
		}
	}
}

func TestChallengeOriginAndTimeBounds(t *testing.T) {
	v := loadVector(t)
	now := time.Date(2026, 10, 20, 1, 2, 3, 0, time.UTC)
	nonce := strings.Repeat("a", 32)
	message, err := Message("https://sinet.example", v.Address, "walletalice", nonce, now, now.Add(5*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"sinet.example wants", v.Address, "URI: https://sinet.example/wallet/link", "Chain ID: 1", "Nonce: " + nonce, "Issued At: 2026-10-20T01:02:03Z", "No transactions are authorized"} {
		if !strings.Contains(message, expected) {
			t.Fatal("missing challenge binding", expected)
		}
	}
	for _, origin := range []string{"http://sinet.example", "https://user:pass@sinet.example", "https://sinet.example/path", "https://sinet.example?x=1", "https://sinet.example#x"} {
		if _, err := Message(origin, v.Address, "walletalice", nonce, now, now.Add(time.Minute)); err == nil {
			t.Fatal("unsafe origin accepted", origin)
		}
	}
	if _, err := Message("http://127.0.0.1:8080", v.Address, "walletalice", nonce, now, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	for _, end := range []time.Time{now, now.Add(6 * time.Minute)} {
		if _, err := Message("https://sinet.example", v.Address, "walletalice", nonce, now, end); err == nil {
			t.Fatal("invalid time bound accepted")
		}
	}
	if _, err := Message("https://sinet.example", v.Address, "walletalice", "bad\nnonce", now, now.Add(time.Minute)); err == nil {
		t.Fatal("invalid nonce accepted")
	}
}
