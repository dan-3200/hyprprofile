// HyprProfile session launcher starts the active, validated Hyprland profile.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/dan-3200/hyprprofile/internal/application/usecases"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/inspection"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/session"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/storage"
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
	plan, err := usecases.PrepareLaunch(store, session.NewManager(paths), inspection.DefaultInspector(), exec.LookPath)
	if err != nil {
		return fmt.Errorf("prepare Hyprland launch: %w", err)
	}
	// #nosec G204 -- PrepareLaunch resolves a fixed compositor executable and validates the entrypoint.
	if err := syscall.Exec(plan.Executable, plan.Arguments, os.Environ()); err != nil {
		return fmt.Errorf("execute Hyprland: %w", err)
	}
	return nil
}
