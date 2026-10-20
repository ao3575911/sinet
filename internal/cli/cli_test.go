package cli

import (
	"strings"
	"testing"
)

func TestTerminalOutputAndPagination(t *testing.T) {
	clean := SafeText("hello\x1b[2J\x00\u202euser\ntext")
	if strings.ContainsAny(clean, "\x1b\x00\u202e") || !strings.Contains(clean, "\n") {
		t.Fatal("terminal controls not removed")
	}
	args, q, err := pageFlags([]string{"lobby", "read", "--limit", "5", "--before", "42"})
	if err != nil || len(args) != 2 || !strings.Contains(q, "before=42") {
		t.Fatal("pagination flags invalid")
	}
	if _, _, err = pageFlags([]string{"--offset", "-1"}); err == nil {
		t.Fatal("negative pagination accepted")
	}
}
