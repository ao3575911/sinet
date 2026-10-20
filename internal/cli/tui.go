package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"
)

func (a *App) menu(title string, items []string) (int, error) {
	if !a.TTY {
		line, err := a.line(title + " (number, or q): ")
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
		return 0, errors.New("choose a listed number")
	}
	state, err := term.MakeRaw(int(a.Input.Fd()))
	if err != nil {
		return 0, err
	}
	defer term.Restore(int(a.Input.Fd()), state)
	selected := 0
	draw := func() {
		_, _ = fmt.Fprintf(a.Out, "\r\n%s\r\n", title)
		for i, item := range items {
			marker := " "
			if i == selected {
				marker = ">"
			}
			_, _ = fmt.Fprintf(a.Out, "%s %d. %s\r\n", marker, i+1, item)
		}
		_, _ = fmt.Fprint(a.Out, "↑/↓ or j/k · Enter selects · number selects · ? help · Esc/q back\r\n")
	}
	draw()
	buf := make([]byte, 8)
	for {
		n, err := a.Input.Read(buf)
		if err != nil {
			return 0, err
		}
		for i := 0; i < n; i++ {
			key := string(buf[i])
			if buf[i] == 27 && i+2 < n && buf[i+1] == '[' {
				key = string(buf[i : i+3])
				i += 2
			}
			switch key {
			case "q", "\x1b", "\x03":
				return len(items) - 1, nil
			case "\r", "\n":
				return selected, nil
			case "k", "\x1b[A":
				selected = (selected + len(items) - 1) % len(items)
			case "j", "\x1b[B":
				selected = (selected + 1) % len(items)
			case "?":
				_, _ = fmt.Fprint(a.Out, "\r\nChoose an action. Every action also has a plain command: project-name help.\r\n")
			default:
				if len(key) == 1 && key[0] >= '1' && int(key[0]-'1') < len(items) {
					return int(key[0] - '1'), nil
				}
				continue
			}
			draw()
		}
	}
}
func (a *App) dashboard() error {
	_, _ = fmt.Fprintln(a.Out, "SINet · Your terminal. Your network.")
	if !a.C.Authenticated() {
		choice, err := a.menu("Welcome", []string{"Sign in with email", "Help", "Exit"})
		if err != nil {
			return err
		}
		if choice == 2 {
			return nil
		}
		if choice == 1 {
			_, err = fmt.Fprint(a.Out, Help)
			return err
		}
		if err = a.login(""); err != nil {
			return err
		}
	}
	me, err := a.C.Do("GET", "/v1/me", nil)
	if err != nil {
		return err
	}
	if me["handle"] == nil {
		handle, err := a.line("Choose your handle (3–24 ASCII letters, numbers, underscores): ")
		if err != nil {
			return err
		}
		if err = a.run([]string{"claim", handle}); err != nil {
			return err
		}
	}
	for {
		choice, err := a.menu("Your network", []string{"Profile and share card", "Find people", "Community lobby", "Notes editor", "Connectors", "Personal agent", "Settings", "Help", "Exit"})
		if err != nil {
			return err
		}
		switch choice {
		case 0:
			err = a.run([]string{"share"})
		case 1:
			var q string
			q, err = a.line("Find people: ")
			if err == nil {
				err = a.run([]string{"find", q})
			}
		case 2:
			err = a.run([]string{"room", "lobby", "join"})
			if err == nil {
				err = a.run([]string{"room", "lobby", "read"})
			}
			if err == nil {
				var text string
				text, err = a.line("Message (empty goes back): ")
				if err == nil && text != "" {
					err = a.run([]string{"room", "lobby", "say", text})
				}
			}
		case 3:
			err = a.noteEditor()
		case 4:
			err = a.connectorMenu()
		case 5:
			err = a.run([]string{"agent", "history"})
			if err == nil {
				var msg string
				msg, err = a.line("Ask your agent (empty goes back): ")
				if err == nil && msg != "" {
					var approval string
					approval, err = a.line("If a model is configured, this sends your conversation to its provider. Type send to approve: ")
					if err == nil {
						args := []string{"agent", "ask", msg}
						if approval == "send" {
							args = append(args, "--allow-model")
						}
						err = a.run(args)
					}
				}
			}
		case 6:
			err = a.run([]string{"settings"})
			if err == nil {
				var key string
				key, err = a.line("Setting to edit (theme/public/notifications/display_name/bio, empty goes back): ")
				if err == nil && key != "" {
					var v string
					v, err = a.line("New value: ")
					if err == nil {
						err = a.run([]string{"settings", "set", key, v})
					}
				}
			}
		case 7:
			_, err = fmt.Fprint(a.Out, Help)
		case 8:
			return nil
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			_, _ = fmt.Fprintln(a.Out, SafeText(err.Error()))
		}
	}
}
func (a *App) noteEditor() error {
	if err := a.run([]string{"note", "list"}); err != nil {
		return err
	}
	id, err := a.line("Note ID to edit, or empty for new (q goes back): ")
	if err != nil {
		return err
	}
	if id == "q" {
		return nil
	}
	if id != "" {
		if err = a.run([]string{"note", "get", id}); err != nil {
			return err
		}
	}
	title, err := a.line("Title: ")
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(a.Out, "Enter replacement text. A single . saves; .cancel discards.")
	var body strings.Builder
	for {
		line, err := a.In.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(line) == ".cancel" {
			return nil
		}
		if strings.TrimSpace(line) == "." {
			break
		}
		body.WriteString(line)
		if body.Len() > 65536 {
			return errors.New("note exceeds 65536 bytes")
		}
	}
	if id == "" {
		return a.run([]string{"note", "new", title, body.String()})
	}
	return a.run([]string{"note", "edit", id, title, body.String()})
}
func (a *App) connectorMenu() error {
	if err := a.run([]string{"plug", "list"}); err != nil {
		return err
	}
	choice, err := a.menu("Connectors", []string{"Inspect connector", "Install connector", "Grant operation", "Run echo through agent", "Installed and permissions", "Revoke operation", "Remove connector", "Back"})
	if err != nil {
		return err
	}
	if choice == 7 {
		return nil
	}
	if choice == 3 {
		text, err := a.line("Text for echo: ")
		if err != nil {
			return err
		}
		raw := fmt.Sprintf("{\"text\":%q}", text)
		return a.run([]string{"agent", "call", "echo", "1.0.0", "repeat", raw})
	}
	if choice == 4 {
		return a.run([]string{"plug", "installed"})
	}
	id, err := a.line("Connector ID: ")
	if err != nil {
		return err
	}
	if choice == 6 {
		return a.run([]string{"plug", "remove", id})
	}
	if choice == 5 {
		op, err := a.line("Operation: ")
		if err != nil {
			return err
		}
		return a.run([]string{"plug", "revoke", id, op})
	}
	version, err := a.line("Version: ")
	if err != nil {
		return err
	}
	if err = a.run([]string{"plug", "info", id, version}); err != nil {
		return err
	}
	if choice == 0 {
		return nil
	}
	if choice == 1 {
		approve, err := a.line("Type install to pin this reviewed version (grants are separate): ")
		if err != nil {
			return err
		}
		if approve != "install" {
			return nil
		}
		return a.run([]string{"plug", "install", id, version})
	}
	op, err := a.line("Operation to grant: ")
	if err != nil {
		return err
	}
	approve, err := a.line("Type grant to allow that operation. Writes and exports still need per-call consent: ")
	if err != nil {
		return err
	}
	if approve != "grant" {
		return nil
	}
	return a.run([]string{"plug", "grant", id, version, op, "--approve"})
}
