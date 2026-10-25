// HyprProfile session launcher starts the active, validated Hyprland profile.
package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/dan-3200/hyprprofile/internal/session"
	"github.com/dan-3200/hyprprofile/internal/storage"
)

func main() {
	if err := launch(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "hyprprofile-session-launcher:", err)
		os.Exit(1)
	}
}

func launch() error {
	paths, err := storage.PathsFromEnvironment()
	if err != nil {
		return fmt.Errorf("resolve XDG paths: %w", err)
	}
	store := storage.New(paths)
	plan, err := session.PrepareLaunch(store, session.NewManager(paths))
	if err != nil {
		return fmt.Errorf("prepare Hyprland launch: %w", err)
	}
	// #nosec G204 -- PrepareLaunch resolves a fixed compositor executable and validates the entrypoint.
	if err := syscall.Exec(plan.Executable, plan.Arguments, os.Environ()); err != nil {
		return fmt.Errorf("execute Hyprland: %w", err)
	}
	return nil
}
