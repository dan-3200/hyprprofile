// Package ports defines dependencies supplied to the application by adapters.
package ports

import "github.com/dan-3200/hyprprofile/internal/domain/profile"

// ProfileReader supplies profile data to use cases.
type ProfileReader interface {
	List() ([]profile.Manifest, error)
	Load(string) (profile.Manifest, error)
	ProfilePath(string) (string, error)
}

// ProfileWriter supplies profile mutations to use cases.
type ProfileWriter interface {
	CreateFromTemplate(profile.Manifest, string) error
	Clone(string, profile.Manifest) error
	Import(string, profile.Manifest) error
	Archive(string) (string, error)
	Restore(string) error
}

// ActiveReader reads the profile selected for the next login.
type ActiveReader interface {
	Get() (string, error)
}

// ActiveSelection also updates the selected profile.
type ActiveSelection interface {
	ActiveReader
	Set(string) error
}

// ProfileInspector checks profile files and resolves safe entrypoints.
type ProfileInspector interface {
	Validate(string, profile.Manifest) profile.Result
	ResolveEntrypoint(string, string) (string, error)
}

// TemplateCatalog lists the templates available for profile creation.
type TemplateCatalog interface {
	Templates() []profile.Template
}

// StoragePreparer prepares directories checked by doctor.
type StoragePreparer interface {
	Ensure() error
}

// SessionEntry installs and removes the display-manager entry.
type SessionEntry interface {
	Install(string, string) (string, error)
	Remove(string) (string, error)
}

// LauncherResolver finds a usable session launcher.
type LauncherResolver interface {
	Resolve(string) (string, error)
}

// ExecutableLookup finds an executable by name.
type ExecutableLookup func(string) (string, error)
