package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
	"sinet"
)

// Menus render inline so command results and disclosure prompts remain in scrollback.
// Bubble Tea owns input decoding, redraws, resize events and terminal restoration.
func (a *App) menu(title string, items []string) (int, error) {
	if len(items) == 0 || len(items) > 9 {
		return 0, errors.New("menu requires 1–9 choices")
	}
	if !a.TTY || os.Getenv("TERM") == "dumb" {
		if _, err := fmt.Fprintf(a.Out, "\nSINet %s · %s\n", sinet.Version(), oneLine(title)); err != nil {
			return 0, err
		}
		for i, item := range items {
			if _, err := fmt.Fprintf(a.Out, "%d. %s\n", i+1, oneLine(item)); err != nil {
				return 0, err
			}
		}
		for {
			line, err := a.line("Choose a number, or q to go back: ")
			if err != nil {
				return 0, err
			}
			if line == "q" || line == "" {
				return len(items) - 1, nil
			}
			for i := range items {
				if line == fmt.Sprint(i+1) {
					return i, nil
				}
			}
			if _, err := fmt.Fprintln(a.Out, "Choose one of the listed numbers."); err != nil {
				return 0, err
			}
		}
	}
	prefs, err := loadAppearance()
	if err != nil {
		return 0, err
	}
	width, height, err := term.GetSize(int(a.Input.Fd()))
	if err != nil || width < 1 || height < 1 {
		width, height = 80, 24
	}
	m := menuModel{title: title, items: items, width: width, height: height, prefs: prefs, choice: len(items) - 1}
	opts := []tea.ProgramOption{tea.WithInput(a.Input), tea.WithOutput(a.Out), tea.WithWindowSize(width, height), tea.WithFilter(func(_ tea.Model, msg tea.Msg) tea.Msg {
		if size, ok := msg.(tea.WindowSizeMsg); ok && (size.Width <= 0 || size.Height <= 0) {
			return tea.WindowSizeMsg{Width: width, Height: height}
		}
		return msg
	})}
	_, noColor := os.LookupEnv("NO_COLOR")
	if noColor || prefs.Color == "never" {
		opts = append(opts, tea.WithColorProfile(colorprofile.NoTTY))
	} else if prefs.Color == "always" {
		opts = append(opts, tea.WithColorProfile(colorprofile.TrueColor))
	}
	result, err := tea.NewProgram(m, opts...).Run()
	if err != nil {
		return 0, fmt.Errorf("terminal menu: %w", err)
	}
	return result.(menuModel).choice, nil
}

type menuModel struct {
	title                           string
	items                           []string
	selected, choice, width, height int
	prefs                           appearance
	help                            bool
}

func (m menuModel) Init() tea.Cmd { return nil }
func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(1, msg.Width), max(1, msg.Height)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c", "ctrl+d":
			m.choice = len(m.items) - 1
			return m, tea.Quit
		case "enter":
			m.choice = m.selected
			return m, tea.Quit
		case "up", "k", "shift+tab":
			m.selected = (m.selected + len(m.items) - 1) % len(m.items)
		case "down", "j", "tab":
			m.selected = (m.selected + 1) % len(m.items)
		case "home":
			m.selected = 0
		case "end":
			m.selected = len(m.items) - 1
		case "?":
			m.help = !m.help
		default:
			key := msg.String()
			if len(key) == 1 && key[0] >= '1' && int(key[0]-'1') < len(m.items) {
				m.choice = int(key[0] - '1')
				return m, tea.Quit
			}
		}
		// PasteMsg is deliberately ignored: pasting text must never select an action.
	}
	return m, nil
}
func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
}
func (m menuModel) View() tea.View {
	width := max(1, min(76, m.width-1)) // Leave the last column free to prevent terminal auto-wrap.
	accent, muted, fg, bg := "#00D1FF", "#B8C6D8", "#F8FAFC", "#101928"
	if m.prefs.Theme == "light" {
		accent, muted, fg, bg = "#1554A0", "#475569", "#0F172A", "#E2E8F0"
	}
	style := func(s string, color string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(s)
	}
	fit := func(s string) string { return ansi.Truncate(s, width, "…") }
	header := fit(style("SINet", accent) + style("  /  "+sinet.Version(), muted))
	title := fit(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(fg)).Render(oneLine(m.title)))
	footer := "↑↓ / j k  move · Enter / 1–9  select · ? help · q back"
	if width < 58 {
		footer = "↑↓ move · Enter select · q back"
	}
	if m.help {
		footer = "Numbers select directly. Esc goes back. SINet help: commands."
	}
	// Keep selection visible in small terminals; compact mode removes decoration.
	decoration := m.prefs.Density != "compact" && m.height >= len(m.items)+7 && width >= 40
	overhead := 3
	if decoration {
		overhead = 6
	}
	rows := max(1, min(len(m.items), m.height-overhead-1))
	start := max(0, min(m.selected-rows+1, len(m.items)-rows))
	lines := []string{header, title}
	if decoration {
		lines = append(lines, style(strings.Repeat("─", width), muted))
	}
	for i := start; i < start+rows; i++ {
		text := fmt.Sprintf("  %d  %s", i+1, oneLine(m.items[i]))
		if i == m.selected {
			text = fmt.Sprintf("› %d  %s", i+1, oneLine(m.items[i]))
		}
		text = fit(text)
		if i == m.selected {
			text = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accent)).Background(lipgloss.Color(bg)).Width(width).Render(text)
		} else {
			text = style(text, fg)
		}
		lines = append(lines, text)
	}
	if decoration {
		lines = append(lines, "", style(strings.Repeat("─", width), muted))
	}
	if rows < len(m.items) {
		footer = fmt.Sprintf("%d/%d · ↑↓ move · Enter · q back", m.selected+1, len(m.items))
	}
	lines = append(lines, fit(style(footer, muted)))
	// Even unusually short terminals get a bounded, usable selection row.
	if m.height < 5 {
		lines = lines[2 : 2+rows]
	}
	return tea.NewView(strings.Join(lines, "\n"))
}
