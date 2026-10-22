package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestMenuNavigationAndBounds(t *testing.T) {
	items := []string{"Profile", "Find", "Community", "Notes", "Connectors", "Agent", "Settings", "Wallet, Email and Services", "Exit"}
	for _, size := range [][2]int{{80, 24}, {40, 14}, {20, 8}, {8, 3}} {
		m := menuModel{title: "Your network", items: items, width: size[0], height: size[1], prefs: defaultAppearance()}
		for i := 0; i < 27; i++ {
			result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
			m = result.(menuModel)
			view := m.View().Content
			if len(strings.Split(view, "\n")) > size[1] {
				t.Fatalf("height overflow %v: %q", size, view)
			}
			for _, line := range strings.Split(view, "\n") {
				if ansi.StringWidth(line) >= size[0] {
					t.Fatalf("width overflow %v: %q", size, line)
				}
			}
		}
		if m.selected != 0 {
			t.Fatal("navigation did not wrap")
		}
		before := m.View().Content
		result, cmd := m.Update(tea.PasteMsg{Content: "8\n"})
		if cmd != nil || result.(menuModel).View().Content != before {
			t.Fatal("paste selected action")
		}
		for i := 0; i < 9; i++ {
			result, cmd = m.Update(tea.KeyPressMsg{Code: rune('1' + i), Text: string(rune('1' + i))})
			if cmd == nil || result.(menuModel).choice != i {
				t.Fatalf("shortcut %d", i+1)
			}
		}
	}
}
func TestAppearancePersistenceAndRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.json")
	t.Setenv("SINET_UI_CONFIG", path)
	p, err := loadAppearance()
	if err != nil || p != defaultAppearance() {
		t.Fatal(p, err)
	}
	p.Theme = "light"
	p.Color = "never"
	p.Density = "compact"
	if err = saveAppearance(p); err != nil {
		t.Fatal(err)
	}
	got, err := loadAppearance()
	if err != nil || got != p {
		t.Fatal(got, err)
	}
	for _, bad := range []string{`{"theme":"unknown"}`, `{} {}`, strings.Repeat("x", 4097)} {
		if err = os.WriteFile(path, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = loadAppearance(); err == nil || !strings.Contains(err.Error(), "SINet ui reset") {
			t.Fatal("missing recovery guidance", err)
		}
	}
	if err = saveAppearance(defaultAppearance()); err != nil {
		t.Fatal(err)
	}
}
func TestMenuSanitizesTerminalContent(t *testing.T) {
	m := menuModel{title: "Title\x1b\n\u202e", items: []string{"Item\x1b\r\n"}, width: 80, height: 24, prefs: defaultAppearance()}
	plain := ansi.Strip(m.View().Content)
	if strings.ContainsAny(plain, "\r\u202e") || strings.Contains(plain, "Title\n\n") {
		t.Fatal(plain)
	}
}
