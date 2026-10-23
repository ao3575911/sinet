//go:build !darwin && !windows

package client

import "errors"

const nativeCredentials = false

func readCredential(string, Session) (string, error) {
	return "", errors.New("native credential store is unsupported on this platform")
}
func writeCredential(string, Session) error {
	return errors.New("native credential store is unsupported on this platform")
}
func deleteCredential(string, Session) error {
	return errors.New("native credential store is unsupported on this platform")
}
