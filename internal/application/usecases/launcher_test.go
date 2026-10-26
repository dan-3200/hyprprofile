package usecases

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
	plan, err := PrepareLaunch(store, manager, inspection.DefaultInspector(), lookup)
	if err != nil {
		t.Fatalf("prepareLaunch() error = %v", err)
	}
	if plan.Executable != "/usr/bin/start-hyprland" {
		t.Fatalf("executable = %q", plan.Executable)
	}
	if len(plan.Arguments) != 4 || plan.Arguments[1] != "--" || plan.Arguments[2] != "--config" {
		t.Fatalf("arguments = %#v", plan.Arguments)
	}
	if err := os.Remove(filepath.Join(paths.ProfilesDir, manifest.DirectoryID(), manifest.Entrypoint)); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareLaunch(store, manager, inspection.DefaultInspector(), lookup); err == nil {
		t.Fatal("launcher accepted a missing active configuration")
	}
}
