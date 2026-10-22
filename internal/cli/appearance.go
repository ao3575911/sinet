package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sinet/internal/strictjson"
)

type appearance struct {
	Theme   string `json:"theme"`
	Color   string `json:"color"`
	Density string `json:"density"`
}

func defaultAppearance() appearance { return appearance{"dark", "auto", "comfortable"} }
func appearancePath() (string, error) {
	if path := os.Getenv("SINET_UI_CONFIG"); path != "" {
		return path, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "SINet", "ui.json"), nil
}
func (p appearance) validate() error {
	if p.Theme != "dark" && p.Theme != "light" {
		return errors.New("theme must be dark or light")
	}
	if p.Color != "auto" && p.Color != "always" && p.Color != "never" {
		return errors.New("color must be auto, always or never")
	}
	if p.Density != "comfortable" && p.Density != "compact" {
		return errors.New("density must be comfortable or compact")
	}
	return nil
}
func loadAppearance() (appearance, error) {
	p := defaultAppearance()
	path, err := appearancePath()
	if err != nil {
		return p, err
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	if !info.Mode().IsRegular() {
		return p, errors.New("appearance settings must be a regular file; run SINet ui reset")
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
	if err != nil {
		return p, fmt.Errorf("open appearance settings: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 4097))
	if err == nil && len(data) > 4096 {
		err = errors.New("file exceeds 4096 bytes")
	}
	if err == nil {
		err = strictjson.Decode(data, &p)
	}
	if err == nil {
		err = p.validate()
	}
	if err != nil {
		return p, fmt.Errorf("invalid appearance settings: %w; run SINet ui reset", err)
	}
	return p, nil
}
func saveAppearance(p appearance) error {
	if err := p.validate(); err != nil {
		return err
	}
	path, err := appearancePath()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create appearance directory: %w", err)
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".sinet-ui-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = json.NewEncoder(f).Encode(p); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return fmt.Errorf("save appearance settings: %w", err)
	}
	return nil
}
func (a *App) uiCommand(args []string) error {
	if len(args) == 1 && args[0] == "reset" {
		if err := saveAppearance(defaultAppearance()); err != nil {
			return err
		}
		return a.print(defaultAppearance())
	}
	if len(args) == 1 && args[0] == "show" {
		p, err := loadAppearance()
		if err != nil {
			return err
		}
		return a.print(p)
	}
	if len(args) != 0 {
		return errors.New("usage: SINet ui [show|reset]")
	}
	for {
		p, err := loadAppearance()
		if err != nil {
			return err
		}
		choice, err := a.menu("Appearance", []string{"Theme: " + p.Theme, "Colour: " + p.Color, "Layout: " + p.Density, "Restore defaults", "Back"})
		if err != nil {
			return err
		}
		switch choice {
		case 0:
			var n int
			n, err = a.menu("Theme", []string{"Dark · signal cyan", "Light · signal blue", "Back"})
			if n < 2 {
				p.Theme = []string{"dark", "light"}[n]
			}
		case 1:
			var n int
			n, err = a.menu("Colour", []string{"Automatic terminal detection", "Always use colour", "No colour", "Back"})
			if n < 3 {
				p.Color = []string{"auto", "always", "never"}[n]
			}
		case 2:
			var n int
			n, err = a.menu("Layout", []string{"Comfortable", "Compact", "Back"})
			if n < 2 {
				p.Density = []string{"comfortable", "compact"}[n]
			}
		case 3:
			p = defaultAppearance()
		case 4:
			return nil
		}
		if err != nil {
			return err
		}
		if err = saveAppearance(p); err != nil {
			return err
		}
	}
}
func (a *App) settingsMenu() error {
	for {
		n, err := a.menu("Settings", []string{"Appearance", "Display name", "Bio", "Profile visibility", "Notifications", "Show account settings", "Version and launchers", "Back"})
		if err != nil {
			return err
		}
		switch n {
		case 0:
			err = a.uiCommand(nil)
		case 1, 2:
			key := []string{"display_name", "bio"}[n-1]
			var value string
			value, err = a.line("New " + key + " (empty cancels): ")
			if err == nil && value != "" {
				err = a.run([]string{"settings", "set", key, value})
			}
		case 3, 4:
			key, title, items := "public", "Profile visibility", []string{"Public", "Private", "Back"}
			if n == 4 {
				key, title, items = "notifications", "Notifications", []string{"Enabled", "Disabled", "Back"}
			}
			var value int
			value, err = a.menu(title, items)
			if err == nil && value < 2 {
				err = a.run([]string{"settings", "set", key, []string{"true", "false"}[value]})
			}
		case 5:
			err = a.run([]string{"settings"})
		case 6:
			err = a.doctor()
		case 7:
			return nil
		}
		if err != nil {
			return err
		}
	}
}
