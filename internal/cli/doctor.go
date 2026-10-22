package cli

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sinet"
	"strings"
)

// Inspect paths without executing possibly untrusted binaries from PATH or cwd.
func (a *App) doctor() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return err
	}
	digest := func(path string) string {
		f, err := os.Open(path)
		if err != nil {
			return "unreadable"
		}
		defer f.Close()
		h := sha256.New()
		if _, err = io.Copy(h, io.LimitReader(f, 128<<20)); err != nil {
			return "unreadable"
		}
		return fmt.Sprintf("%x", h.Sum(nil))
	}
	current := digest(executable)
	type launcher struct {
		Path   string `json:"path"`
		Target string `json:"target"`
		Same   bool   `json:"same_binary"`
	}
	launchers := []launcher{}
	seen := map[string]bool{}
	name := "SINet"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		path, err := filepath.Abs(filepath.Join(dir, name))
		if err != nil || seen[path] {
			continue
		}
		seen[path] = true
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		target, err := filepath.EvalSymlinks(path)
		if err != nil {
			continue
		}
		launchers = append(launchers, launcher{path, target, digest(target) == current})
	}
	ui, err := appearancePath()
	if err != nil {
		return err
	}
	return a.print(map[string]any{"version": sinet.Version(), "executable": executable, "sha256": current, "path_launchers": launchers, "appearance_config": ui, "guidance": "Use make install from the reviewed source checkout to update your user launcher. Explicit copies and pinned registry packages require separate updates."})
}
