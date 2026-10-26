package storage

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/dan-3200/hyprprofile/internal/domain/profile"
)

// Clone copies a profile, updates immutable metadata, and publishes it atomically.
func (s *Store) Clone(sourceID string, target profile.Manifest) error {
	if err := target.Validate(); err != nil {
		return fmt.Errorf("validate clone manifest: %w", err)
	}
	source, err := s.ProfilePath(sourceID)
	if err != nil {
		return fmt.Errorf("resolve clone source: %w", err)
	}
	if dirErr := requireDirectory(source); dirErr != nil {
		return fmt.Errorf("inspect clone source: %w", dirErr)
	}
	if ensureErr := s.Ensure(); ensureErr != nil {
		return fmt.Errorf("prepare profile storage: %w", ensureErr)
	}

	destination, err := s.ProfilePath(target.DirectoryID())
	if err != nil {
		return err
	}
	if _, statErr := os.Lstat(destination); statErr == nil {
		return fmt.Errorf("profile %q already exists", target.DirectoryID())
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return fmt.Errorf("inspect clone destination: %w", statErr)
	}

	temporary, err := os.MkdirTemp(s.Paths.ProfilesDir, ".hyprprofile-")
	if err != nil {
		return fmt.Errorf("create temporary clone directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(temporary) }()
	if err := copyTree(source, temporary); err != nil {
		return fmt.Errorf("copy profile %q: %w", sourceID, err)
	}
	if err := writeManifest(temporary, target); err != nil {
		return fmt.Errorf("write cloned manifest: %w", err)
	}
	if err := os.Rename(temporary, destination); err != nil {
		return fmt.Errorf("publish cloned profile: %w", err)
	}
	return nil
}

// Import copies an existing profile directory and adds its manifest.
func (s *Store) Import(source string, manifest profile.Manifest) error {
	if !filepath.IsAbs(source) {
		return errors.New("import source must be an absolute directory")
	}
	if dirErr := requireDirectory(source); dirErr != nil {
		return fmt.Errorf("inspect import source: %w", dirErr)
	}
	if manifest.Entrypoint == "" {
		entrypoint, err := detectEntrypoint(source)
		if err != nil {
			return fmt.Errorf("detect import entrypoint: %w", err)
		}
		manifest.Entrypoint = entrypoint
	}
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("validate import manifest: %w", err)
	}
	return s.CloneFromDirectory(source, manifest)
}

// CloneFromDirectory is the import primitive exposed for validated migration sources.
func (s *Store) CloneFromDirectory(source string, manifest profile.Manifest) error {
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("validate import manifest: %w", err)
	}
	if ensureErr := s.Ensure(); ensureErr != nil {
		return fmt.Errorf("prepare profile storage: %w", ensureErr)
	}

	destination, err := s.ProfilePath(manifest.DirectoryID())
	if err != nil {
		return err
	}
	if _, statErr := os.Lstat(destination); statErr == nil {
		return fmt.Errorf("profile %q already exists", manifest.DirectoryID())
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return fmt.Errorf("inspect import destination: %w", statErr)
	}

	temporary, err := os.MkdirTemp(s.Paths.ProfilesDir, ".hyprprofile-")
	if err != nil {
		return fmt.Errorf("create temporary import directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(temporary) }()
	if err := copyTree(source, temporary); err != nil {
		return fmt.Errorf("copy import source: %w", err)
	}
	if err := writeManifest(temporary, manifest); err != nil {
		return fmt.Errorf("write imported manifest: %w", err)
	}
	if err := os.Rename(temporary, destination); err != nil {
		return fmt.Errorf("publish imported profile: %w", err)
	}
	return nil
}

// Archive moves a profile to the recoverable trash directory.
func (s *Store) Archive(id string) (string, error) {
	source, err := s.ProfilePath(id)
	if err != nil {
		return "", fmt.Errorf("resolve profile path: %w", err)
	}
	if dirErr := requireDirectory(source); dirErr != nil {
		return "", fmt.Errorf("inspect profile directory: %w", dirErr)
	}
	if ensureErr := s.Ensure(); ensureErr != nil {
		return "", fmt.Errorf("prepare profile storage: %w", ensureErr)
	}

	stamp := s.now().UTC().Format("20060102T150405.000000000Z")
	destination := filepath.Join(s.Paths.TrashDir, stamp+"-"+id)
	if err := moveDirectory(source, destination); err != nil {
		return "", fmt.Errorf("archive profile %q: %w", id, err)
	}
	return destination, nil
}

// Restore returns the latest archived profile with the requested ID.
func (s *Store) Restore(id string) error {
	if err := profile.ValidateID(id); err != nil {
		return fmt.Errorf("validate profile ID: %w", err)
	}
	if ensureErr := s.Ensure(); ensureErr != nil {
		return fmt.Errorf("prepare profile storage: %w", ensureErr)
	}

	destination, err := s.ProfilePath(id)
	if err != nil {
		return err
	}
	if _, statErr := os.Lstat(destination); statErr == nil {
		return fmt.Errorf("profile %q already exists", id)
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return fmt.Errorf("inspect restore destination: %w", statErr)
	}

	entries, err := os.ReadDir(s.Paths.TrashDir)
	if err != nil {
		return fmt.Errorf("read trash directory: %w", err)
	}
	prefix := "-" + id
	candidates := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() && entry.Type()&os.ModeSymlink == 0 && strings.HasSuffix(entry.Name(), prefix) {
			candidates = append(candidates, entry.Name())
		}
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no archived profile found for %q", id)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(candidates)))
	source := filepath.Join(s.Paths.TrashDir, candidates[0])
	if err := moveDirectory(source, destination); err != nil {
		return fmt.Errorf("restore profile %q: %w", id, err)
	}
	return nil
}

func detectEntrypoint(root string) (string, error) {
	for _, candidate := range []string{"hyprland.lua", "hyprland.conf"} {
		info, err := os.Lstat(filepath.Join(root, candidate))
		if err == nil && info.Mode().IsRegular() {
			return candidate, nil
		}
	}
	return "", errors.New("could not find hyprland.lua or hyprland.conf in import source")
}

func moveDirectory(source, destination string) error {
	if err := os.Rename(source, destination); err == nil {
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return fmt.Errorf("rename directory: %w", err)
	}
	if err := copyTree(source, destination); err != nil {
		return fmt.Errorf("copy across filesystems: %w", err)
	}
	if err := os.RemoveAll(source); err != nil {
		return fmt.Errorf("remove copied source: %w", err)
	}
	return nil
}

func copyTree(source, destination string) error {
	if err := requireDirectory(source); err != nil {
		return err
	}
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk source: %w", walkErr)
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("derive relative path: %w", err)
		}
		if relative == "." {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symbolic link %q", relative)
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return fmt.Errorf("create directory %q: %w", relative, err)
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("refusing non-regular file %q", relative)
		}
		return copyPrivateFile(path, target)
	})
	if err != nil {
		return fmt.Errorf("walk profile tree: %w", err)
	}
	return nil
}

func copyPrivateFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer func() { _ = input.Close() }()

	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer func() { _ = output.Close() }()
	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("copy file contents: %w", err)
	}
	return nil
}
