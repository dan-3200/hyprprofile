package system

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
