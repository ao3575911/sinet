//go:build !windows

package client

import (
	"errors"
	"os"
)

func protectSessionFile(f *os.File) error { return f.Chmod(0600) }
func checkSessionPermissions(f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0077 != 0 {
		return errors.New("session file permissions must be 0600")
	}
	return nil
}
