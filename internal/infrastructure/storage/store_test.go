package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/domain/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, base.Name, loaded.Name)
	assert.Equal(t, profile.DefaultTemplate, loaded.Template)

	clone := testManifest("serpantinum", "Serpantinum")
	clone.Entrypoint = base.Entrypoint
	cloneErr := store.Clone("minimal", clone)
	require.NoError(t, cloneErr)
	_, archiveErr := store.Archive("minimal")
	require.NoError(t, archiveErr)
	_, err = store.Load("minimal")
	require.Error(t, err)
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
	assert.Equal(t, "bare", loaded.Template)
	data, err := os.ReadFile(filepath.Join(store.Paths.ProfilesDir, manifest.ID, manifest.Entrypoint))
	if err != nil {
		t.Fatalf("read entrypoint: %v", err)
	}
	assert.Contains(t, string(data), "without optional desktop components")
}

func TestCreateFromTemplateRejectsUnknownTemplate(t *testing.T) {
	store := testStore(t)
	manifest := testManifest("unknown-template", "Unknown template")
	require.Error(t, store.CreateFromTemplate(manifest, "unknown"))
}

func TestCreateFromTemplateRequiresLuaEntrypoint(t *testing.T) {
	store := testStore(t)
	manifest := testManifest("legacy-entrypoint", "Legacy entrypoint")
	manifest.Entrypoint = "hyprland.conf"
	require.Error(t, store.CreateFromTemplate(manifest, "minimal"))
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
