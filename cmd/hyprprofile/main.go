// HyprProfile manages self-contained Hyprland profiles.
package main

import (
	"fmt"
	"os"

	"github.com/dan-3200/hyprprofile/internal/session"
	"github.com/dan-3200/hyprprofile/internal/storage"
	"github.com/dan-3200/hyprprofile/internal/tui"
)

func main() {
	if err := mainError(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "hyprprofile:", err)
		os.Exit(1)
	}
}

func mainError() error {
	paths, err := storage.PathsFromEnvironment()
	if err != nil {
		return fmt.Errorf("resolve XDG paths: %w", err)
	}
	store := storage.New(paths)
	manager := session.NewManager(paths)

	if len(os.Args) == 1 {
		if err := tui.Run(store, manager); err != nil {
			return fmt.Errorf("run dashboard: %w", err)
		}
		return nil
	}
	return run(os.Args[1:], store, manager)
}
