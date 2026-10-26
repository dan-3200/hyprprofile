package profile

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"uuid"
)

const (
	// ManifestFile is the TOML metadata filename stored in every profile.
	ManifestFile = ".hyprprofile.toml"
	// CurrentFormat is the supported profile manifest schema version.
	CurrentFormat = 2
	// LegacyFormat identifies manifests whose ID was the directory name.
	LegacyFormat = 1
	// DefaultEntrypoint is the default relative Hyprland configuration path.
	DefaultEntrypoint = "hyprland.lua"
)

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// Manifest describes a self-contained Hyprland profile.
type Manifest struct {
	FormatVersion    int       `toml:"format_version"`
	ID               string    `toml:"id"`
	Slug             string    `toml:"slug,omitempty"`
	Name             string    `toml:"name"`
	Description      string    `toml:"description,omitempty"`
	Entrypoint       string    `toml:"entrypoint"`
	Template         string    `toml:"template,omitempty"`
	Components       []string  `toml:"components,omitempty"`
	RequiredCommands []string  `toml:"required_commands,omitempty"`
	Origin           string    `toml:"origin,omitempty"`
	CreatedAt        time.Time `toml:"created_at"`
	UpdatedAt        time.Time `toml:"updated_at"`
}

// NewManifest constructs a legacy-format manifest for existing callers.
func NewManifest(id, name, description, entrypoint string, now time.Time) Manifest {
	if entrypoint == "" {
		entrypoint = DefaultEntrypoint
	}

	stamp := now.UTC()
	return Manifest{
		FormatVersion: LegacyFormat,
		ID:            id,
		Name:          name,
		Description:   description,
		Entrypoint:    entrypoint,
		CreatedAt:     stamp,
		UpdatedAt:     stamp,
	}
}

// NewUUIDManifest constructs a new manifest with a UUID identity and directory slug.
func NewUUIDManifest(slug, name, description, entrypoint string, now time.Time) Manifest {
	manifest := NewManifest(uuid.NewV4().String(), name, description, entrypoint, now)
	manifest.FormatVersion = CurrentFormat
	manifest.Slug = slug
	return manifest
}

// NewUUID generates a canonical random version 4 UUID.
func NewUUID() string {
	return uuid.NewV4().String()
}

// DirectoryID is the stable name used by commands, paths and active selection.
func (m Manifest) DirectoryID() string {
	if m.FormatVersion == LegacyFormat {
		return m.ID
	}
	return m.Slug
}

// ValidateID validates immutable profile IDs used as directory names.
func ValidateID(id string) error {
	if !idPattern.MatchString(id) {
		return fmt.Errorf("profile ID %q must match [a-z0-9][a-z0-9-]{0,63}", id)
	}
	return nil
}

// Validate checks the manifest fields that do not require filesystem access.
func (m Manifest) Validate() error {
	if m.FormatVersion != CurrentFormat && m.FormatVersion != LegacyFormat {
		return fmt.Errorf("unsupported manifest format version %d", m.FormatVersion)
	}
	if m.FormatVersion == LegacyFormat {
		if m.Slug != "" {
			return errors.New("legacy manifest cannot define a slug")
		}
		if err := ValidateID(m.ID); err != nil {
			return err
		}
	} else {
		parsed, err := uuid.Parse(m.ID)
		if err != nil || parsed.String() != m.ID || parsed[6]>>4 == 0 || parsed[6]>>4 > 8 || parsed[8]&0xc0 != 0x80 {
			return fmt.Errorf("profile ID %q must be a canonical UUID", m.ID)
		}
		if err := ValidateID(m.Slug); err != nil {
			return fmt.Errorf("invalid profile slug: %w", err)
		}
	}
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("profile name cannot be empty")
	}
	if err := ValidateRelativePath(m.Entrypoint); err != nil {
		return fmt.Errorf("invalid entrypoint: %w", err)
	}
	if m.Template != "" && !idPattern.MatchString(m.Template) {
		return fmt.Errorf("invalid template name %q", m.Template)
	}
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() {
		return errors.New("manifest timestamps cannot be empty")
	}
	return nil
}

// ValidateRelativePath rejects absolute paths and parent-directory traversal.
func ValidateRelativePath(candidate string) error {
	if candidate == "" {
		return errors.New("path cannot be empty")
	}
	if filepath.IsAbs(candidate) || filepath.VolumeName(candidate) != "" {
		return fmt.Errorf("path %q must be relative", candidate)
	}

	clean := filepath.Clean(candidate)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path %q escapes the profile", candidate)
	}
	return nil
}
