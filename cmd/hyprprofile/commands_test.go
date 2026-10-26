package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/application/profiles"
	"github.com/dan-3200/hyprprofile/internal/application/system"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/inspection"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/session"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/storage"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/templates"
	"github.com/dan-3200/hyprprofile/internal/presentation/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIProfileFlowAndOutput(t *testing.T) {
	root := t.TempDir()
	paths, err := storage.NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	store := storage.New(paths)
	profiles := profiles.NewProfileService(store, store, session.NewManager(paths), inspection.DefaultInspector(), templates.Catalog{}, time.Now)
	system := system.NewService(store, exec.LookPath, session.DesktopEntry{}, session.DefaultLauncherResolver())
	cases := []struct {
		arguments []string
		want      string
	}{
		{[]string{"status"}, "No active profile\n"},
		{[]string{"create", "work", "--name", "Work"}, "Created profile \"work\" from template \"minimal\"\n"},
		{[]string{"validate", "work"}, "Profile \"work\" is valid\n"},
		{[]string{"activate", "work"}, "Activated profile \"work\" for the next login\n"},
		{[]string{"status"}, "Active profile: Work (work)\n"},
		{[]string{"list"}, "* work                 Work\n"},
		{[]string{"clone", "work", "copy"}, "Cloned \"work\" as \"copy\"\n"},
		{[]string{"remove", "copy"}, "Archived profile \"copy\" to "},
		{[]string{"restore", "copy"}, "Restored profile \"copy\"\n"},
	}
	for _, tc := range cases {
		output, runErr := capturedOutput(func() error { return cli.Run(tc.arguments, profiles, system, session.DefaultSessionDirectory) })
		require.NoError(t, runErr, "run(%q)", tc.arguments)
		assert.True(t, strings.HasPrefix(output, tc.want), "run(%q) = %q; want prefix %q", tc.arguments, output, tc.want)
	}
	_, err = runError([]string{"remove", "work"}, profiles, system)
	require.EqualError(t, err, "cannot remove the active profile; activate another profile first")
}

func TestCLIExitCodeOnMissingProfile(t *testing.T) {
	if os.Getenv("HYPRPROFILE_TEST_CHILD") == "1" {
		os.Args = []string{"hyprprofile", "activate", "missing"}
		main()
		return
	}
	root := t.TempDir()
	// #nosec G702 -- this test reruns its own executable with fixed test arguments.
	command := exec.Command(os.Args[0], "-test.run=^TestCLIExitCodeOnMissingProfile$")
	command.Env = append(os.Environ(),
		"HYPRPROFILE_TEST_CHILD=1",
		"XDG_CONFIG_HOME="+filepath.Join(root, "config"),
		"XDG_STATE_HOME="+filepath.Join(root, "state"),
	)
	output, err := command.CombinedOutput()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
	assert.Contains(t, string(output), "hyprprofile: load profile:")
}

func runError(arguments []string, profiles *profiles.ProfileService, system *system.Service) (string, error) {
	return capturedOutput(func() error { return cli.Run(arguments, profiles, system, session.DefaultSessionDirectory) })
}

func capturedOutput(action func() error) (string, error) {
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("create output pipe: %w", err)
	}
	os.Stdout = writer
	defer func() { os.Stdout = original }()
	runErr := action()
	_ = writer.Close()
	output, readErr := io.ReadAll(reader)
	_ = reader.Close()
	if readErr != nil {
		return "", fmt.Errorf("read output: %w", readErr)
	}
	return string(output), runErr
}
