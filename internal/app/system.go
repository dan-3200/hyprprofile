package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dan-3200/hyprprofile/internal/session"
)

// StoragePreparer creates the directories checked by doctor.
type StoragePreparer interface {
	Ensure() error
}

// CommandStatus describes whether a diagnostic executable is available.
type CommandStatus struct {
	Name      string
	Available bool
}

// SystemService coordinates local installation and diagnostics.
type SystemService struct {
	storage    StoragePreparer
	lookup     func(string) (string, error)
	executable func() (string, error)
}

// NewSystemService wires system dependencies.
func NewSystemService(storage StoragePreparer, lookup func(string) (string, error), executable func() (string, error)) *SystemService {
	return &SystemService{storage: storage, lookup: lookup, executable: executable}
}

// Doctor checks storage preparation and expected executables.
func (s *SystemService) Doctor() ([]CommandStatus, error) {
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
func (s *SystemService) InstallSession(directory, launcher string) (string, error) {
	resolved, err := s.resolveSessionLauncher(launcher)
	if err != nil {
		return "", err
	}
	destination, err := session.InstallDesktopEntry(directory, resolved)
	if err != nil {
		return "", fmt.Errorf("install Wayland session: %w", err)
	}
	return destination, nil
}

// UninstallSession removes the display-manager entry.
func (s *SystemService) UninstallSession(directory string) (string, error) {
	destination, err := session.RemoveDesktopEntry(directory)
	if err != nil {
		return "", fmt.Errorf("uninstall Wayland session: %w", err)
	}
	return destination, nil
}

func (s *SystemService) resolveSessionLauncher(explicit string) (string, error) {
	if explicit != "" {
		if !filepath.IsAbs(explicit) {
			return "", errors.New("launcher path must be absolute")
		}
		return filepath.Clean(explicit), nil
	}
	current, err := s.executable()
	if err == nil {
		sibling := filepath.Join(filepath.Dir(current), "hyprprofile-session-launcher")
		if info, statErr := os.Stat(sibling); statErr == nil && info.Mode().IsRegular() && info.Mode()&0o111 != 0 {
			return sibling, nil
		}
	}
	launcher, err := s.lookup("hyprprofile-session-launcher")
	if err != nil {
		return "", errors.New("hyprprofile-session-launcher was not found; install it or use --launcher")
	}
	launcher, err = filepath.Abs(launcher)
	if err != nil {
		return "", fmt.Errorf("resolve launcher path: %w", err)
	}
	return launcher, nil
}

// DefaultSystemService uses the process environment for executable discovery.
func DefaultSystemService(storage StoragePreparer) *SystemService {
	return NewSystemService(storage, exec.LookPath, os.Executable)
}
