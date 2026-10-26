// Package system groups application use cases for system integration.
package system

import (
	"fmt"
)

// Service coordinates local installation and diagnostics.
type Service struct {
	storage  StoragePreparer
	lookup   ExecutableLookup
	entry    SessionEntry
	resolver LauncherResolver
}

// NewService wires system dependencies.
func NewService(storage StoragePreparer, lookup ExecutableLookup, entry SessionEntry, resolver LauncherResolver) *Service {
	return &Service{storage: storage, lookup: lookup, entry: entry, resolver: resolver}
}

// Doctor checks storage preparation and expected executables.
func (s *Service) Doctor() ([]CommandStatus, error) {
	if err := s.storage.Ensure(); err != nil {
		return nil, fmt.Errorf("prepare profile storage: %w", err)
	}
	report := make([]CommandStatus, 0, 3)
	for _, command := range []string{"Hyprland", "hyprctl", "start-hyprland"} {
		_, err := s.lookup(command)
		report = append(report, CommandStatus{Name: command, Available: err == nil})
	}
	return report, nil
}

// InstallSession publishes the display-manager entry with an explicit or resolved launcher.
func (s *Service) InstallSession(directory, launcher string) (string, error) {
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
func (s *Service) UninstallSession(directory string) (string, error) {
	destination, err := s.entry.Remove(directory)
	if err != nil {
		return "", fmt.Errorf("uninstall Wayland session: %w", err)
	}
	return destination, nil
}
