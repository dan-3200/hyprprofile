package profiles

import "github.com/dan-3200/hyprprofile/internal/domain/profile"

// Reader supplies profile data to profile use cases.
type Reader interface {
	List() ([]profile.Manifest, error)
	Load(string) (profile.Manifest, error)
	ProfilePath(string) (string, error)
}

// Writer supplies profile mutations to profile use cases.
type Writer interface {
	CreateFromTemplate(profile.Manifest, string) error
	Clone(string, profile.Manifest) error
	Import(string, profile.Manifest) error
	Archive(string) (string, error)
	Restore(string) error
}

// ActiveSelection reads and changes the selected profile.
type ActiveSelection interface {
	Get() (string, error)
	Set(string) error
}

// Inspector checks profile files.
type Inspector interface {
	Validate(string, profile.Manifest) profile.Result
}

// TemplateCatalog lists the templates available for profile creation.
type TemplateCatalog interface {
	Templates() []profile.Template
}
