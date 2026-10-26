// HyprProfile manages self-contained Hyprland profiles.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	app "github.com/dan-3200/hyprprofile/internal/application/usecases"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/inspection"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/session"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/storage"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/templates"
	"github.com/dan-3200/hyprprofile/internal/presentation/cli"
	"github.com/dan-3200/hyprprofile/internal/presentation/tui"
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
	profiles := app.NewProfileService(store, store, manager, inspection.DefaultInspector(), templates.Catalog{}, time.Now)
	system := app.NewSystemService(store, exec.LookPath, session.DesktopEntry{}, session.DefaultLauncherResolver())

	if len(os.Args) == 1 {
		if err := tui.Run(profiles); err != nil {
			return fmt.Errorf("run dashboard: %w", err)
		}
		return nil
	}
	if err := cli.Run(os.Args[1:], profiles, system, session.DefaultSessionDirectory); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
