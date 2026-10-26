package usecases

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/application/dto"
	"github.com/dan-3200/hyprprofile/internal/domain/profile"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/inspection"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/session"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/storage"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/templates"
)

func TestProfileServiceLifecycle(t *testing.T) {
	root := t.TempDir()
	paths, err := storage.NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	store := storage.New(paths)
	selection := session.NewManager(paths)
	stamp := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	service := NewProfileService(store, store, selection, inspection.DefaultInspector(), templates.Catalog{}, func() time.Time { return stamp })

	if status, selected, statusErr := service.Status(); statusErr != nil || selected {
		t.Fatalf("initial status = %v, %v, %v", status, selected, statusErr)
	}
	created, err := service.CreateProfile(dto.CreateProfileInput{ID: "first", Name: "First"})
	if err != nil || created.Template != profile.DefaultTemplate || !created.CreatedAt.Equal(stamp) || created.DirectoryID() != "first" || created.ID == "first" || created.FormatVersion != profile.CurrentFormat {
		t.Fatalf("create = %+v, %v", created, err)
	}
	cloned, err := service.CloneProfile(dto.CloneProfileInput{SourceID: "first", ID: "second"})
	if err != nil || cloned.Name != "First (clone)" || cloned.Origin != "first" || cloned.ID == created.ID || cloned.DirectoryID() != "second" {
		t.Fatalf("clone = %+v, %v", cloned, err)
	}
	assertWarningsDoNotBlock(t, service)
	if _, activateErr := service.ActivateProfile("first"); activateErr != nil {
		t.Fatalf("activate = %v", activateErr)
	}
	items, err := service.ListProfiles()
	if err != nil || len(items) != 2 || !items[0].Active || items[1].Active {
		t.Fatalf("list = %+v, %v", items, err)
	}
	assertArchiveAndRestore(t, service)
}

func assertWarningsDoNotBlock(t *testing.T, service *ProfileService) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
	validationResult, err := service.ValidateProfile("first")
	if err != nil || !validationResult.Valid() || len(validationResult.Issues) == 0 {
		t.Fatalf("missing command warnings = %+v, %v", validationResult, err)
	}
}

func TestLegacyProfileCanActivateAndCloneIntoUUIDFormat(t *testing.T) {
	root := t.TempDir()
	paths, err := storage.NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	store := storage.New(paths)
	legacy := profile.NewManifest("legacy", "Legacy", "", profile.DefaultEntrypoint, time.Now())
	if createErr := store.Create(legacy); createErr != nil {
		t.Fatal(createErr)
	}
	selection := session.NewManager(paths)
	service := NewProfileService(store, store, selection, inspection.DefaultInspector(), templates.Catalog{}, time.Now)
	if _, activateErr := service.ActivateProfile("legacy"); activateErr != nil {
		t.Fatalf("activate legacy: %v", activateErr)
	}
	if active, activeErr := selection.Get(); activeErr != nil || active != "legacy" {
		t.Fatalf("active legacy slug = %q, %v", active, activeErr)
	}
	clone, err := service.CloneProfile(dto.CloneProfileInput{SourceID: "legacy", ID: "modern"})
	if err != nil || clone.FormatVersion != profile.CurrentFormat || clone.DirectoryID() != "modern" || clone.ID == "legacy" {
		t.Fatalf("clone legacy = %+v, %v", clone, err)
	}
	if _, err := store.Load("legacy"); err != nil {
		t.Fatalf("load original legacy: %v", err)
	}
}

func assertArchiveAndRestore(t *testing.T, service *ProfileService) {
	t.Helper()
	if _, archiveErr := service.ArchiveProfile("first"); archiveErr == nil {
		t.Fatal("archived active profile")
	}
	archived, err := service.ArchiveProfile("second")
	if err != nil || !strings.Contains(archived.Destination, "second") {
		t.Fatalf("archive = %+v, %v", archived, err)
	}
	if err := service.RestoreProfile("second"); err != nil {
		t.Fatalf("restore = %v", err)
	}
	if status, selected, statusErr := service.Status(); statusErr != nil || !selected || status.DirectoryID() != "first" {
		t.Fatalf("status = %+v, %v, %v", status, selected, statusErr)
	}
}

func TestProfileServiceImportAndValidation(t *testing.T) {
	root := t.TempDir()
	paths, err := storage.NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	store := storage.New(paths)
	selection := session.NewManager(paths)
	service := NewProfileService(store, store, selection, inspection.DefaultInspector(), templates.Catalog{}, time.Now)
	source := filepath.Join(root, "source")
	if mkdirErr := os.Mkdir(source, 0o700); mkdirErr != nil {
		t.Fatal(mkdirErr)
	}
	if writeErr := os.WriteFile(filepath.Join(source, "hyprland.lua"), []byte("-- config\n"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	imported, err := service.ImportProfile(dto.ImportProfileInput{Source: source, ID: "imported"})
	if err != nil || imported.Name != "imported" || imported.Origin != source || imported.Entrypoint != "hyprland.lua" || imported.DirectoryID() != "imported" {
		t.Fatalf("import = %+v, %v", imported, err)
	}
	stored, err := store.Load("imported")
	if err != nil || stored.Entrypoint != "hyprland.lua" {
		t.Fatalf("stored import = %+v, %v", stored, err)
	}
	result, err := service.ValidateProfile("imported")
	if err != nil || !result.Valid() {
		t.Fatalf("validation = %+v, %v", result, err)
	}
	if removeErr := os.Remove(filepath.Join(paths.ProfilesDir, "imported", "hyprland.lua")); removeErr != nil {
		t.Fatal(removeErr)
	}
	result, err = service.ValidateProfile("imported")
	if err != nil || result.Valid() {
		t.Fatalf("invalid validation = %+v, %v", result, err)
	}
	if _, err := service.ActivateProfile("imported"); err == nil {
		t.Fatal("activated invalid profile")
	}
	if active, err := selection.Get(); err != nil || active != "" {
		t.Fatalf("selection after failed activation = %q, %v", active, err)
	}
}
