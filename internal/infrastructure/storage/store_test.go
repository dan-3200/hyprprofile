package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/domain/profile"
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
	if loaded.Template != profile.DefaultTemplate {
		t.Fatalf("template = %q, want %q", loaded.Template, profile.DefaultTemplate)
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

func TestCreateFromTemplateRecordsAndWritesSelection(t *testing.T) {
	store := testStore(t)
	manifest := testManifest("bare-profile", "Bare profile")
	if err := store.CreateFromTemplate(manifest, "bare"); err != nil {
		t.Fatalf("CreateFromTemplate() error = %v", err)
	}
	loaded, err := store.Load(manifest.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Template != "bare" {
		t.Fatalf("template = %q, want bare", loaded.Template)
	}
	data, err := os.ReadFile(filepath.Join(store.Paths.ProfilesDir, manifest.ID, manifest.Entrypoint))
	if err != nil {
		t.Fatalf("read entrypoint: %v", err)
	}
	if !strings.Contains(string(data), "without optional desktop components") {
		t.Fatalf("entrypoint does not contain bare template marker")
	}
}

func TestCreateFromTemplateRejectsUnknownTemplate(t *testing.T) {
	store := testStore(t)
	manifest := testManifest("unknown-template", "Unknown template")
	if err := store.CreateFromTemplate(manifest, "unknown"); err == nil {
		t.Fatal("CreateFromTemplate() unexpectedly accepted an unknown template")
	}
}

func TestCreateFromTemplateRequiresLuaEntrypoint(t *testing.T) {
	store := testStore(t)
	manifest := testManifest("legacy-entrypoint", "Legacy entrypoint")
	manifest.Entrypoint = "hyprland.conf"
	if err := store.CreateFromTemplate(manifest, "minimal"); err == nil {
		t.Fatal("CreateFromTemplate() unexpectedly accepted a non-Lua entrypoint")
	}
}

func TestCreateFromTemplateWritesMultipleFiles(t *testing.T) {
	store := testStore(t)
	manifest := testManifest("serpantinum-profile", "Serpantinum profile")
	if err := store.CreateFromTemplate(manifest, "serpantinum"); err != nil {
		t.Fatalf("CreateFromTemplate() error = %v", err)
	}
	for _, relative := range []string{"hyprland.lua", "config/keybinds.lua", "config/settings.lua"} {
		if _, err := os.Stat(filepath.Join(store.Paths.ProfilesDir, manifest.ID, relative)); err != nil {
			t.Errorf("template file %q is unavailable: %v", relative, err)
		}
	}
	loaded, err := store.Load(manifest.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Template != "serpantinum" {
		t.Fatalf("template = %q, want serpantinum", loaded.Template)
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

func TestLegacyManifestLoadsWithoutRewrite(t *testing.T) {
	store := testStore(t)
	legacy := testManifest("old-profile", "Old profile")
	if err := store.Create(legacy); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Paths.ProfilesDir, legacy.ID, profile.ManifestFile)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load(legacy.ID)
	if err != nil || loaded.FormatVersion != profile.LegacyFormat || loaded.DirectoryID() != legacy.ID {
		t.Fatalf("loaded legacy manifest = %+v, %v", loaded, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("legacy manifest changed during load: %v", err)
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
