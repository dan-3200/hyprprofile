package usecases

import (
	"errors"
	"fmt"

	"github.com/dan-3200/hyprprofile/internal/application/dto"
	"github.com/dan-3200/hyprprofile/internal/application/ports"
)

// PrepareLaunch validates the selected profile and constructs a compositor command.
func PrepareLaunch(store ports.ProfileReader, manager ports.ActiveReader, inspector ports.ProfileInspector, lookup ports.ExecutableLookup) (dto.LaunchPlan, error) {
	id, err := manager.Get()
	if err != nil {
		return dto.LaunchPlan{}, fmt.Errorf("read active profile: %w", err)
	}
	if id == "" {
		return dto.LaunchPlan{}, errors.New("no active profile is selected")
	}
	manifest, err := store.Load(id)
	if err != nil {
		return dto.LaunchPlan{}, fmt.Errorf("load active profile: %w", err)
	}
	directory, err := store.ProfilePath(id)
	if err != nil {
		return dto.LaunchPlan{}, fmt.Errorf("resolve active profile path: %w", err)
	}
	result := inspector.Validate(directory, manifest)
	if !result.Valid() {
		return dto.LaunchPlan{}, errors.New("active profile failed validation")
	}
	entrypoint, err := inspector.ResolveEntrypoint(directory, manifest.Entrypoint)
	if err != nil {
		return dto.LaunchPlan{}, fmt.Errorf("resolve active entrypoint: %w", err)
	}
	if executable, lookupErr := lookup("start-hyprland"); lookupErr == nil {
		return dto.LaunchPlan{
			Executable: executable,
			Arguments:  []string{"start-hyprland", "--", "--config", entrypoint},
		}, nil
	}
	executable, lookupErr := lookup("Hyprland")
	if lookupErr != nil {
		return dto.LaunchPlan{}, errors.New("neither start-hyprland nor Hyprland is available")
	}
	return dto.LaunchPlan{
		Executable: executable,
		Arguments:  []string{"Hyprland", "--config", entrypoint},
	}, nil
}
