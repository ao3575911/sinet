//go:build darwin || windows

package client

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const nativeCredentials = true
const credentialService = "SINet session v1"

// The file location and API origin are part of the key: copying or editing the
// metadata cannot send a credential to another origin or another profile.
func credentialKey(path string, s Session) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(path + "\x00" + s.API + "\x00" + s.Credential))
	return hex.EncodeToString(sum[:]), nil
}

func readCredential(path string, s Session) (string, error) {
	key, err := credentialKey(path, s)
	if err != nil {
		return "", err
	}
	token, err := keyring.Get(credentialService, key)
	if err != nil || !validToken(token) {
		return "", errors.New("OS credential unavailable; unlock the keychain or remove the session file and sign in again")
	}
	return token, nil
}

func writeCredential(path string, s Session) error {
	key, err := credentialKey(path, s)
	if err != nil {
		return err
	}
	if err = keyring.Set(credentialService, key, s.Token); err != nil {
		return errors.New("could not save OS credential; unlock the keychain and retry")
	}
	// Some native tools can report successful process exit after an operation
	// fails. Read back before replacing the session file; never fall back to text.
	got, err := readCredential(path, s)
	if err != nil || got != s.Token {
		_ = keyring.Delete(credentialService, key)
		return errors.New("could not verify saved OS credential")
	}
	return nil
}

func deleteCredential(path string, s Session) error {
	key, err := credentialKey(path, s)
	if err != nil {
		return err
	}
	err = keyring.Delete(credentialService, key)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return errors.New("could not remove OS credential; unlock the keychain and retry")
	}
	return nil
}
