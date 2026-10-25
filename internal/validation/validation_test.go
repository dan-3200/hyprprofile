package validation

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/profile"
)

func TestValidateProfileRejectsEntrypointSymlinkOutsideProfile(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.lua")
	if err := os.WriteFile(outside, []byte("-- outside\n"), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, profile.DefaultEntrypoint)); err != nil {
		t.Fatalf("create symbolic link: %v", err)
	}
	manifest := profile.NewManifest("minimal", "Minimal", "", profile.DefaultEntrypoint, time.Now())
	result := ValidateProfile(root, manifest, false)
	if result.Valid() {
		t.Fatal("ValidateProfile() unexpectedly accepted an escaping symbolic link")
	}
}
