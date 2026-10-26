package profile

import (
	"testing"
	"time"
	"uuid"
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

func TestUUIDManifestIdentityAndLegacyCompatibility(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	modern := NewUUIDManifest("work", "Work", "", DefaultEntrypoint, now)
	parsed, err := uuid.Parse(modern.ID)
	if err != nil || modern.Validate() != nil || modern.DirectoryID() != "work" || parsed.String() != modern.ID || parsed[6]>>4 != 4 {
		t.Fatalf("UUID manifest = %+v, %v", modern, err)
	}
	legacy := NewManifest("work", "Work", "", DefaultEntrypoint, now)
	if err := legacy.Validate(); err != nil || legacy.DirectoryID() != "work" {
		t.Fatalf("legacy manifest = %+v, %v", legacy, err)
	}
	modern.ID = "work"
	if err := modern.Validate(); err == nil {
		t.Fatal("new manifest accepted a non-UUID ID")
	}
	legacy.Slug = "different"
	if err := legacy.Validate(); err == nil {
		t.Fatal("legacy manifest accepted a second directory identity")
	}
}
