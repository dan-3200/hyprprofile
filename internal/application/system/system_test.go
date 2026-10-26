package system

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dan-3200/hyprprofile/internal/infrastructure/session"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemServiceDoctorAndSessionEntry(t *testing.T) {
	root := t.TempDir()
	paths, err := storage.NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(root, "hyprprofile-session-launcher")
	if writeErr := os.WriteFile(launcher, []byte("#!/bin/sh\n"), 0o700); writeErr != nil {
		t.Fatal(writeErr)
	}
	lookup := func(name string) (string, error) {
		if name == "Hyprland" {
			return "/usr/bin/Hyprland", nil
		}
		return "", errors.New("missing")
	}
	resolver := session.LauncherResolver{Lookup: lookup, Executable: func() (string, error) { return filepath.Join(root, "hyprprofile"), nil }}
	service := NewService(storage.New(paths), lookup, session.DesktopEntry{}, resolver)
	report, err := service.Doctor()
	require.NoError(t, err)
	require.Len(t, report, 3)
	assert.True(t, report[0].Available)
	assert.False(t, report[1].Available)
	assert.False(t, report[2].Available)
	directory := filepath.Join(root, "wayland-sessions")
	destination, err := service.InstallSession(directory, "")
	if err != nil {
		t.Fatalf("install = %v", err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || !strings.Contains(string(data), launcher) {
		t.Fatalf("session entry = %q, %v", data, err)
	}
	removed, err := service.UninstallSession(directory)
	require.NoError(t, err)
	assert.Equal(t, destination, removed)
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("session entry still exists: %v", err)
	}
	if _, err := service.InstallSession(directory, "relative-launcher"); err == nil {
		t.Fatal("accepted a relative launcher")
	}
}
