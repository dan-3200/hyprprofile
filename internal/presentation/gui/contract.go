// Package gui provides the native GNOME interface for HyprProfile.
package gui

import (
	"github.com/dan-3200/hyprprofile/internal/application/profiles"
	"github.com/dan-3200/hyprprofile/internal/domain/profile"
)

// ProfileController provides the profile actions used by the GUI.
//
//nolint:iface // Its consumer is included only by the gui build tag.
type ProfileController interface {
	ListProfiles() ([]profiles.ProfileListItem, error)
	ActivateProfile(string) (profile.Manifest, error)
}
