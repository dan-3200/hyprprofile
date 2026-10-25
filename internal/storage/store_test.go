package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/profile"
)

func TestCreateCloneArchiveAndRestore(t *testing.T) {
	store := testStore(t)
	base := testManifest("minimal", "Minimal")
	if err := store.Create(base); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	loaded, err := store.Load("minimal")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Name != base.Name {
		t.Fatalf("loaded name = %q, want %q", loaded.Name, base.Name)
	}

	clone := testManifest("serpantinum", "Serpantinum")
	clone.Entrypoint = base.Entrypoint
	if err := store.Clone("minimal", clone); err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	if _, err := store.Archive("minimal"); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}
	if _, err := store.Load("minimal"); err == nil {
		t.Fatal("Load(minimal) unexpectedly succeeded after archive")
	}
	if err := store.Restore("minimal"); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if _, err := store.Load("minimal"); err != nil {
		t.Fatalf("Load(minimal) after restore error = %v", err)
	}
}

func TestImportRejectsSymbolicLinks(t *testing.T) {
	store := testStore(t)
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "hyprland.lua"), []byte("-- config\n"), 0o600); err != nil {
		t.Fatalf("write entrypoint: %v", err)
	}
	if err := os.Symlink("/etc/passwd", filepath.Join(source, "outside")); err != nil {
		t.Fatalf("create symbolic link: %v", err)
	}
	manifest := testManifest("imported", "Imported")
	if err := store.Import(source, manifest); err == nil {
		t.Fatal("Import() unexpectedly accepted a symbolic link")
	}
}

func testStore(t *testing.T) *Store {
	t.Helper()
	root := t.TempDir()
	paths, err := NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatalf("NewPaths() error = %v", err)
	}
	return New(paths)
}

func testManifest(id, name string) profile.Manifest {
	return profile.NewManifest(id, name, "", profile.DefaultEntrypoint, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
}
