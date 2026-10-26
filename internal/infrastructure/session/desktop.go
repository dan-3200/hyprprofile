package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultSessionDirectory is the conventional system Wayland session directory.
	DefaultSessionDirectory = "/usr/share/wayland-sessions"
	// SessionDesktopFile is the display manager entry owned by HyprProfile.
	SessionDesktopFile = "hyprprofile.desktop"
)

// InstallDesktopEntry publishes a Wayland session entry for the launcher.
func InstallDesktopEntry(directory, launcher string) (string, error) {
	launcher, err := validateLauncher(launcher)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(directory) {
		return "", errors.New("session directory must be absolute")
	}
	// #nosec G301 -- display managers must be able to traverse the public session directory.
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", fmt.Errorf("create session directory: %w", err)
	}

	destination := filepath.Join(filepath.Clean(directory), SessionDesktopFile)
	if info, statErr := os.Lstat(destination); statErr == nil && !info.Mode().IsRegular() {
		return "", fmt.Errorf("refusing to replace non-regular session entry %q", destination)
	} else if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
		return "", fmt.Errorf("inspect session entry: %w", statErr)
	}

	contents := []byte("[Desktop Entry]\n" +
		"Name=HyprProfile\n" +
		"Comment=Start Hyprland with the active HyprProfile\n" +
		"Exec=\"" + escapeDesktopExec(launcher) + "\"\n" +
		"Type=Application\n" +
		"DesktopNames=Hyprland\n")
	if err := writePublicAtomicFile(destination, contents); err != nil {
		return "", err
	}
	return destination, nil
}

// RemoveDesktopEntry removes the Wayland session entry owned by HyprProfile.
func RemoveDesktopEntry(directory string) (string, error) {
	if !filepath.IsAbs(directory) {
		return "", errors.New("session directory must be absolute")
	}
	destination := filepath.Join(filepath.Clean(directory), SessionDesktopFile)
	info, err := os.Lstat(destination)
	if errors.Is(err, fs.ErrNotExist) {
		return destination, nil
	}
	if err != nil {
		return "", fmt.Errorf("inspect session entry: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("refusing to remove non-regular session entry %q", destination)
	}
	if err := os.Remove(destination); err != nil {
		return "", fmt.Errorf("remove session entry: %w", err)
	}
	return destination, nil
}

func validateLauncher(launcher string) (string, error) {
	if !filepath.IsAbs(launcher) {
		return "", errors.New("launcher path must be absolute")
	}
	launcher = filepath.Clean(launcher)
	if strings.ContainsAny(launcher, "\r\n") {
		return "", errors.New("launcher path contains a line break")
	}
	info, err := os.Lstat(launcher)
	if err != nil {
		return "", fmt.Errorf("inspect launcher: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return "", fmt.Errorf("launcher %q is not an executable regular file", launcher)
	}
	return launcher, nil
}

func escapeDesktopExec(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"$", "\\$",
		"`", "\\`",
		"%", "%%",
	)
	return replacer.Replace(value)
}

func writePublicAtomicFile(path string, data []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".hyprprofile-session-")
	if err != nil {
		return fmt.Errorf("create temporary session entry: %w", err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set session entry permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write session entry: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync session entry: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close session entry: %w", err)
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("publish session entry: %w", err)
	}
	return nil
}
