package profile

import (
	"testing"
	"time"
)

func TestValidateID(t *testing.T) {
	t.Parallel()
	valid := []string{"minimal", "serpantinum", "a", "profile-42"}
	for _, id := range valid {
		if err := ValidateID(id); err != nil {
			t.Errorf("ValidateID(%q) returned %v", id, err)
		}
	}
}

func TestValidateIDRejectsUnsafeValues(t *testing.T) {
	t.Parallel()
	invalid := []string{"", "Upper", "-prefix", "has_space", "../escape", "a/child"}
	for _, id := range invalid {
		if err := ValidateID(id); err == nil {
			t.Errorf("ValidateID(%q) unexpectedly succeeded", id)
		}
	}
}

func TestValidateRelativePath(t *testing.T) {
	t.Parallel()
	valid := []string{"hyprland.lua", "themes/dark.conf", "scripts/start"}
	for _, candidate := range valid {
		if err := ValidateRelativePath(candidate); err != nil {
			t.Errorf("ValidateRelativePath(%q) returned %v", candidate, err)
		}
	}
	invalid := []string{"", "/etc/passwd", "../outside", "a/../../outside", "."}
	for _, candidate := range invalid {
		if err := ValidateRelativePath(candidate); err == nil {
			t.Errorf("ValidateRelativePath(%q) unexpectedly succeeded", candidate)
		}
	}
}

func TestManifestRejectsUnsafeTemplateName(t *testing.T) {
	t.Parallel()
	manifest := NewManifest("minimal", "Minimal", "", DefaultEntrypoint, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	manifest.Template = "../outside"
	if err := manifest.Validate(); err == nil {
		t.Fatal("Validate() unexpectedly accepted an unsafe template name")
	}
}
