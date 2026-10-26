//go:build !gui

package gui

import "errors"

// Run reports how to enable the native GTK and libadwaita implementation.
func Run(ProfileController) error {
	return errors.New("the GUI was not included; rebuild with -tags gui after installing gtk4-devel and libadwaita-devel")
}
