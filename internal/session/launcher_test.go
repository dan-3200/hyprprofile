package session

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/profile"
	"github.com/dan-3200/hyprprofile/internal/storage"
)

func TestPrepareLaunchUsesStartHyprland(t *testing.T) {
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
	manager := NewManager(paths)
	if activateErr := manager.Set(manifest.DirectoryID()); activateErr != nil {
		t.Fatalf("Set() error = %v", activateErr)
	}
	lookup := func(command string) (string, error) {
		if command == "start-hyprland" {
			return "/usr/bin/start-hyprland", nil
		}
		return "", fmt.Errorf("%s unavailable", command)
	}
	plan, err := prepareLaunch(store, manager, lookup)
	if err != nil {
		t.Fatalf("prepareLaunch() error = %v", err)
	}
	if plan.Executable != "/usr/bin/start-hyprland" {
		t.Fatalf("executable = %q", plan.Executable)
	}
	if len(plan.Arguments) != 4 || plan.Arguments[1] != "--" || plan.Arguments[2] != "--config" {
		t.Fatalf("arguments = %#v", plan.Arguments)
	}
}
