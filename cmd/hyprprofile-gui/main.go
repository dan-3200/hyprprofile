// HyprProfile GUI manages local profiles through a native GNOME interface.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dan-3200/hyprprofile/internal/application/profiles"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/inspection"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/session"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/storage"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/templates"
	"github.com/dan-3200/hyprprofile/internal/presentation/gui"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "hyprprofile-gui:", err)
		os.Exit(1)
	}
}

func run() error {
	paths, err := storage.PathsFromEnvironment()
	if err != nil {
		return fmt.Errorf("resolve XDG paths: %w", err)
	}
	store := storage.New(paths)
	service := profiles.NewProfileService(store, store, session.NewManager(paths), inspection.DefaultInspector(), templates.Catalog{}, time.Now)
	if err := gui.Run(service); err != nil {
		return fmt.Errorf("run GUI: %w", err)
	}
	return nil
}
