package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/dan-3200/hyprprofile/internal/profile"
	"github.com/dan-3200/hyprprofile/internal/storage"
)

// Manager manages the active profile selector used by the session launcher.
type Manager struct {
	paths storage.Paths
}

// NewManager creates an active-profile manager.
func NewManager(paths storage.Paths) Manager {
	return Manager{paths: paths}
}

// Get returns the selected profile ID. A missing selector means no active profile.
func (m Manager) Get() (string, error) {
	data, err := os.ReadFile(m.paths.ActiveProfileFile)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read active profile: %w", err)
	}
	id := strings.TrimSpace(string(data))
	if err := profile.ValidateID(id); err != nil {
		return "", fmt.Errorf("active profile state is invalid: %w", err)
	}
	return id, nil
}

// Activate switches the session selector only after the profile is present.
func (m Manager) Activate(store *storage.Store, id string) error {
	if _, err := store.Load(id); err != nil {
		return fmt.Errorf("load profile to activate: %w", err)
	}
	if err := storage.WriteAtomicFile(m.paths.ActiveProfileFile, []byte(id+"\n")); err != nil {
		return fmt.Errorf("write active profile: %w", err)
	}
	return nil
}
