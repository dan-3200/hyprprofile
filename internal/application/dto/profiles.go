// Package dto defines application inputs and results shared with user interfaces.
package dto

import "github.com/dan-3200/hyprprofile/internal/domain/profile"

// ProfileListItem includes selected state alongside profile metadata.
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

// CommandStatus describes whether a diagnostic executable is available.
type CommandStatus struct {
	Name      string
	Available bool
}

// LaunchPlan is the validated process invocation for a Hyprland session.
type LaunchPlan struct {
	Executable string
	Arguments  []string
}
