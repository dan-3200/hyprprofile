package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Paths contains all user-owned locations used by HyprProfile.
type Paths struct {
	ConfigHome        string
	StateHome         string
	ProfilesDir       string
	AppConfigDir      string
	StateDir          string
	ActiveProfileFile string
	TrashDir          string
	BackupsDir        string
	LogsDir           string
}

// PathsFromEnvironment resolves XDG paths using the current process environment.
func PathsFromEnvironment() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve home directory: %w", err)
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		stateHome = filepath.Join(home, ".local", "state")
	}
	return NewPaths(configHome, stateHome)
}

// NewPaths builds paths from absolute XDG config and state homes.
func NewPaths(configHome, stateHome string) (Paths, error) {
	if !filepath.IsAbs(configHome) || !filepath.IsAbs(stateHome) {
		return Paths{}, errors.New("XDG paths must be absolute")
	}

	configHome = filepath.Clean(configHome)
	stateHome = filepath.Clean(stateHome)
	stateDir := filepath.Join(stateHome, "hyprprofile")
	return Paths{
		ConfigHome:        configHome,
		StateHome:         stateHome,
		ProfilesDir:       filepath.Join(configHome, "hypr-profiles"),
		AppConfigDir:      filepath.Join(configHome, "hyprprofile"),
		StateDir:          stateDir,
		ActiveProfileFile: filepath.Join(stateDir, "active-profile"),
		TrashDir:          filepath.Join(stateDir, "trash"),
		BackupsDir:        filepath.Join(stateDir, "backups"),
		LogsDir:           filepath.Join(stateDir, "logs"),
	}, nil
}

// Ensure creates the user-owned directories with private permissions.
func (p Paths) Ensure() error {
	for _, dir := range []string{p.ProfilesDir, p.AppConfigDir, p.StateDir, p.TrashDir, p.BackupsDir, p.LogsDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}
