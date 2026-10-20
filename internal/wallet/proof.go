// Package wallet verifies noncustodial Ethereum account-ownership proofs.
// It never generates, receives or stores a user's signing key or transactions.
package wallet

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/sha3"
)

func keccak(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write(data)
	return hash.Sum(nil)
}

// Address normalizes a nonzero Ethereum address, checking mixed-case EIP-55 input.
func Address(raw string) (string, error) {
	if len(raw) != 42 || !strings.HasPrefix(raw, "0x") {
		return "", errors.New("invalid Ethereum address")
	}
	value, err := hex.DecodeString(raw[2:])
	if err != nil || bytes.Equal(value, make([]byte, 20)) {
		return "", errors.New("invalid Ethereum address")
	}
	lower := strings.ToLower(raw)
	if raw[2:] != strings.ToLower(raw[2:]) && raw[2:] != strings.ToUpper(raw[2:]) && raw != Checksum(lower) {
		return "", errors.New("invalid EIP-55 checksum")
	}
	return lower, nil
}

// Checksum formats an already validated lowercase address using EIP-55.
func Checksum(address string) string {
	text := []byte(strings.ToLower(address[2:]))
	hash := hex.EncodeToString(keccak(text))
	for i, c := range text {
		if c >= 'a' && c <= 'f' && hash[i] >= '8' {
			text[i] -= 'a' - 'A'
		}
	}
	return "0x" + string(text)
}

// Message creates a SIWE-format message for wallet linking, not account login.
// The server binds the stored challenge to the signed-in account and one nonce.
func Message(origin, address, account, nonce string, issued, expires time.Time) (string, error) {
	u, err := url.Parse(origin)
	if err != nil || len(origin) > 512 || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") ||
		(u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"))) {
		return "", errors.New("wallet origin must be HTTPS or loopback HTTP")
	}
	normalized, err := Address(address)
	if err != nil {
		return "", err
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_]{2,23}$`).MatchString(account) {
		return "", errors.New("invalid SINet handle")
	}
	nonceBytes, err := hex.DecodeString(nonce)
	if err != nil || len(nonceBytes) != 16 || len(nonce) != 32 || !expires.After(issued) || expires.Sub(issued) > 5*time.Minute {
		return "", errors.New("invalid challenge bounds")
	}
	origin = strings.TrimRight(origin, "/")
	message := fmt.Sprintf("%s wants you to sign in with your Ethereum account:\n%s\n\nLink this wallet to SINet @%s. No transactions are authorized.\n\nURI: %s/wallet/link\nVersion: 1\nChain ID: 1\nNonce: %s\nIssued At: %s\nExpiration Time: %s\nResources:\n- urn:sinet:wallet-link", u.Host, Checksum(normalized), account, origin, nonce, issued.UTC().Format(time.RFC3339), expires.UTC().Format(time.RFC3339))
	if len(message) > 2048 {
		return "", errors.New("challenge too long")
	}
	return message, nil
}

// MessageHash is the EIP-191 personal_sign digest of the exact message bytes.
func MessageHash(message string) []byte {
	return keccak(append([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len([]byte(message)))), []byte(message)...))
}

// Verify accepts a canonical 65-byte personal_sign EOA signature, with v=0/1/27/28.
// Contract wallets (EIP-1271), chain access, balances and transactions are excluded.
func Verify(address, message, signature string) error {
	normalized, err := Address(address)
	if err != nil || len(message) > 2048 || len(signature) != 132 || !strings.HasPrefix(signature, "0x") {
		return errors.New("invalid wallet proof")
	}
	raw, err := hex.DecodeString(signature[2:])
	if err != nil {
		return errors.New("invalid wallet proof")
	}
	recovery := raw[64]
	if recovery >= 27 {
		recovery -= 27
	}
	if recovery > 1 {
		return errors.New("invalid recovery id")
	}
	var r, s secp256k1.ModNScalar
	if r.SetByteSlice(raw[:32]) || s.SetByteSlice(raw[32:64]) || r.IsZero() || s.IsZero() || s.IsOverHalfOrder() {
		return errors.New("noncanonical wallet signature")
	}
	compact := append([]byte{27 + recovery}, raw[:64]...)
	public, _, err := ecdsa.RecoverCompact(compact, MessageHash(message))
	if err != nil {
		return errors.New("invalid wallet proof")
	}
	actual := "0x" + hex.EncodeToString(keccak(public.SerializeUncompressed()[1:])[12:])
	if actual != normalized {
		return errors.New("wallet signer mismatch")
	}
	return nil
}
