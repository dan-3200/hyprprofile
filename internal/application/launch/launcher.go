// Package launch prepares a validated Hyprland process invocation.
package launch

import (
	"errors"
	"fmt"
)

// Prepare validates the selected profile and constructs a compositor command.
func Prepare(store ProfileReader, manager ActiveReader, inspector Inspector, lookup ExecutableLookup) (Plan, error) {
	id, err := manager.Get()
	if err != nil {
		return Plan{}, fmt.Errorf("read active profile: %w", err)
	}
	if id == "" {
		return Plan{}, errors.New("no active profile is selected")
	}
	manifest, err := store.Load(id)
	if err != nil {
		return Plan{}, fmt.Errorf("load active profile: %w", err)
	}
	directory, err := store.ProfilePath(id)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve active profile path: %w", err)
	}
	result := inspector.Validate(directory, manifest)
	if !result.Valid() {
		return Plan{}, errors.New("active profile failed validation")
	}
	entrypoint, err := inspector.ResolveEntrypoint(directory, manifest.Entrypoint)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve active entrypoint: %w", err)
	}
	if executable, lookupErr := lookup("start-hyprland"); lookupErr == nil {
		return Plan{
			Executable: executable,
			Arguments:  []string{"start-hyprland", "--", "--config", entrypoint},
		}, nil
	}
	executable, lookupErr := lookup("Hyprland")
	if lookupErr != nil {
		return Plan{}, errors.New("neither start-hyprland nor Hyprland is available")
	}
	return Plan{
		Executable: executable,
		Arguments:  []string{"Hyprland", "--config", entrypoint},
	}, nil
}
