package client

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestWindowsCredentialACL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	t.Setenv("SINET_CONFIG", path)
	t.Setenv("SINET_API", "https://api.example.test")
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Forget() })
	if err := c.Save(strings.Repeat("ab", 32)); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	err = checkSessionPermissions(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	// An explicit Everyone-read ACE is rejected even for token-free metadata.
	sd, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, dacl, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := New(); err == nil {
		t.Fatal("broad Windows ACL accepted")
	}
}
