package launch

import "github.com/dan-3200/hyprprofile/internal/domain/profile"

// ProfileReader supplies the selected profile for launch preparation.
type ProfileReader interface {
	Load(string) (profile.Manifest, error)
	ProfilePath(string) (string, error)
}

// ActiveReader reads the profile selected for the next login.
type ActiveReader interface {
	Get() (string, error)
}

// Inspector checks a profile and resolves its entrypoint.
type Inspector interface {
	Validate(string, profile.Manifest) profile.Result
	ResolveEntrypoint(string, string) (string, error)
}

// ExecutableLookup finds a compositor launcher by name.
type ExecutableLookup func(string) (string, error)
