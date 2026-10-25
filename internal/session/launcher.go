package session

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/dan-3200/hyprprofile/internal/storage"
	"github.com/dan-3200/hyprprofile/internal/validation"
)

// LaunchPlan is the validated process invocation for a managed Hyprland session.
type LaunchPlan struct {
	Executable string
	Arguments  []string
}

// PrepareLaunch validates the selected profile and constructs a safe compositor command.
func PrepareLaunch(store *storage.Store, manager Manager) (LaunchPlan, error) {
	return prepareLaunch(store, manager, exec.LookPath)
}

func prepareLaunch(store *storage.Store, manager Manager, lookup func(string) (string, error)) (LaunchPlan, error) {
	id, err := manager.Get()
	if err != nil {
		return LaunchPlan{}, fmt.Errorf("read active profile: %w", err)
	}
	if id == "" {
		return LaunchPlan{}, errors.New("no active profile is selected")
	}
	manifest, err := store.Load(id)
	if err != nil {
		return LaunchPlan{}, fmt.Errorf("load active profile: %w", err)
	}
	directory, err := store.ProfilePath(id)
	if err != nil {
		return LaunchPlan{}, fmt.Errorf("resolve active profile path: %w", err)
	}
	result := validation.ValidateProfile(directory, manifest, true)
	if !result.Valid() {
		return LaunchPlan{}, errors.New("active profile failed validation")
	}
	entrypoint, err := validation.ResolveEntrypoint(directory, manifest.Entrypoint)
	if err != nil {
		return LaunchPlan{}, fmt.Errorf("resolve active entrypoint: %w", err)
	}
	if executable, lookupErr := lookup("start-hyprland"); lookupErr == nil {
		return LaunchPlan{
			Executable: executable,
			Arguments:  []string{"start-hyprland", "--", "--config", entrypoint},
		}, nil
	}
	executable, lookupErr := lookup("Hyprland")
	if lookupErr != nil {
		return LaunchPlan{}, errors.New("neither start-hyprland nor Hyprland is available")
	}
	return LaunchPlan{
		Executable: executable,
		Arguments:  []string{"Hyprland", "--config", entrypoint},
	}, nil
}
