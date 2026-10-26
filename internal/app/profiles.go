// Package app coordinates profile and system operations for user interfaces.
package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/dan-3200/hyprprofile/internal/profile"
	"github.com/dan-3200/hyprprofile/internal/validation"
)

// ProfileReader supplies the profile data used by application operations.
type ProfileReader interface {
	List() ([]profile.Manifest, error)
	Load(string) (profile.Manifest, error)
	ProfilePath(string) (string, error)
}

// ProfileWriter supplies the storage mutations used by application operations.
type ProfileWriter interface {
	CreateFromTemplate(profile.Manifest, string) error
	Clone(string, profile.Manifest) error
	Import(string, profile.Manifest) error
	Archive(string) (string, error)
	Restore(string) error
}

// ActiveSelection persists the profile selected for the next login.
type ActiveSelection interface {
	Get() (string, error)
	Set(string) error
}

// ProfileService coordinates profile operations for any local UI.
type ProfileService struct {
	reader ProfileReader
	writer ProfileWriter
	active ActiveSelection
	now    func() time.Time
}

// NewProfileService wires profile storage, selection and a clock.
func NewProfileService(reader ProfileReader, writer ProfileWriter, active ActiveSelection, now func() time.Time) *ProfileService {
	return &ProfileService{reader: reader, writer: writer, active: active, now: now}
}

// ProfileListItem includes the selected state alongside profile metadata.
type ProfileListItem struct {
	Manifest profile.Manifest
	Active   bool
}

// CreateProfileInput describes a new embedded-template profile.
type CreateProfileInput struct {
	ID, Name, Description, Template string
}

// CloneProfileInput describes a copy of an existing profile.
type CloneProfileInput struct {
	SourceID, ID, Name string
}

// ImportProfileInput describes a profile copied from a local directory.
type ImportProfileInput struct {
	Source, ID, Name, Description, Entrypoint string
}

// ArchiveResult identifies the recoverable location of a removed profile.
type ArchiveResult struct {
	ID, Destination string
}

// ListProfiles returns profiles in storage order with the selected profile marked.
func (s *ProfileService) ListProfiles() ([]ProfileListItem, error) {
	items, err := s.reader.List()
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	active, err := s.active.Get()
	if err != nil {
		return nil, fmt.Errorf("read active profile: %w", err)
	}
	result := make([]ProfileListItem, len(items))
	for i, item := range items {
		result[i] = ProfileListItem{Manifest: item, Active: item.DirectoryID() == active}
	}
	return result, nil
}

// Status reports whether a profile has been selected.
func (s *ProfileService) Status() (profile.Manifest, bool, error) {
	id, err := s.active.Get()
	if err != nil {
		return profile.Manifest{}, false, fmt.Errorf("read active profile: %w", err)
	}
	if id == "" {
		return profile.Manifest{}, false, nil
	}
	item, err := s.reader.Load(id)
	if err != nil {
		return profile.Manifest{}, false, fmt.Errorf("active profile %q is unavailable: %w", id, err)
	}
	return item, true, nil
}

// Templates returns the embedded template catalog.
func (s *ProfileService) Templates() []profile.Template {
	return profile.Templates()
}

// CreateProfile creates a profile with application-level defaults.
func (s *ProfileService) CreateProfile(input CreateProfileInput) (profile.Manifest, error) {
	if input.Template == "" {
		input.Template = profile.DefaultTemplate
	}
	item := profile.NewUUIDManifest(input.ID, input.Name, input.Description, profile.DefaultEntrypoint, s.now())
	item.Template = input.Template
	if err := s.writer.CreateFromTemplate(item, input.Template); err != nil {
		return profile.Manifest{}, fmt.Errorf("create profile: %w", err)
	}
	return item, nil
}

// CloneProfile duplicates a stored profile and updates its identity.
func (s *ProfileService) CloneProfile(input CloneProfileInput) (profile.Manifest, error) {
	source, err := s.reader.Load(input.SourceID)
	if err != nil {
		return profile.Manifest{}, fmt.Errorf("load source profile: %w", err)
	}
	if input.Name == "" {
		input.Name = source.Name + " (clone)"
	}
	target := source
	target.ID = profile.NewUUID()
	target.FormatVersion = profile.CurrentFormat
	target.Slug, target.Name = input.ID, input.Name
	target.CreatedAt = s.now().UTC()
	target.UpdatedAt = target.CreatedAt
	target.Origin = source.DirectoryID()
	if err := s.writer.Clone(source.DirectoryID(), target); err != nil {
		return profile.Manifest{}, fmt.Errorf("clone profile: %w", err)
	}
	return target, nil
}

// ImportProfile copies a local configuration with entrypoint detection when omitted.
func (s *ProfileService) ImportProfile(input ImportProfileInput) (profile.Manifest, error) {
	if input.Name == "" {
		input.Name = input.ID
	}
	item := profile.NewUUIDManifest(input.ID, input.Name, input.Description, input.Entrypoint, s.now())
	if input.Entrypoint == "" {
		item.Entrypoint = ""
	}
	item.Origin = input.Source
	if importErr := s.writer.Import(input.Source, item); importErr != nil {
		return profile.Manifest{}, fmt.Errorf("import profile: %w", importErr)
	}
	imported, err := s.reader.Load(item.DirectoryID())
	if err != nil {
		return profile.Manifest{}, fmt.Errorf("load imported profile: %w", err)
	}
	return imported, nil
}

// ValidateProfile checks a stored profile without executing its configuration.
func (s *ProfileService) ValidateProfile(id string) (validation.Result, error) {
	item, err := s.reader.Load(id)
	if err != nil {
		return validation.Result{}, fmt.Errorf("load profile: %w", err)
	}
	directory, err := s.reader.ProfilePath(item.DirectoryID())
	if err != nil {
		return validation.Result{}, fmt.Errorf("resolve profile path: %w", err)
	}
	return validation.ValidateProfile(directory, item, true), nil
}

// ActivateProfile selects a structurally valid profile for the next login.
func (s *ProfileService) ActivateProfile(id string) (profile.Manifest, error) {
	item, err := s.reader.Load(id)
	if err != nil {
		return profile.Manifest{}, fmt.Errorf("load profile: %w", err)
	}
	directory, err := s.reader.ProfilePath(item.DirectoryID())
	if err != nil {
		return profile.Manifest{}, fmt.Errorf("resolve profile path: %w", err)
	}
	if !validation.ValidateProfile(directory, item, true).Valid() {
		return profile.Manifest{}, errors.New("cannot activate an invalid profile; run validate for details")
	}
	if err := s.active.Set(item.DirectoryID()); err != nil {
		return profile.Manifest{}, fmt.Errorf("activate profile: %w", err)
	}
	return item, nil
}

// ArchiveProfile moves an inactive profile to recoverable storage.
func (s *ProfileService) ArchiveProfile(id string) (ArchiveResult, error) {
	active, err := s.active.Get()
	if err != nil {
		return ArchiveResult{}, fmt.Errorf("read active profile: %w", err)
	}
	if id == active {
		return ArchiveResult{}, errors.New("cannot remove the active profile; activate another profile first")
	}
	destination, err := s.writer.Archive(id)
	if err != nil {
		return ArchiveResult{}, fmt.Errorf("archive profile: %w", err)
	}
	return ArchiveResult{ID: id, Destination: destination}, nil
}

// RestoreProfile restores the latest archived profile with the ID.
func (s *ProfileService) RestoreProfile(id string) error {
	if err := s.writer.Restore(id); err != nil {
		return fmt.Errorf("restore profile: %w", err)
	}
	return nil
}
