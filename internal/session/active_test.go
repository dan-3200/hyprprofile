package session

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/profile"
	"github.com/dan-3200/hyprprofile/internal/storage"
)

func TestActivateWritesAndReadsProfileID(t *testing.T) {
	root := t.TempDir()
	paths, err := storage.NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatalf("NewPaths() error = %v", err)
	}
	store := storage.New(paths)
	manifest := profile.NewManifest("minimal", "Minimal", "", profile.DefaultEntrypoint, time.Now())
	if createErr := store.Create(manifest); createErr != nil {
		t.Fatalf("Create() error = %v", createErr)
	}
	manager := NewManager(paths)
	if activateErr := manager.Activate(store, manifest.ID); activateErr != nil {
		t.Fatalf("Activate() error = %v", activateErr)
	}
	active, err := manager.Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if active != manifest.ID {
		t.Fatalf("active profile = %q, want %q", active, manifest.ID)
	}
}
