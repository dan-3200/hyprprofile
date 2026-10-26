package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/dan-3200/hyprprofile/internal/profile"
	"github.com/pelletier/go-toml/v2"
)

// Store owns profile directories and manifests.
type Store struct {
	Paths Paths
	now   func() time.Time
}

// New creates a profile store.
func New(paths Paths) *Store {
	return &Store{Paths: paths, now: time.Now}
}

// Ensure creates storage roots before an operation.
func (s *Store) Ensure() error {
	return s.Paths.Ensure()
}

// ProfilePath returns the validated path to a profile directory.
func (s *Store) ProfilePath(id string) (string, error) {
	if err := profile.ValidateID(id); err != nil {
		return "", fmt.Errorf("validate profile ID: %w", err)
	}
	return filepath.Join(s.Paths.ProfilesDir, id), nil
}

// List returns readable profile manifests ordered by directory slug.
func (s *Store) List() ([]profile.Manifest, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.Paths.ProfilesDir)
	if err != nil {
		return nil, fmt.Errorf("read profiles directory: %w", err)
	}

	manifests := make([]profile.Manifest, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		manifest, err := s.Load(entry.Name())
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, manifest)
	}
	sort.Slice(manifests, func(i, j int) bool { return manifests[i].DirectoryID() < manifests[j].DirectoryID() })
	return manifests, nil
}

// Load reads and validates a profile manifest.
func (s *Store) Load(id string) (profile.Manifest, error) {
	dir, err := s.ProfilePath(id)
	if err != nil {
		return profile.Manifest{}, err
	}
	if dirErr := requireDirectory(dir); dirErr != nil {
		return profile.Manifest{}, fmt.Errorf("inspect profile directory: %w", dirErr)
	}

	data, err := os.ReadFile(filepath.Join(dir, profile.ManifestFile))
	if err != nil {
		return profile.Manifest{}, fmt.Errorf("read manifest for %q: %w", id, err)
	}
	var manifest profile.Manifest
	if err := toml.Unmarshal(data, &manifest); err != nil {
		return profile.Manifest{}, fmt.Errorf("decode manifest for %q: %w", id, err)
	}
	if err := manifest.Validate(); err != nil {
		return profile.Manifest{}, fmt.Errorf("validate manifest for %q: %w", id, err)
	}
	if manifest.DirectoryID() != id {
		return profile.Manifest{}, fmt.Errorf("manifest directory ID %q does not match directory %q", manifest.DirectoryID(), id)
	}
	return manifest, nil
}

// Create creates a profile through a temporary sibling directory and atomic rename.
func (s *Store) Create(manifest profile.Manifest) error {
	return s.CreateFromTemplate(manifest, profile.DefaultTemplate)
}

// CreateFromTemplate creates a profile from a named embedded template.
func (s *Store) CreateFromTemplate(manifest profile.Manifest, templateName string) error {
	template, files, loadErr := profile.LoadTemplate(templateName)
	if loadErr != nil {
		return fmt.Errorf("load profile template: %w", loadErr)
	}
	if manifest.Entrypoint != template.Entrypoint {
		return fmt.Errorf("template %q requires entrypoint %q", templateName, template.Entrypoint)
	}
	manifest.Template = templateName
	manifest.RequiredCommands = append([]string(nil), template.RequiredCommands...)
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("validate new manifest: %w", err)
	}
	if err := s.Ensure(); err != nil {
		return err
	}

	destination, err := s.ProfilePath(manifest.DirectoryID())
	if err != nil {
		return err
	}
	if _, statErr := os.Lstat(destination); statErr == nil {
		return fmt.Errorf("profile %q already exists", manifest.DirectoryID())
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return fmt.Errorf("inspect destination: %w", statErr)
	}

	temporary, err := os.MkdirTemp(s.Paths.ProfilesDir, ".hyprprofile-")
	if err != nil {
		return fmt.Errorf("create temporary profile directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(temporary) }()
	if err := writeManifest(temporary, manifest); err != nil {
		return err
	}
	for _, templateFile := range files {
		if err := writePrivateFile(filepath.Join(temporary, templateFile.Path), templateFile.Data); err != nil {
			return fmt.Errorf("write template file %q: %w", templateFile.Path, err)
		}
	}
	if err := os.Rename(temporary, destination); err != nil {
		return fmt.Errorf("activate new profile: %w", err)
	}
	return nil
}
