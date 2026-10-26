package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallAndRemoveDesktopEntry(t *testing.T) {
	root := t.TempDir()
	launcher := filepath.Join(root, "hyprprofile-session-launcher")
	if err := os.WriteFile(launcher, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write launcher: %v", err)
	}
	directory := filepath.Join(root, "wayland-sessions")
	destination, err := InstallDesktopEntry(directory, launcher)
	if err != nil {
		t.Fatalf("InstallDesktopEntry() error = %v", err)
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read desktop entry: %v", err)
	}
	if !strings.Contains(string(data), "Exec=\""+launcher+"\"") {
		t.Fatalf("desktop entry = %q", data)
	}
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatalf("stat desktop entry: %v", err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("desktop entry mode = %o", info.Mode().Perm())
	}
	if _, err := RemoveDesktopEntry(directory); err != nil {
		t.Fatalf("RemoveDesktopEntry() error = %v", err)
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("desktop entry still exists: %v", err)
	}
}

func TestInstallDesktopEntryRejectsRelativeLauncher(t *testing.T) {
	if _, err := InstallDesktopEntry(t.TempDir(), "hyprprofile-session-launcher"); err == nil {
		t.Fatal("InstallDesktopEntry() accepted a relative launcher")
	}
}
