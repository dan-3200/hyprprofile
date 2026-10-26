package usecases

import (
	"fmt"

	"github.com/dan-3200/hyprprofile/internal/application/dto"
	"github.com/dan-3200/hyprprofile/internal/application/ports"
)

// SystemService coordinates local installation and diagnostics.
type SystemService struct {
	storage  ports.StoragePreparer
	lookup   ports.ExecutableLookup
	entry    ports.SessionEntry
	resolver ports.LauncherResolver
}

// NewSystemService wires system dependencies.
func NewSystemService(storage ports.StoragePreparer, lookup ports.ExecutableLookup, entry ports.SessionEntry, resolver ports.LauncherResolver) *SystemService {
	return &SystemService{storage: storage, lookup: lookup, entry: entry, resolver: resolver}
}

// Doctor checks storage preparation and expected executables.
func (s *SystemService) Doctor() ([]dto.CommandStatus, error) {
	if err := s.storage.Ensure(); err != nil {
		return nil, fmt.Errorf("prepare profile storage: %w", err)
	}
	report := make([]dto.CommandStatus, 0, 3)
	for _, command := range []string{"Hyprland", "hyprctl", "start-hyprland"} {
		_, err := s.lookup(command)
		report = append(report, dto.CommandStatus{Name: command, Available: err == nil})
	}
	return report, nil
}

// InstallSession publishes the display-manager entry with an explicit or resolved launcher.
func (s *SystemService) InstallSession(directory, launcher string) (string, error) {
	resolved, err := s.resolver.Resolve(launcher)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	destination, err := s.entry.Install(directory, resolved)
	if err != nil {
		return "", fmt.Errorf("install Wayland session: %w", err)
	}
	return destination, nil
}

// UninstallSession removes the display-manager entry.
func (s *SystemService) UninstallSession(directory string) (string, error) {
	destination, err := s.entry.Remove(directory)
	if err != nil {
		return "", fmt.Errorf("uninstall Wayland session: %w", err)
	}
	return destination, nil
}
