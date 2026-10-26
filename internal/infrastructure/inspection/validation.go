package inspection

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dan-3200/hyprprofile/internal/domain/profile"
)

// Result is the domain validation result used by this adapter.
type Result = profile.Result

// Issue is one domain validation finding.
type Issue = profile.Issue

// SeverityError blocks activation.
const SeverityError = profile.SeverityError

// SeverityWarning is a non-blocking diagnostic.
const SeverityWarning = profile.SeverityWarning

// Inspector checks profile files and commands in the local environment.
type Inspector struct {
	Lookup func(string) (string, error)
}

// DefaultInspector uses the current process PATH.
func DefaultInspector() Inspector {
	return Inspector{Lookup: exec.LookPath}
}

// Validate checks a profile for application and launcher use.
func (i Inspector) Validate(root string, manifest profile.Manifest) profile.Result {
	return validateProfile(root, manifest, true, i.Lookup)
}

// ResolveEntrypoint resolves the validated entrypoint for session launch.
func (i Inspector) ResolveEntrypoint(root, entrypoint string) (string, error) {
	return ResolveEntrypoint(root, entrypoint)
}

// ValidateProfile checks a manifest, its entrypoint, and declared commands.
func ValidateProfile(root string, manifest profile.Manifest, checkCommands bool) Result {
	return validateProfile(root, manifest, checkCommands, exec.LookPath)
}

func validateProfile(root string, manifest profile.Manifest, checkCommands bool, lookup func(string) (string, error)) Result {
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
			if _, lookupErr := lookup(command); lookupErr != nil {
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
