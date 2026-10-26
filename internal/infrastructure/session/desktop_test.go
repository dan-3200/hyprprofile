package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Contains(t, string(data), "Exec=\""+launcher+"\"")
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatalf("stat desktop entry: %v", err)
	}
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
	_, removeErr := RemoveDesktopEntry(directory)
	require.NoError(t, removeErr)
	_, err = os.Stat(destination)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestInstallDesktopEntryRejectsRelativeLauncher(t *testing.T) {
	_, err := InstallDesktopEntry(t.TempDir(), "hyprprofile-session-launcher")
	require.Error(t, err)
}
