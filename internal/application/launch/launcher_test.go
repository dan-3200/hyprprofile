package launch

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/domain/profile"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/inspection"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/session"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareUsesStartHyprland(t *testing.T) {
	root := t.TempDir()
	paths, err := storage.NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatalf("NewPaths() error = %v", err)
	}
	store := storage.New(paths)
	manifest := profile.NewUUIDManifest("minimal", "Minimal", "", profile.DefaultEntrypoint, time.Now())
	if createErr := store.Create(manifest); createErr != nil {
		t.Fatalf("Create() error = %v", createErr)
	}
	manager := session.NewManager(paths)
	if activateErr := manager.Set(manifest.DirectoryID()); activateErr != nil {
		t.Fatalf("Set() error = %v", activateErr)
	}
	lookup := func(command string) (string, error) {
		if command == "start-hyprland" {
			return "/usr/bin/start-hyprland", nil
		}
		return "", fmt.Errorf("%s unavailable", command)
	}
	plan, err := Prepare(store, manager, inspection.DefaultInspector(), lookup)
	require.NoError(t, err)
	assert.Equal(t, "/usr/bin/start-hyprland", plan.Executable)
	assert.Equal(t, []string{"start-hyprland", "--", "--config", filepath.Join(paths.ProfilesDir, manifest.DirectoryID(), manifest.Entrypoint)}, plan.Arguments)
	removeErr := os.Remove(filepath.Join(paths.ProfilesDir, manifest.DirectoryID(), manifest.Entrypoint))
	require.NoError(t, removeErr)
	_, err = Prepare(store, manager, inspection.DefaultInspector(), lookup)
	require.Error(t, err)
}
