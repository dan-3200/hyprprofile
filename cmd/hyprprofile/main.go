// HyprProfile manages self-contained Hyprland profiles.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dan-3200/hyprprofile/internal/app"
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
	profiles := app.NewProfileService(store, store, manager, time.Now)
	system := app.DefaultSystemService(store)

	if len(os.Args) == 1 {
		if err := tui.Run(profiles); err != nil {
			return fmt.Errorf("run dashboard: %w", err)
		}
		return nil
	}
	return run(os.Args[1:], profiles, system)
}
