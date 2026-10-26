package session

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DesktopEntry adapts the Wayland session file operations.
type DesktopEntry struct{}

// Install publishes the Wayland session entry.
func (DesktopEntry) Install(directory, launcher string) (string, error) {
	return InstallDesktopEntry(directory, launcher)
}

// Remove deletes the Wayland session entry.
func (DesktopEntry) Remove(directory string) (string, error) {
	return RemoveDesktopEntry(directory)
}

// LauncherResolver discovers the session launcher from an explicit or installed path.
type LauncherResolver struct {
	Lookup     func(string) (string, error)
	Executable func() (string, error)
}

// DefaultLauncherResolver uses the current process environment.
func DefaultLauncherResolver() LauncherResolver {
	return LauncherResolver{Lookup: exec.LookPath, Executable: os.Executable}
}

// Resolve finds a launcher using the explicit path or installed locations.
func (r LauncherResolver) Resolve(explicit string) (string, error) {
	if explicit != "" {
		if !filepath.IsAbs(explicit) {
			return "", errors.New("launcher path must be absolute")
		}
		return filepath.Clean(explicit), nil
	}
	current, err := r.Executable()
	if err == nil {
		sibling := filepath.Join(filepath.Dir(current), "hyprprofile-session-launcher")
		if info, statErr := os.Stat(sibling); statErr == nil && info.Mode().IsRegular() && info.Mode()&0o111 != 0 {
			return sibling, nil
		}
	}
	launcher, err := r.Lookup("hyprprofile-session-launcher")
	if err != nil {
		return "", errors.New("hyprprofile-session-launcher was not found; install it or use --launcher")
	}
	launcher, err = filepath.Abs(launcher)
	if err != nil {
		return "", fmt.Errorf("resolve launcher path: %w", err)
	}
	return launcher, nil
}
