package profile

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// ManifestFile is the TOML metadata filename stored in every profile.
	ManifestFile = ".hyprprofile.toml"
	// CurrentFormat is the supported profile manifest schema version.
	CurrentFormat = 1
	// DefaultEntrypoint is the default relative Hyprland configuration path.
	DefaultEntrypoint = "hyprland.lua"
)

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// Manifest describes a self-contained Hyprland profile.
type Manifest struct {
	FormatVersion    int       `toml:"format_version"`
	ID               string    `toml:"id"`
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

// NewManifest returns a valid manifest with timestamps in UTC.
func NewManifest(id, name, description, entrypoint string, now time.Time) Manifest {
	if entrypoint == "" {
		entrypoint = DefaultEntrypoint
	}

	stamp := now.UTC()
	return Manifest{
		FormatVersion: CurrentFormat,
		ID:            id,
		Name:          name,
		Description:   description,
		Entrypoint:    entrypoint,
		CreatedAt:     stamp,
		UpdatedAt:     stamp,
	}
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
	if m.FormatVersion != CurrentFormat {
		return fmt.Errorf("unsupported manifest format version %d", m.FormatVersion)
	}
	if err := ValidateID(m.ID); err != nil {
		return err
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
