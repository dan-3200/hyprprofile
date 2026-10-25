package validation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dan-3200/hyprprofile/internal/profile"
)

// Severity distinguishes blocking validation errors from non-blocking warnings.
type Severity string

const (
	// SeverityError blocks profile activation.
	SeverityError Severity = "error"
	// SeverityWarning describes a non-blocking diagnostic.
	SeverityWarning Severity = "warning"
)

// Issue records a validation finding.
type Issue struct {
	Severity Severity
	Message  string
}

// Result contains findings from independent validation layers.
type Result struct {
	Issues []Issue
}

// Valid reports whether there are no blocking errors.
func (r Result) Valid() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			return false
		}
	}
	return true
}

// ValidateProfile checks a manifest, its entrypoint, and declared commands.
func ValidateProfile(root string, manifest profile.Manifest, checkCommands bool) Result {
	result := Result{}
	if err := manifest.Validate(); err != nil {
		return Result{Issues: []Issue{{Severity: SeverityError, Message: err.Error()}}}
	}

	entrypoint, err := ResolveEntrypoint(root, manifest.Entrypoint)
	if err != nil {
		result.Issues = append(result.Issues, Issue{Severity: SeverityError, Message: err.Error()})
	} else if info, statErr := os.Stat(entrypoint); statErr != nil || !info.Mode().IsRegular() {
		result.Issues = append(result.Issues, Issue{Severity: SeverityError, Message: "entrypoint is not a regular file"})
	}

	if checkCommands {
		for _, command := range manifest.RequiredCommands {
			if _, lookupErr := exec.LookPath(command); lookupErr != nil {
				result.Issues = append(result.Issues, Issue{
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("required command %q is unavailable", command),
				})
			}
		}
	}
	return result
}

// ResolveEntrypoint resolves an entrypoint and rejects symbolic links escaping root.
func ResolveEntrypoint(root, entrypoint string) (string, error) {
	if err := profile.ValidateRelativePath(entrypoint); err != nil {
		return "", fmt.Errorf("invalid entrypoint: %w", err)
	}
	rootInfo, err := os.Lstat(root)
	if err != nil || rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return "", fmt.Errorf("profile directory is unsafe: %s", root)
	}

	candidate := filepath.Join(root, entrypoint)
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve entrypoint: %w", err)
	}
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve profile directory: %w", err)
	}
	relative, err := filepath.Rel(rootResolved, resolved)
	if err != nil {
		return "", fmt.Errorf("compare entrypoint location: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("entrypoint escapes the profile: %s", entrypoint)
	}
	return resolved, nil
}
